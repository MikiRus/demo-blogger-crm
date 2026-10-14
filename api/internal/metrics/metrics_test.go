package metrics

import "testing"

func TestDRRAndROAS(t *testing.T) {
	tests := []struct {
		name          string
		cost, revenue int64
		drr, roas     *float64
	}{
		{"обычный случай", 50_000, 400_000, ptr(12.5), ptr(8)},
		{"округление", 1, 3, ptr(33.33), ptr(3)},
		{"нет выручки", 10_000, 0, nil, ptr(0)},
		{"нет расходов (бартер)", 0, 5_000, ptr(0), nil},
		{"пусто", 0, 0, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPtr(t, "drr", DRR(tt.cost, tt.revenue), tt.drr)
			assertPtr(t, "roas", ROAS(tt.cost, tt.revenue), tt.roas)
		})
	}
}

func assertPtr(t *testing.T, name string, got, want *float64) {
	t.Helper()
	switch {
	case got == nil && want == nil:
	case got == nil || want == nil:
		t.Fatalf("%s: got %v, want %v", name, got, want)
	case *got != *want:
		t.Fatalf("%s: got %v, want %v", name, *got, *want)
	}
}
