package report

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Promos(ctx context.Context, codes []string) (map[string]PromoInfo, error) {
	out := map[string]PromoInfo{}
	if len(codes) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT i.promo_code, i.id, b.name
		FROM integrations i JOIN bloggers b ON b.id = i.blogger_id
		WHERE i.promo_code = ANY($1)`, codes)
	if err != nil {
		return nil, err
	}
	var (
		code string
		info PromoInfo
	)
	_, err = pgx.ForEachRow(rows, []any{&code, &info.IntegrationID, &info.Blogger}, func() error {
		out[code] = info
		return nil
	})
	return out, err
}

// Existing возвращает уже загруженные (интеграция, день, источник) среди ids.
func (r *Repository) Existing(ctx context.Context, ids []int64) (map[Key]bool, error) {
	out := map[Key]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT integration_id, day::text, source FROM integration_metrics
		WHERE integration_id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	var k Key
	_, err = pgx.ForEachRow(rows, []any{&k.IntegrationID, &k.Day, &k.Source}, func() error {
		out[k] = true
		return nil
	})
	return out, err
}

// Insert пишет строки одной транзакцией. ON CONFLICT DO NOTHING — страховка
// от гонки двух одновременных загрузок: уже записанный день не перезаписывается.
func (r *Repository) Insert(ctx context.Context, rows []PreviewRow) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, row := range rows {
		batch.Queue(`
			INSERT INTO integration_metrics (integration_id, day, source, orders, revenue_rub)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING`,
			row.IntegrationID, row.Day, row.Source, row.Orders, row.RevenueRub)
	}
	res := tx.SendBatch(ctx, batch)
	inserted := 0
	for range rows {
		tag, err := res.Exec()
		if err != nil {
			res.Close()
			return 0, err
		}
		inserted += int(tag.RowsAffected())
	}
	if err := res.Close(); err != nil {
		return 0, err
	}
	return inserted, tx.Commit(ctx)
}

type sampleSource struct {
	Promo   string
	Reach   int
	Publish time.Time
}

// SampleIntegrations — опубликованные интеграции с продажами, из которых
// собирается демонстрационный отчёт WB.
func (r *Repository) SampleIntegrations(ctx context.Context, limit int) ([]sampleSource, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.promo_code, b.avg_reach, i.publish_date
		FROM integrations i JOIN bloggers b ON b.id = i.blogger_id
		WHERE i.status = 'published'
		ORDER BY i.publish_date DESC, i.id
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (sampleSource, error) {
		var s sampleSource
		err := row.Scan(&s.Promo, &s.Reach, &s.Publish)
		return s, err
	})
}
