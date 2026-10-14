package campaign

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"crmdemo/internal/httpx"
	"crmdemo/internal/metrics"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const listSQL = `
WITH cost AS (
    SELECT campaign_id, COUNT(*) AS n, SUM(cost_rub)::bigint AS cost
    FROM integrations GROUP BY campaign_id
), sales AS (
    SELECT i.campaign_id, SUM(m.revenue_rub)::bigint AS revenue, SUM(m.orders)::bigint AS orders
    FROM integration_metrics m JOIN integrations i ON i.id = m.integration_id
    GROUP BY i.campaign_id
)
SELECT c.id, c.title, c.product, c.status, c.start_date::text, c.end_date::text, c.budget_rub,
       COALESCE(k.n, 0), COALESCE(k.cost, 0), COALESCE(s.revenue, 0), COALESCE(s.orders, 0)
FROM campaigns c
LEFT JOIN cost k ON k.campaign_id = c.id
LEFT JOIN sales s ON s.campaign_id = c.id
WHERE %s
ORDER BY c.start_date DESC, c.id DESC`

func scanListItem(row pgx.Row) (ListItem, error) {
	var (
		it                    ListItem
		cost, revenue, orders int64
	)
	err := row.Scan(&it.ID, &it.Title, &it.Product, &it.Status, &it.StartDate, &it.EndDate, &it.BudgetRub,
		&it.Integrations, &cost, &revenue, &orders)
	it.Totals = metrics.NewTotals(cost, revenue, orders)
	return it, err
}

func (r *Repository) List(ctx context.Context, status string) ([]ListItem, error) {
	where, args := "TRUE", []any{}
	if status != "" {
		where, args = "c.status = $1", append(args, status)
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(listSQL, where), args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (ListItem, error) { return scanListItem(row) })
}

func (r *Repository) Get(ctx context.Context, id int64) (Details, error) {
	var d Details
	it, err := scanListItem(r.pool.QueryRow(ctx, fmt.Sprintf(listSQL, "c.id = $1"), id))
	if errors.Is(err, pgx.ErrNoRows) {
		return d, httpx.ErrNotFound
	}
	if err != nil {
		return d, err
	}
	d.ListItem = it

	rows, err := r.pool.Query(ctx, `
		SELECT i.id, b.id, b.name, b.handle, p.code, i.publish_date::text, i.promo_code, i.utm_campaign, i.status,
		       i.cost_rub, COALESCE(SUM(m.revenue_rub), 0)::bigint, COALESCE(SUM(m.orders), 0)::bigint
		FROM integrations i
		JOIN bloggers b ON b.id = i.blogger_id
		JOIN platforms p ON p.id = b.platform_id
		LEFT JOIN integration_metrics m ON m.integration_id = i.id
		WHERE i.campaign_id = $1
		GROUP BY i.id, b.id, p.code
		ORDER BY i.publish_date, i.id`, id)
	if err != nil {
		return d, err
	}
	d.Integrations, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (Integration, error) {
		var (
			in                    Integration
			cost, revenue, orders int64
		)
		err := row.Scan(&in.ID, &in.BloggerID, &in.Blogger, &in.Handle, &in.Platform, &in.PublishDate,
			&in.PromoCode, &in.UTM, &in.Status, &cost, &revenue, &orders)
		in.Totals = metrics.NewTotals(cost, revenue, orders)
		return in, err
	})
	if err != nil {
		return d, err
	}

	rows, err = r.pool.Query(ctx, `
		SELECT m.day::text, SUM(m.orders)::bigint, SUM(m.revenue_rub)::bigint
		FROM integration_metrics m JOIN integrations i ON i.id = m.integration_id
		WHERE i.campaign_id = $1
		GROUP BY m.day ORDER BY m.day`, id)
	if err != nil {
		return d, err
	}
	d.Daily, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (DayPoint, error) {
		var p DayPoint
		err := row.Scan(&p.Day, &p.Orders, &p.RevenueRub)
		return p, err
	})
	return d, err
}

func (r *Repository) Create(ctx context.Context, in CreateInput, start, end time.Time, status string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO campaigns (title, product, status, start_date, end_date, budget_rub)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		in.Title, in.Product, status, start, end, in.BudgetRub).Scan(&id)
	return id, err
}

// AddIntegration создаёт интеграцию с уникальным промокодом;
// при редкой коллизии кода пробует сгенерировать новый.
func (r *Repository) AddIntegration(ctx context.Context, campaignID int64, in AddIntegrationInput, publish time.Time, status string) (int64, string, error) {
	var handle string
	err := r.pool.QueryRow(ctx, `SELECT handle FROM bloggers WHERE id = $1`, in.BloggerID).Scan(&handle)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", httpx.ValidationError{Msg: "bloggerId: блогер не найден"}
	}
	if err != nil {
		return 0, "", err
	}

	for attempt := 0; attempt < 5; attempt++ {
		promo := NewPromoCode(handle)
		var id int64
		err = r.pool.QueryRow(ctx, `
			INSERT INTO integrations (campaign_id, blogger_id, publish_date, cost_rub, promo_code, utm_campaign, status)
			SELECT c.id, $2, $3, $4, $5, 'crm_' || c.id || '_' || $6::text, $7 FROM campaigns c WHERE c.id = $1
			RETURNING id`,
			campaignID, in.BloggerID, publish, in.CostRub, promo, handle, status).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", httpx.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}
		return id, promo, err
	}
	return 0, "", fmt.Errorf("could not generate unique promo code: %w", err)
}
