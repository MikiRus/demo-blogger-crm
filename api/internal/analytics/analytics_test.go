package analytics

import "testing"

func TestBuildMonthly(t *testing.T) {
	months := map[string]*monthAgg{
		"2026-09": {cost: 100_000, revenue: 900_000, orders: 120},
		"2026-07": {cost: 50_000, revenue: 0, orders: 0},
		"2026-08": {cost: 0, revenue: 300_000, orders: 40},
	}
	out, total := buildMonthly(months)

	if len(out) != 3 || out[0].Month != "2026-07" || out[2].Month != "2026-09" {
		t.Fatalf("months not sorted: %+v", out)
	}
	if out[0].DRR != nil {
		t.Errorf("июль без выручки: DRR должен быть nil")
	}
	if out[1].ROAS != nil {
		t.Errorf("август без расходов: ROAS должен быть nil")
	}
	if total.CostRub != 150_000 || total.RevenueRub != 1_200_000 || total.Orders != 160 {
		t.Fatalf("total: %+v", total)
	}
	if *total.DRR != 12.5 || *total.ROAS != 8 {
		t.Errorf("total ratios: drr=%v roas=%v", *total.DRR, *total.ROAS)
	}
}
