// Package campaign — рекламные кампании и интеграции блогеров в них.
package campaign

import (
	"crypto/rand"
	"strings"
	"time"

	"crmdemo/internal/httpx"
	"crmdemo/internal/metrics"
)

var statuses = map[string]bool{"planned": true, "active": true, "finished": true}

type Campaign struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Product   string `json:"product"`
	Status    string `json:"status"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	BudgetRub int64  `json:"budgetRub"`
}

type ListItem struct {
	Campaign
	Integrations int64 `json:"integrations"`
	metrics.Totals
}

type Integration struct {
	ID          int64  `json:"id"`
	BloggerID   int64  `json:"bloggerId"`
	Blogger     string `json:"blogger"`
	Handle      string `json:"handle"`
	Platform    string `json:"platform"`
	PublishDate string `json:"publishDate"`
	PromoCode   string `json:"promoCode"`
	UTM         string `json:"utmCampaign"`
	Status      string `json:"status"`
	metrics.Totals
}

type DayPoint struct {
	Day        string `json:"day"`
	Orders     int64  `json:"orders"`
	RevenueRub int64  `json:"revenueRub"`
}

type Details struct {
	ListItem
	Integrations []Integration `json:"integrationList"`
	Daily        []DayPoint    `json:"daily"`
}

type CreateInput struct {
	Title     string `json:"title"`
	Product   string `json:"product"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	BudgetRub int64  `json:"budgetRub"`
}

func (in *CreateInput) Validate() (start, end time.Time, err error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Product = strings.TrimSpace(in.Product)
	if in.Title == "" || len(in.Title) > 160 {
		return start, end, httpx.ValidationError{Msg: "title: от 1 до 160 символов"}
	}
	if in.Product == "" {
		return start, end, httpx.ValidationError{Msg: "product: обязательное поле"}
	}
	if in.BudgetRub < 0 {
		return start, end, httpx.ValidationError{Msg: "budgetRub не может быть отрицательным"}
	}
	if start, err = time.Parse(time.DateOnly, in.StartDate); err != nil {
		return start, end, httpx.ValidationError{Msg: "startDate: ожидается YYYY-MM-DD"}
	}
	if end, err = time.Parse(time.DateOnly, in.EndDate); err != nil {
		return start, end, httpx.ValidationError{Msg: "endDate: ожидается YYYY-MM-DD"}
	}
	if end.Before(start) {
		return start, end, httpx.ValidationError{Msg: "endDate раньше startDate"}
	}
	return start, end, nil
}

// StatusFor выводит статус кампании из её дат относительно today.
func StatusFor(start, end, today time.Time) string {
	switch {
	case today.Before(start):
		return "planned"
	case today.After(end):
		return "finished"
	default:
		return "active"
	}
}

type AddIntegrationInput struct {
	BloggerID   int64  `json:"bloggerId"`
	PublishDate string `json:"publishDate"`
	CostRub     int64  `json:"costRub"`
}

func (in AddIntegrationInput) Validate() (time.Time, error) {
	if in.BloggerID <= 0 {
		return time.Time{}, httpx.ValidationError{Msg: "bloggerId: обязательное поле"}
	}
	if in.CostRub < 0 {
		return time.Time{}, httpx.ValidationError{Msg: "costRub не может быть отрицательным"}
	}
	d, err := time.Parse(time.DateOnly, in.PublishDate)
	if err != nil {
		return d, httpx.ValidationError{Msg: "publishDate: ожидается YYYY-MM-DD"}
	}
	return d, nil
}

// promoAlphabet без похожих символов (0/O, 1/I), чтобы промокод легко продиктовать.
const promoAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// NewPromoCode строит промокод вида HANDLE-XXXX из ника блогера.
func NewPromoCode(handle string) string {
	prefix := strings.ToUpper(strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, handle))
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	if prefix == "" {
		prefix = "PROMO"
	}
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	for i := range buf {
		buf[i] = promoAlphabet[int(buf[i])%len(promoAlphabet)]
	}
	return prefix + "-" + string(buf)
}
