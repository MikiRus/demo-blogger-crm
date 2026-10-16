package seed

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

// ozonRunAt — во сколько по UTC планировщик заказывает отчёт за прошедший день (06:00 МСК).
const ozonRunAt = 3 * time.Hour

// insertOzonReports строит журнал ежедневных загрузок за всю историю продаж:
// цифры журнала совпадают с тем, что лежит в integration_metrics по источнику ozon.
func insertOzonReports(ctx context.Context, tx pgx.Tx, rnd *rand.Rand, metricRows [][]any, today time.Time) (int, error) {
	type agg struct{ orders, revenue int64 }
	byDay := map[time.Time]*agg{}
	first := today
	for _, r := range metricRows {
		if r[2] != "ozon" {
			continue
		}
		day := r[1].(time.Time)
		if byDay[day] == nil {
			byDay[day] = &agg{}
		}
		byDay[day].orders += int64(r[3].(int))
		byDay[day].revenue += r[4].(int64)
		if day.Before(first) {
			first = day
		}
	}

	var rows [][]any
	for day := first; day.Before(today); day = day.AddDate(0, 0, 1) {
		a := byDay[day]
		if a == nil {
			a = &agg{}
		}
		requested := day.AddDate(0, 0, 1).Add(ozonRunAt + time.Duration(rnd.IntN(240))*time.Second)
		attempts, errText := 1, any(nil)
		// Изредка Ozon отвечает 429 — планировщик повторяет через 15 минут.
		if rnd.IntN(25) == 0 {
			attempts, errText = 2, "Ozon API: 429 Too Many Requests"
			requested = requested.Add(15 * time.Minute)
		}
		finished := requested.Add(time.Duration(40+rnd.IntN(140)) * time.Second)
		rows = append(rows, []any{
			day, fmt.Sprintf("rpt-%06d", rnd.IntN(1_000_000)), "success", "schedule", attempts,
			int(a.orders) + 30 + rnd.IntN(90), int(a.orders), a.revenue, errText, requested, finished,
		})
	}
	n, err := tx.CopyFrom(ctx, pgx.Identifier{"ozon_reports"},
		[]string{"day", "report_id", "status", "trigger", "attempts", "rows_total", "orders", "revenue_rub", "error", "requested_at", "finished_at"},
		pgx.CopyFromRows(rows))
	return int(n), err
}

var users = []struct {
	Name, Email, Role, Status string
	SeenAgo                   time.Duration // 0 — ещё не заходил
}{
	{"Анна Смирнова", "anna.smirnova@example.com", "admin", "active", 3 * time.Minute},
	{"Игорь Белов", "igor.belov@example.com", "marketer", "active", 40 * time.Minute},
	{"Мария Козлова", "maria.kozlova@example.com", "manager", "active", 2 * time.Hour},
	{"Олег Петров", "oleg.petrov@example.com", "manager", "active", 26 * time.Hour},
	{"Елена Волкова", "elena.volkova@example.com", "analyst", "active", 5 * time.Hour},
	{"Сергей Новиков", "sergey.novikov@example.com", "finance", "active", 4 * 24 * time.Hour},
	{"Дарья Орлова", "daria.orlova@example.com", "manager", "invited", 0},
	{"Павел Зайцев", "pavel.zaitsev@example.com", "marketer", "blocked", 62 * 24 * time.Hour},
}

func insertUsers(ctx context.Context, tx pgx.Tx, now time.Time) (int, error) {
	rows := make([][]any, 0, len(users))
	for _, u := range users {
		var seen any
		if u.SeenAgo > 0 {
			seen = now.Add(-u.SeenAgo)
		}
		rows = append(rows, []any{u.Name, u.Email, u.Role, u.Status, seen})
	}
	n, err := tx.CopyFrom(ctx, pgx.Identifier{"users"}, []string{"name", "email", "role", "status", "last_seen_at"}, pgx.CopyFromRows(rows))
	return int(n), err
}
