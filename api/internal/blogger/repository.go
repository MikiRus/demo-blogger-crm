package blogger

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) Platforms(ctx context.Context) ([]Platform, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name FROM platforms ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Platform, error) {
		var p Platform
		err := row.Scan(&p.ID, &p.Code, &p.Name)
		return p, err
	})
}

// listSQL считает расходы и выручку отдельными CTE: join интеграций
// с метриками напрямую размножил бы стоимость на число дней.
const listSQL = `
WITH cost AS (
    SELECT blogger_id, COUNT(*) AS n, SUM(cost_rub)::bigint AS cost
    FROM integrations GROUP BY blogger_id
), sales AS (
    SELECT i.blogger_id, SUM(m.revenue_rub)::bigint AS revenue, SUM(m.orders)::bigint AS orders
    FROM integration_metrics m JOIN integrations i ON i.id = m.integration_id
    GROUP BY i.blogger_id
), agg AS (
    SELECT b.id, b.name, b.handle, b.followers, b.avg_reach, b.category,
           p.id AS platform_id, p.code AS platform_code, p.name AS platform_name,
           COALESCE(c.n, 0) AS n, COALESCE(c.cost, 0) AS cost,
           COALESCE(s.revenue, 0) AS revenue, COALESCE(s.orders, 0) AS orders
    FROM bloggers b
    JOIN platforms p ON p.id = b.platform_id
    LEFT JOIN cost c ON c.blogger_id = b.id
    LEFT JOIN sales s ON s.blogger_id = b.id
)
SELECT b.id, b.name, b.handle, b.followers, b.avg_reach, b.category,
       b.platform_id, b.platform_code, b.platform_name,
       b.n, b.cost, b.revenue, b.orders, COUNT(*) OVER () AS total
FROM agg b
WHERE %s
ORDER BY %s
LIMIT $1 OFFSET $2`

func scanListItem(row pgx.Row, total *int64) (ListItem, error) {
	var (
		it                   ListItem
		cost, revenue, order int64
	)
	err := row.Scan(&it.ID, &it.Name, &it.Handle, &it.Followers, &it.AvgReach, &it.Category,
		&it.Platform.ID, &it.Platform.Code, &it.Platform.Name,
		&it.Integrations, &cost, &revenue, &order, total)
	it.Totals = metrics.NewTotals(cost, revenue, order)
	return it, err
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]ListItem, int64, error) {
	where := "TRUE"
	args := []any{f.Limit, f.Offset}
	if f.Platform != "" {
		args = append(args, f.Platform)
		where += fmt.Sprintf(" AND b.platform_code = $%d", len(args))
	}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		where += fmt.Sprintf(" AND (b.name ILIKE $%[1]d OR b.handle ILIKE $%[1]d)", len(args))
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(listSQL, where, f.orderBy()), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var (
		items []ListItem
		total int64
	)
	for rows.Next() {
		it, err := scanListItem(rows, &total)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id int64) (Details, error) {
	var (
		d     Details
		total int64
	)
	row := r.pool.QueryRow(ctx, fmt.Sprintf(listSQL, "b.id = $3", "b.id"), 1, 0, id)
	it, err := scanListItem(row, &total)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, httpx.ErrNotFound
	}
	if err != nil {
		return d, err
	}
	d.ListItem = it

	rows, err := r.pool.Query(ctx, `
		SELECT i.id, c.id, c.title, i.publish_date::text, i.promo_code, i.status, i.cost_rub,
		       COALESCE(SUM(m.revenue_rub), 0)::bigint, COALESCE(SUM(m.orders), 0)::bigint
		FROM integrations i
		JOIN campaigns c ON c.id = i.campaign_id
		LEFT JOIN integration_metrics m ON m.integration_id = i.id
		WHERE i.blogger_id = $1
		GROUP BY i.id, c.id
		ORDER BY i.publish_date DESC`, id)
	if err != nil {
		return d, err
	}
	d.History, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (IntegrationRow, error) {
		var (
			h                     IntegrationRow
			cost, revenue, orders int64
		)
		err := row.Scan(&h.ID, &h.CampaignID, &h.Campaign, &h.PublishDate, &h.PromoCode, &h.Status, &cost, &revenue, &orders)
		h.Totals = metrics.NewTotals(cost, revenue, orders)
		return h, err
	})
	return d, err
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO bloggers (name, handle, platform_id, followers, avg_reach, category)
		SELECT $1, $2, p.id, $4, $5, $6 FROM platforms p WHERE p.code = $3
		RETURNING id`,
		in.Name, in.Handle, in.Platform, in.Followers, in.AvgReach, in.Category).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, httpx.ValidationError{Msg: "platform: неизвестная площадка"}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, httpx.ValidationError{Msg: "handle: такой блогер уже есть"}
	}
	return id, err
}
