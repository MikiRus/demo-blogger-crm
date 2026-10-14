package campaign

import (
	"regexp"
	"testing"
	"time"
)

func TestStatusFor(t *testing.T) {
	d := func(s string) time.Time { v, _ := time.Parse(time.DateOnly, s); return v }
	start, end := d("2026-05-01"), d("2026-05-31")
	cases := map[string]string{
		"2026-04-30": "planned",
		"2026-05-01": "active",
		"2026-05-31": "active",
		"2026-06-01": "finished",
	}
	for today, want := range cases {
		if got := StatusFor(start, end, d(today)); got != want {
			t.Errorf("%s: got %s, want %s", today, got, want)
		}
	}
}

func TestNewPromoCode(t *testing.T) {
	re := regexp.MustCompile(`^[A-Z0-9]{1,8}-[A-HJ-NP-Z2-9]{4}$`)
	for _, handle := range []string{"kitchen_lab", "Очень.Длинный_ник2026", "", "ab"} {
		code := NewPromoCode(handle)
		if !re.MatchString(code) {
			t.Errorf("%q → %q: unexpected format", handle, code)
		}
	}
	if got := NewPromoCode("kitchen_lab")[:9]; got != "KITCHENL-" {
		t.Errorf("prefix: got %q", got)
	}
	if got := NewPromoCode("Очень")[:6]; got != "PROMO-" {
		t.Errorf("non-latin fallback: got %q", got)
	}
}

func TestCreateInputValidate(t *testing.T) {
	in := CreateInput{Title: "  Осень  ", Product: "Блендер", StartDate: "2026-09-01", EndDate: "2026-08-01"}
	if _, _, err := in.Validate(); err == nil {
		t.Fatal("expected error for end before start")
	}
	in.EndDate = "2026-09-30"
	if _, _, err := in.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.Title != "Осень" {
		t.Errorf("title not trimmed: %q", in.Title)
	}
}
