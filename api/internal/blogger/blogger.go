// Package blogger — справочник блогеров и их статистика по интеграциям.
package blogger

import (
	"strings"

	"crmdemo/internal/httpx"
	"crmdemo/internal/metrics"
)

type Platform struct {
	ID   int16  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Blogger struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Handle    string   `json:"handle"`
	Platform  Platform `json:"platform"`
	Followers int      `json:"followers"`
	AvgReach  int      `json:"avgReach"`
	Category  string   `json:"category"`
}

type ListItem struct {
	Blogger
	Integrations int64 `json:"integrations"`
	metrics.Totals
}

type IntegrationRow struct {
	ID          int64  `json:"id"`
	CampaignID  int64  `json:"campaignId"`
	Campaign    string `json:"campaign"`
	PublishDate string `json:"publishDate"`
	PromoCode   string `json:"promoCode"`
	Status      string `json:"status"`
	metrics.Totals
}

type Details struct {
	ListItem
	History []IntegrationRow `json:"history"`
}

type ListFilter struct {
	Platform string
	Query    string
	Sort     string
	Limit    int
	Offset   int
}

// sortColumns — белый список сортировок: в ORDER BY попадает только отсюда.
var sortColumns = map[string]string{
	"revenue":   "revenue DESC, b.id",
	"roas":      "revenue::float8 / NULLIF(cost, 0) DESC NULLS LAST, b.id",
	"followers": "b.followers DESC, b.id",
	"name":      "b.name, b.id",
}

func (f ListFilter) orderBy() string {
	if s, ok := sortColumns[f.Sort]; ok {
		return s
	}
	return sortColumns["revenue"]
}

type CreateInput struct {
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	Platform  string `json:"platform"`
	Followers int    `json:"followers"`
	AvgReach  int    `json:"avgReach"`
	Category  string `json:"category"`
}

func (in *CreateInput) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Handle = strings.TrimPrefix(strings.TrimSpace(in.Handle), "@")
	in.Category = strings.TrimSpace(in.Category)
	switch {
	case in.Name == "" || len(in.Name) > 120:
		return httpx.ValidationError{Msg: "name: от 1 до 120 символов"}
	case in.Handle == "" || len(in.Handle) > 64:
		return httpx.ValidationError{Msg: "handle: от 1 до 64 символов"}
	case in.Platform == "":
		return httpx.ValidationError{Msg: "platform: обязательное поле"}
	case in.Followers < 0 || in.AvgReach < 0:
		return httpx.ValidationError{Msg: "followers и avgReach не могут быть отрицательными"}
	case in.Category == "":
		return httpx.ValidationError{Msg: "category: обязательное поле"}
	}
	return nil
}
