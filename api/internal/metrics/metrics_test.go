package metrics

import "testing"

func TestDRR(t *testing.T) {
	tests := []struct {
		name          string
		cost, revenue int64
		drr           *float64
	}{
		{"обычный случай", 50_000, 400_000, ptr(12.5)},
		{"округление", 1, 3, ptr(33.33)},
		{"расходы больше выручки", 30_000, 20_000, ptr(150)},
		{"нет выручки", 10_000, 0, nil},
		{"нет расходов (бартер)", 0, 5_000, ptr(0)},
		{"пусто", 0, 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DRR(tt.cost, tt.revenue)
			switch {
			case got == nil && tt.drr == nil:
			case got == nil || tt.drr == nil:
				t.Fatalf("got %v, want %v", got, tt.drr)
			case *got != *tt.drr:
				t.Fatalf("got %v, want %v", *got, *tt.drr)
			}
		})
	}
}
