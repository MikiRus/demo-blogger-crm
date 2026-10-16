// Package worker имитирует фоновые интеграции: скачивает «отчёт Ozon»
// и «заказы из CRM» за сегодня и раскладывает продажи по интеграциям.
// Внешних API нет — цифры генерируются, но шаги и логи повторяют реальный процесс.
//
// Отчёты Ozon идут через очередь в таблице ozon_reports: расписание и кнопка
// в интерфейсе только ставят день в очередь, а забирает его этот воркер.
package worker

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"crmdemo/internal/ozon"
	"crmdemo/internal/seed"
)

// queuePoll — как часто воркер проверяет очередь отчётов.
const queuePoll = 2 * time.Second

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
	schedule := time.NewTicker(w.every)
	defer schedule.Stop()
	poll := time.NewTicker(queuePoll)
	defer poll.Stop()

	w.scheduled(ctx)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker stopped")
			return
		case <-schedule.C:
			w.scheduled(ctx)
		case <-poll.C:
		}
		// Разбираем очередь до конца: ручной запуск не ждёт следующего тика.
		for {
			done, err := w.processOzon(ctx)
			if err != nil && ctx.Err() == nil {
				w.log.Error("ozon report failed", "err", err)
			}
			if done || err != nil {
				break
			}
		}
	}
}

// scheduled — плановые задачи: публикация интеграций, заказы сайта
// и постановка сегодняшнего отчёта Ozon в очередь.
func (w *Worker) scheduled(ctx context.Context) {
	if err := w.publishDue(ctx); err != nil {
		w.log.Error("publish due integrations", "err", err)
	}
	if err := ozon.Enqueue(ctx, w.pool, w.today(), "schedule"); err != nil && !errors.Is(err, ozon.ErrBusy) {
		w.log.Error("enqueue ozon report", "err", err)
	}
	if err := w.pullCRM(ctx); err != nil && ctx.Err() == nil {
		w.log.Error("crm job failed", "err", err)
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

// processOzon забирает из очереди один отчёт и повторяет реальный флоу:
// отчёт заказывается асинхронно, статус опрашивается, потом файл разбирается
// и раскладывается по промокодам. done=true — очередь пуста.
func (w *Worker) processOzon(ctx context.Context) (done bool, err error) {
	var (
		day      time.Time
		reportID string
	)
	// SKIP LOCKED: несколько воркеров не возьмут один и тот же день.
	err = w.pool.QueryRow(ctx, `
		UPDATE ozon_reports SET status = 'processing', attempts = attempts + 1
		WHERE day = (SELECT day FROM ozon_reports WHERE status = 'queued'
		             ORDER BY requested_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		RETURNING day, report_id`).Scan(&day, &reportID)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	log := w.log.With("job", "ozon", "report_id", reportID, "day", day.Format(time.DateOnly))
	log.Info("отчёт по продажам запрошен")
	for _, status := range []string{"waiting", "processing", "success"} {
		if err := w.sleep(ctx); err != nil {
			return false, w.failOzon(day, "загрузка прервана: воркер остановлен")
		}
		// Иногда Ozon отвечает 429 — день останется с ошибкой, расписание повторит.
		if status == "processing" && w.rnd.IntN(12) == 0 {
			log.Warn("Ozon ответил 429, повторим по расписанию")
			return false, w.failOzon(day, "Ozon API: 429 Too Many Requests")
		}
		log.Info("статус отчёта", "status", status)
	}

	rows := 0
	if day.Equal(w.today()) {
		items, err := w.liveIntegrations(ctx)
		if err != nil {
			return false, errors.Join(err, w.failOzon(day, "не удалось разобрать отчёт"))
		}
		var matched int
		if rows, matched, err = w.addSales(ctx, items, "ozon", 0.45); err != nil {
			return false, errors.Join(err, w.failOzon(day, "не удалось записать продажи"))
		}
		log.Info("отчёт разобран", "rows", rows, "matched_promo", matched, "unmatched", rows-matched)
	}

	// Итоги дня пересчитываются из integration_metrics, а не суммируются,
	// поэтому журнал всегда сходится с тем, что видно в аналитике.
	_, err = w.pool.Exec(ctx, `
		UPDATE ozon_reports r SET status = 'success', finished_at = now(),
		       rows_total = r.rows_total + $2, orders = s.orders, revenue_rub = s.revenue
		FROM (SELECT COALESCE(SUM(orders), 0)::int AS orders, COALESCE(SUM(revenue_rub), 0)::bigint AS revenue
		      FROM integration_metrics WHERE day = $1 AND source = 'ozon') s
		WHERE r.day = $1`, day, rows)
	return false, err
}

// failOzon помечает отчёт ошибкой. Контекст отдельный: при остановке воркера
// день не должен навсегда застрять в processing.
func (w *Worker) failOzon(day time.Time, msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := w.pool.Exec(ctx, `
		UPDATE ozon_reports SET status = 'error', error = $2, finished_at = now() WHERE day = $1`, day, msg)
	return err
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
