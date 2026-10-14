// Package analytics — сводка окупаемости интеграций для дашборда.
//
// Расходы относятся к периоду по дате публикации, выручка — по дню продажи:
// так дашборд отвечает на вопрос «сколько потратили и сколько заработали в месяце».
package analytics

import (
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"crmdemo/internal/metrics"
)

type Month struct {
	Month string `json:"month"` // YYYY-MM
	metrics.Totals
}

type TopBlogger struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Handle   string `json:"handle"`
	Platform string `json:"platform"`
	metrics.Totals
}

type HeatCell struct {
	Weekday    int    `json:"weekday"` // 1 = понедельник … 7 = воскресенье
	Platform   string `json:"platform"`
	RevenueRub int64  `json:"revenueRub"`
}

type SourceRow struct {
	Source     string `json:"source"`
	Orders     int64  `json:"orders"`
	RevenueRub int64  `json:"revenueRub"`
}

type Summary struct {
	From        string         `json:"from"`
	To          string         `json:"to"`
	Total       metrics.Totals `json:"total"`
	Monthly     []Month        `json:"monthly"`
	TopBloggers []TopBlogger   `json:"topBloggers"`
	Heatmap     []HeatCell     `json:"heatmap"`
	BySource    []SourceRow    `json:"bySource"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type monthAgg struct{ cost, revenue, orders int64 }

func (r *Repository) Summary(ctx context.Context, from, to time.Time) (Summary, error) {
	s := Summary{From: from.Format(time.DateOnly), To: to.Format(time.DateOnly)}

	months := map[string]*monthAgg{}
	get := func(m string) *monthAgg {
		if months[m] == nil {
			months[m] = &monthAgg{}
		}
		return months[m]
	}

	rows, err := r.pool.Query(ctx, `
		SELECT to_char(publish_date, 'YYYY-MM'), SUM(cost_rub)::bigint
		FROM integrations WHERE publish_date BETWEEN $1 AND $2
		GROUP BY 1`, from, to)
	if err != nil {
		return s, err
	}
	var (
		month string
		v1    int64
		v2    int64
	)
	if _, err := pgx.ForEachRow(rows, []any{&month, &v1}, func() error {
		get(month).cost = v1
		return nil
	}); err != nil {
		return s, err
	}

	rows, err = r.pool.Query(ctx, `
		SELECT to_char(day, 'YYYY-MM'), SUM(revenue_rub)::bigint, SUM(orders)::bigint
		FROM integration_metrics WHERE day BETWEEN $1 AND $2
		GROUP BY 1`, from, to)
	if err != nil {
		return s, err
	}
	if _, err := pgx.ForEachRow(rows, []any{&month, &v1, &v2}, func() error {
		a := get(month)
		a.revenue, a.orders = v1, v2
		return nil
	}); err != nil {
		return s, err
	}
	s.Monthly, s.Total = buildMonthly(months)

	rows, err = r.pool.Query(ctx, `
		WITH sales AS (
		    SELECT i.blogger_id, SUM(m.revenue_rub)::bigint AS revenue, SUM(m.orders)::bigint AS orders
		    FROM integration_metrics m JOIN integrations i ON i.id = m.integration_id
		    WHERE m.day BETWEEN $1 AND $2
		    GROUP BY i.blogger_id
		), cost AS (
		    SELECT blogger_id, SUM(cost_rub)::bigint AS cost
		    FROM integrations WHERE publish_date BETWEEN $1 AND $2
		    GROUP BY blogger_id
		)
		SELECT b.id, b.name, b.handle, p.code, COALESCE(c.cost, 0), s.revenue, s.orders
		FROM sales s
		JOIN bloggers b ON b.id = s.blogger_id
		JOIN platforms p ON p.id = b.platform_id
		LEFT JOIN cost c ON c.blogger_id = s.blogger_id
		ORDER BY s.revenue DESC, b.id
		LIMIT 5`, from, to)
	if err != nil {
		return s, err
	}
	s.TopBloggers, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (TopBlogger, error) {
		var (
			t                     TopBlogger
			cost, revenue, orders int64
		)
		err := row.Scan(&t.ID, &t.Name, &t.Handle, &t.Platform, &cost, &revenue, &orders)
		t.Totals = metrics.NewTotals(cost, revenue, orders)
		return t, err
	})
	if err != nil {
		return s, err
	}

	rows, err = r.pool.Query(ctx, `
		SELECT EXTRACT(ISODOW FROM m.day)::int, p.code, SUM(m.revenue_rub)::bigint
		FROM integration_metrics m
		JOIN integrations i ON i.id = m.integration_id
		JOIN bloggers b ON b.id = i.blogger_id
		JOIN platforms p ON p.id = b.platform_id
		WHERE m.day BETWEEN $1 AND $2
		GROUP BY 1, 2 ORDER BY 2, 1`, from, to)
	if err != nil {
		return s, err
	}
	s.Heatmap, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (HeatCell, error) {
		var c HeatCell
		err := row.Scan(&c.Weekday, &c.Platform, &c.RevenueRub)
		return c, err
	})
	if err != nil {
		return s, err
	}

	rows, err = r.pool.Query(ctx, `
		SELECT source, SUM(orders)::bigint, SUM(revenue_rub)::bigint
		FROM integration_metrics WHERE day BETWEEN $1 AND $2
		GROUP BY source ORDER BY 3 DESC`, from, to)
	if err != nil {
		return s, err
	}
	s.BySource, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (SourceRow, error) {
		var x SourceRow
		err := row.Scan(&x.Source, &x.Orders, &x.RevenueRub)
		return x, err
	})
	return s, err
}

// buildMonthly сортирует месяцы и считает итог по периоду.
func buildMonthly(months map[string]*monthAgg) ([]Month, metrics.Totals) {
	keys := make([]string, 0, len(months))
	for k := range months {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]Month, 0, len(keys))
	var total monthAgg
	for _, k := range keys {
		a := months[k]
		out = append(out, Month{Month: k, Totals: metrics.NewTotals(a.cost, a.revenue, a.orders)})
		total.cost += a.cost
		total.revenue += a.revenue
		total.orders += a.orders
	}
	return out, metrics.NewTotals(total.cost, total.revenue, total.orders)
}
