// Package worker имитирует фоновые интеграции: забирает «отчёт Ozon»
// и «заказы из CRM» за сегодня и раскладывает продажи по интеграциям.
// Внешних API нет — цифры генерируются, но шаги и логи повторяют реальный процесс.
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"crmdemo/internal/seed"
)

type Worker struct {
	pool  *pgxpool.Pool
	log   *slog.Logger
	every time.Duration
	rnd   *rand.Rand
	// pause между шагами «асинхронного отчёта», в тестах — 0.
	pause time.Duration
}

func New(pool *pgxpool.Pool, log *slog.Logger, every time.Duration) *Worker {
	return &Worker{
		pool:  pool,
		log:   log,
		every: every,
		rnd:   rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 3)),
		pause: 1500 * time.Millisecond,
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.log.Info("worker started", "every", w.every.String())
	ticker := time.NewTicker(w.every)
	defer ticker.Stop()

	jobs := []func(context.Context) error{w.pullOzon, w.pullCRM}
	for i := 0; ; i++ {
		if err := w.publishDue(ctx); err != nil {
			w.log.Error("publish due integrations", "err", err)
		}
		if err := jobs[i%len(jobs)](ctx); err != nil && ctx.Err() == nil {
			w.log.Error("job failed", "err", err)
		}
		select {
		case <-ctx.Done():
			w.log.Info("worker stopped")
			return
		case <-ticker.C:
		}
	}
}

func (w *Worker) today() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// publishDue переводит в published интеграции, чья дата наступила,
// и обновляет статусы кампаний по датам.
func (w *Worker) publishDue(ctx context.Context) error {
	today := w.today()
	tag, err := w.pool.Exec(ctx, `UPDATE integrations SET status = 'published' WHERE status = 'planned' AND publish_date <= $1`, today)
	if err != nil {
		return err
	}
	if n := tag.RowsAffected(); n > 0 {
		w.log.Info("integrations published", "count", n)
	}
	_, err = w.pool.Exec(ctx, `
		UPDATE campaigns SET status = CASE
		    WHEN $1 < start_date THEN 'planned'
		    WHEN $1 > end_date THEN 'finished'
		    ELSE 'active' END
		WHERE status IS DISTINCT FROM CASE
		    WHEN $1 < start_date THEN 'planned'
		    WHEN $1 > end_date THEN 'finished'
		    ELSE 'active' END`, today)
	return err
}

type live struct {
	ID      int64
	Promo   string
	Price   int64 // средний чек по уже известным продажам
	Reach   int
	AgeDays int
}

// liveIntegrations — интеграции, которые сегодня ещё в окне продаж.
func (w *Worker) liveIntegrations(ctx context.Context) ([]live, error) {
	today := w.today()
	rows, err := w.pool.Query(ctx, `
		SELECT i.id, i.promo_code, b.avg_reach, ($1::date - i.publish_date),
		       COALESCE((SELECT SUM(m.revenue_rub) / NULLIF(SUM(m.orders), 0)
		                 FROM integration_metrics m WHERE m.integration_id = i.id), 9000)::bigint
		FROM integrations i JOIN bloggers b ON b.id = i.blogger_id
		WHERE i.status = 'published' AND i.publish_date > $1::date - $2::int`, today, seed.SalesWindow)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (live, error) {
		var l live
		err := row.Scan(&l.ID, &l.Promo, &l.Reach, &l.AgeDays, &l.Price)
		return l, err
	})
}

func (w *Worker) sleep(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(w.pause):
		return nil
	}
}

// pullOzon повторяет реальный флоу: отчёт заказывается асинхронно,
// статус опрашивается, потом файл разбирается и раскладывается по промокодам.
func (w *Worker) pullOzon(ctx context.Context) error {
	log := w.log.With("job", "ozon")
	reportID := fmt.Sprintf("rpt-%06d", w.rnd.IntN(1_000_000))
	log.Info("отчёт по источникам трафика запрошен", "report_id", reportID)
	for _, status := range []string{"waiting", "processing", "success"} {
		if err := w.sleep(ctx); err != nil {
			return err
		}
		log.Info("статус отчёта", "report_id", reportID, "status", status)
	}
	items, err := w.liveIntegrations(ctx)
	if err != nil {
		return err
	}
	rows, matched, err := w.addSales(ctx, items, "ozon", 0.45)
	if err != nil {
		return err
	}
	log.Info("отчёт разобран", "rows", rows, "matched_promo", matched, "unmatched", rows-matched)
	return nil
}

func (w *Worker) pullCRM(ctx context.Context) error {
	log := w.log.With("job", "crm")
	log.Info("забираем заказы сайта за сегодня")
	if err := w.sleep(ctx); err != nil {
		return err
	}
	items, err := w.liveIntegrations(ctx)
	if err != nil {
		return err
	}
	rows, matched, err := w.addSales(ctx, items, "site", 0.2)
	if err != nil {
		return err
	}
	log.Info("заказы сопоставлены по UTM", "orders", rows, "matched", matched)
	return nil
}

// addSales добавляет сегодняшние продажи. Upsert суммирует с уже записанным,
// поэтому повторный прогон за день только наращивает цифры.
func (w *Worker) addSales(ctx context.Context, items []live, source string, share float64) (rows, matched int, err error) {
	today := w.today()
	batch := &pgx.Batch{}
	for _, it := range items {
		// Чем свежее публикация, тем выше шанс заказа в этом прогоне.
		chance := share * float64(it.Reach) / 30_000 / float64(1+it.AgeDays)
		orders := 0
		for chance > 0 {
			if w.rnd.Float64() < min(chance, 0.9) {
				orders++
			}
			chance--
		}
		rows += orders + w.rnd.IntN(2) // часть строк отчёта — заказы без промокода
		if orders == 0 {
			continue
		}
		matched += orders
		batch.Queue(`
			INSERT INTO integration_metrics (integration_id, day, source, orders, revenue_rub)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (integration_id, day, source) DO UPDATE
			SET orders = integration_metrics.orders + EXCLUDED.orders,
			    revenue_rub = integration_metrics.revenue_rub + EXCLUDED.revenue_rub`,
			it.ID, today, source, orders, int64(orders)*it.Price)
	}
	if batch.Len() == 0 {
		return rows, matched, nil
	}
	return rows, matched, w.pool.SendBatch(ctx, batch).Close()
}
