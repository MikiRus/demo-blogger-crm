// Package metrics — показатели окупаемости рекламы.
package metrics

import "math"

// DRR — доля рекламных расходов в выручке, %. Без выручки не определена.
func DRR(costRub, revenueRub int64) *float64 {
	if revenueRub <= 0 {
		return nil
	}
	return ptr(round2(float64(costRub) / float64(revenueRub) * 100))
}

// ROAS — выручка на рубль рекламных расходов. Без расходов не определён.
func ROAS(costRub, revenueRub int64) *float64 {
	if costRub <= 0 {
		return nil
	}
	return ptr(round2(float64(revenueRub) / float64(costRub)))
}

// Totals — общие поля для любой агрегированной строки отчёта.
type Totals struct {
	CostRub    int64    `json:"costRub"`
	RevenueRub int64    `json:"revenueRub"`
	Orders     int64    `json:"orders"`
	DRR        *float64 `json:"drr"`
	ROAS       *float64 `json:"roas"`
}

func NewTotals(cost, revenue, orders int64) Totals {
	return Totals{
		CostRub:    cost,
		RevenueRub: revenue,
		Orders:     orders,
		DRR:        DRR(cost, revenue),
		ROAS:       ROAS(cost, revenue),
	}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func ptr(v float64) *float64 { return &v }
