// Package ozon — журнал отчётов Ozon Seller API: отчёт за каждый день
// продаж скачивается по расписанию и раскладывается по промокодам интеграций.
// Сам разбор делает воркер; API показывает журнал и ставит загрузку в очередь.
package ozon

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Report struct {
	Day         string     `json:"day"`
	ReportID    string     `json:"reportId"`
	Status      string     `json:"status"`
	Trigger     string     `json:"trigger"`
	Attempts    int        `json:"attempts"`
	RowsTotal   int        `json:"rowsTotal"`
	Orders      int        `json:"orders"`
	RevenueRub  int64      `json:"revenueRub"`
	Error       *string    `json:"error"`
	RequestedAt time.Time  `json:"requestedAt"`
	FinishedAt  *time.Time `json:"finishedAt"`
}

// Line — продажи одной интеграции в отчёте за день.
type Line struct {
	IntegrationID int64  `json:"integrationId"`
	PromoCode     string `json:"promoCode"`
	BloggerID     int64  `json:"bloggerId"`
	Blogger       string `json:"blogger"`
	CampaignID    int64  `json:"campaignId"`
	Campaign      string `json:"campaign"`
	Orders        int64  `json:"orders"`
	RevenueRub    int64  `json:"revenueRub"`
}

// ErrBusy — отчёт за этот день уже в очереди или загружается.
var ErrBusy = errors.New("report is already queued or processing")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const reportColumns = `day::text, report_id, status, trigger, attempts, rows_total, orders, revenue_rub, error, requested_at, finished_at`

func scanReport(row pgx.Row) (Report, error) {
	var r Report
	err := row.Scan(&r.Day, &r.ReportID, &r.Status, &r.Trigger, &r.Attempts, &r.RowsTotal, &r.Orders,
		&r.RevenueRub, &r.Error, &r.RequestedAt, &r.FinishedAt)
	return r, err
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Report, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ozon_reports`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+reportColumns+` FROM ozon_reports ORDER BY day DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Report, error) { return scanReport(row) })
	return items, total, err
}

func (r *Repository) Lines(ctx context.Context, day time.Time) ([]Line, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.id, i.promo_code, b.id, b.name, c.id, c.title, m.orders, m.revenue_rub
		FROM integration_metrics m
		JOIN integrations i ON i.id = m.integration_id
		JOIN bloggers b ON b.id = i.blogger_id
		JOIN campaigns c ON c.id = i.campaign_id
		WHERE m.day = $1 AND m.source = 'ozon' AND m.orders > 0
		ORDER BY m.revenue_rub DESC, i.id`, day)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Line, error) {
		var l Line
		err := row.Scan(&l.IntegrationID, &l.PromoCode, &l.BloggerID, &l.Blogger, &l.CampaignID, &l.Campaign, &l.Orders, &l.RevenueRub)
		return l, err
	})
}

// Enqueue ставит загрузку отчёта за day в очередь. Если отчёт уже ждёт
// или скачивается, второй раз его не ставим — возвращаем ErrBusy.
func Enqueue(ctx context.Context, pool *pgxpool.Pool, day time.Time, trigger string) error {
	tag, err := pool.Exec(ctx, `
		INSERT INTO ozon_reports (day, report_id, status, trigger, requested_at)
		VALUES ($1, $2, 'queued', $3, now())
		ON CONFLICT (day) DO UPDATE
		SET status = 'queued', trigger = EXCLUDED.trigger, report_id = EXCLUDED.report_id,
		    requested_at = now(), finished_at = NULL
		WHERE ozon_reports.status NOT IN ('queued', 'processing')`,
		day, fmt.Sprintf("rpt-%06d", rand.IntN(1_000_000)), trigger)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrBusy
	}
	return nil
}
