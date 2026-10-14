// Package seed наполняет базу выдуманными, но правдоподобными данными.
// Даты строятся от текущего дня, поэтому демо всегда выглядит «живым».
package seed

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WBLoadedThrough — отчёт WB в сиде загружен по today-4; свежие дни
// пользователь догружает сам через XLSX (как в реальной работе с WB).
const WBLoadedThrough = 4

// SalesWindow — сколько дней после публикации интеграция приносит продажи.
const SalesWindow = 14

type platform struct {
	Code, Name string
	// weekday[0] = понедельник: в какие дни аудитория площадки покупает активнее.
	Weekday [7]float64
}

var platforms = []platform{
	{"telegram", "Telegram", [7]float64{1.2, 1.15, 1.1, 1.1, 1.0, 0.75, 0.7}},
	{"youtube", "YouTube", [7]float64{0.8, 0.85, 0.9, 0.9, 1.0, 1.3, 1.35}},
	{"vk", "VK Видео", [7]float64{1.0, 1.0, 1.0, 1.05, 1.1, 1.0, 0.85}},
	{"dzen", "Дзен", [7]float64{1.1, 1.1, 1.05, 1.0, 0.9, 0.9, 1.0}},
	{"rutube", "Rutube", [7]float64{0.9, 0.9, 0.95, 1.0, 1.05, 1.15, 1.1}},
}

var (
	firstNames = []string{"Анна", "Мария", "Ольга", "Дарья", "Екатерина", "Ирина", "Полина", "Алина", "Ксения", "Светлана",
		"Дмитрий", "Алексей", "Илья", "Никита", "Павел", "Сергей", "Артём", "Егор"}
	lastInitials = []string{"А.", "Б.", "В.", "Г.", "Д.", "Е.", "З.", "К.", "Л.", "М.", "Н.", "О.", "П.", "Р.", "С.", "Т."}
	handleA      = []string{"kitchen", "home", "smart", "tasty", "cozy", "tech", "daily", "family", "fresh", "urban", "slow", "happy"}
	handleB      = []string{"lab", "mom", "chef", "review", "life", "notes", "geek", "story", "space", "talks"}
	categories   = []string{"Кулинария", "Обзоры техники", "Дом и уют", "Семья", "Лайфстайл", "ЗОЖ"}
)

type product struct {
	Name     string
	PriceRub int
}

var products = []product{
	{"Кофемашина Aroma One", 32_990},
	{"Блендер Vortex 900", 8_490},
	{"Мультиварка ChefPot", 11_990},
	{"Пылесос AirClean S", 18_990},
	{"Чайник Glass Pro", 4_290},
	{"Аэрогриль Crispy", 9_990},
	{"Кухонная машина Bake Master", 27_490},
	{"Соковыжималка Fresh Line", 7_990},
}

type blogger struct {
	ID       int64
	Platform int // индекс в platforms
	Reach    int
	Quality  float64 // насколько аудитория блогера конвертируется в покупки
}

// Run полностью пересоздаёт демо-данные.
func Run(ctx context.Context, pool *pgxpool.Pool, now time.Time) (Stats, error) {
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	rnd := rand.New(rand.NewPCG(2026, 10))

	tx, err := pool.Begin(ctx)
	if err != nil {
		return Stats{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `TRUNCATE integration_metrics, integrations, campaigns, bloggers, platforms RESTART IDENTITY CASCADE`); err != nil {
		return Stats{}, fmt.Errorf("truncate: %w", err)
	}

	platformIDs := make([]int16, len(platforms))
	for i, p := range platforms {
		if err := tx.QueryRow(ctx, `INSERT INTO platforms (code, name) VALUES ($1, $2) RETURNING id`, p.Code, p.Name).Scan(&platformIDs[i]); err != nil {
			return Stats{}, err
		}
	}

	bloggers, err := insertBloggers(ctx, tx, rnd, platformIDs)
	if err != nil {
		return Stats{}, err
	}

	var st Stats
	st.Bloggers = len(bloggers)
	var metricRows [][]any
	for _, c := range planCampaigns(rnd, today) {
		var campaignID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO campaigns (title, product, status, start_date, end_date, budget_rub)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			c.Title, c.Product.Name, c.Status, c.Start, c.End, c.Budget).Scan(&campaignID)
		if err != nil {
			return Stats{}, err
		}
		st.Campaigns++

		picked := rnd.Perm(len(bloggers))[:c.Integrations]
		for k, bi := range picked {
			b := bloggers[bi]
			publish := c.publishDate(k, rnd)
			status := "published"
			if publish.After(today) {
				status = "planned"
			}
			cost := roundTo(float64(b.Reach)*(0.6+rnd.Float64()*1.4), 1000)
			promo := fmt.Sprintf("%s-%04d", promoPrefix(campaignID), 1000+st.Integrations)
			var integrationID int64
			err := tx.QueryRow(ctx, `
				INSERT INTO integrations (campaign_id, blogger_id, publish_date, cost_rub, promo_code, utm_campaign, status)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
				campaignID, b.ID, publish, cost, promo, fmt.Sprintf("crm_%d_%d", campaignID, b.ID), status).Scan(&integrationID)
			if err != nil {
				return Stats{}, err
			}
			st.Integrations++
			if status == "published" {
				metricRows = append(metricRows, salesRows(rnd, integrationID, b, c.Product, publish, today)...)
			}
		}
	}

	n, err := tx.CopyFrom(ctx, pgx.Identifier{"integration_metrics"},
		[]string{"integration_id", "day", "source", "orders", "revenue_rub"}, pgx.CopyFromRows(metricRows))
	if err != nil {
		return Stats{}, fmt.Errorf("copy metrics: %w", err)
	}
	st.MetricRows = int(n)
	return st, tx.Commit(ctx)
}

type Stats struct {
	Bloggers, Campaigns, Integrations, MetricRows int
}

func insertBloggers(ctx context.Context, tx pgx.Tx, rnd *rand.Rand, platformIDs []int16) ([]blogger, error) {
	used := map[string]bool{}
	var out []blogger
	for len(out) < 36 {
		handle := handleA[rnd.IntN(len(handleA))] + "_" + handleB[rnd.IntN(len(handleB))]
		if used[handle] {
			continue
		}
		used[handle] = true

		// Подписчики распределены логарифмически: много микроблогеров, мало миллионников.
		followers := int(math.Exp(math.Log(15_000) + rnd.Float64()*(math.Log(1_500_000)-math.Log(15_000))))
		reach := int(float64(followers) * (0.08 + rnd.Float64()*0.27))
		pi := rnd.IntN(len(platforms))
		b := blogger{Platform: pi, Reach: reach, Quality: 0.25 + rnd.Float64()*1.9}

		name := firstNames[rnd.IntN(len(firstNames))] + " " + lastInitials[rnd.IntN(len(lastInitials))]
		err := tx.QueryRow(ctx, `
			INSERT INTO bloggers (name, handle, platform_id, followers, avg_reach, category)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			name, handle, platformIDs[pi], followers, reach, categories[rnd.IntN(len(categories))]).Scan(&b.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

type campaignPlan struct {
	Title        string
	Product      product
	Status       string
	Start, End   time.Time
	Budget       int64
	Integrations int
}

// publishDate раскладывает интеграции равномерно по окну кампании с небольшим разбросом.
func (c campaignPlan) publishDate(k int, rnd *rand.Rand) time.Time {
	span := int(c.End.Sub(c.Start).Hours() / 24)
	step := float64(span) / float64(c.Integrations)
	day := int(float64(k)*step) + rnd.IntN(max(1, int(step)))
	return c.Start.AddDate(0, 0, min(day, span))
}

var seasons = []string{"Старт сезона", "Распродажа", "Новинка", "Подарки", "Неделя бренда", "Back to school"}

// planCampaigns: кампании за последние полгода, одна идёт прямо сейчас
// и одна запланирована, чтобы в демо были все статусы.
func planCampaigns(rnd *rand.Rand, today time.Time) []campaignPlan {
	var out []campaignPlan
	start := today.AddDate(0, 0, -180)
	for i := 0; i < 9; i++ {
		p := products[i%len(products)]
		length := 21 + rnd.IntN(15)
		out = append(out, campaignPlan{
			Title:        seasons[i%len(seasons)] + ": " + p.Name,
			Product:      p,
			Start:        start,
			End:          start.AddDate(0, 0, length),
			Integrations: 6 + rnd.IntN(7),
		})
		start = start.AddDate(0, 0, 15+rnd.IntN(8))
		if start.After(today.AddDate(0, 0, -40)) {
			break
		}
	}
	out = append(out,
		campaignPlan{Title: "Осенний запуск: " + products[0].Name, Product: products[0],
			Start: today.AddDate(0, 0, -13), End: today.AddDate(0, 0, 14), Integrations: 12},
		campaignPlan{Title: "Подарки к Новому году: " + products[6].Name, Product: products[6],
			Start: today.AddDate(0, 0, 21), End: today.AddDate(0, 0, 50), Integrations: 8},
	)
	for i := range out {
		c := &out[i]
		switch {
		case today.Before(c.Start):
			c.Status = "planned"
		case today.After(c.End):
			c.Status = "finished"
		default:
			c.Status = "active"
		}
		c.Budget = int64(c.Integrations) * 120_000
	}
	return out
}

// salesRows — продажи интеграции по дням: пик в день публикации и затухание.
// ozon и site — по вчера (сегодня их «докачивает» воркер), wb — по today-4.
func salesRows(rnd *rand.Rand, id int64, b blogger, p product, publish, today time.Time) [][]any {
	expected := float64(b.Reach) * 0.0003 * b.Quality * (11_000 / float64(p.PriceRub+3_000))
	share := map[string]float64{"ozon": 0.45, "wb": 0.35, "site": 0.20}
	var rows [][]any
	for k := 0; k < SalesWindow; k++ {
		day := publish.AddDate(0, 0, k)
		weekday := platforms[b.Platform].Weekday[(int(day.Weekday())+6)%7]
		decay := math.Exp(-float64(k)/3) * 0.31
		for _, src := range []string{"ozon", "wb", "site"} {
			limit := today.AddDate(0, 0, -1)
			if src == "wb" {
				limit = today.AddDate(0, 0, -WBLoadedThrough)
			}
			if day.After(limit) {
				continue
			}
			mean := expected * decay * weekday * share[src]
			orders := int(math.Round(mean * (0.6 + rnd.Float64()*0.8)))
			price := float64(p.PriceRub) * (0.88 + rnd.Float64()*0.17)
			rows = append(rows, []any{id, day, src, orders, int64(float64(orders) * price)})
		}
	}
	return rows
}

func promoPrefix(campaignID int64) string {
	return fmt.Sprintf("C%02d", campaignID)
}

func roundTo(v float64, step int64) int64 {
	return int64(math.Round(v/float64(step))) * step
}
