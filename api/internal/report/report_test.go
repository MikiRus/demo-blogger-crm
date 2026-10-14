package report

import (
	"bytes"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

func xlsx(t *testing.T, rows [][]any) *bytes.Reader {
	t.Helper()
	data, err := WriteXLSX(rows)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(data)
}

func TestParseXLSX(t *testing.T) {
	rows, err := ParseXLSX(xlsx(t, [][]any{
		{"Выручка", "Промокод", "Дата", "Заказы"}, // порядок колонок не важен, источника нет → wb
		{"12 345,60 ₽", "c01-1000", "14.10.2026", 3},
		{"", "", "", ""}, // пустая строка пропускается
		{"1000", "C01-1001", "2026-10-13", "много"},
		{"500", "", "13.10.2026", 1},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("want 3 rows, got %d: %+v", len(rows), rows)
	}
	first := rows[0]
	if first.Err != "" || first.Day != "2026-10-14" || first.PromoCode != "C01-1000" ||
		first.Source != "wb" || first.Orders != 3 || first.RevenueRub != 12346 {
		t.Errorf("first row parsed wrong: %+v", first)
	}
	if !strings.Contains(rows[1].Err, "заказов") {
		t.Errorf("want orders error, got %+v", rows[1])
	}
	if rows[1].Line != 4 {
		t.Errorf("line numbers must match the sheet, got %d", rows[1].Line)
	}
	if rows[2].Err != "пустой промокод" {
		t.Errorf("want empty promo error, got %+v", rows[2])
	}
}

func TestParseXLSXMissingColumn(t *testing.T) {
	_, err := ParseXLSX(xlsx(t, [][]any{{"Дата", "Промокод", "Заказы"}, {"14.10.2026", "X", 1}}))
	if err == nil || !strings.Contains(err.Error(), "Выручка") {
		t.Fatalf("want missing column error, got %v", err)
	}
}

func TestParseXLSXGarbage(t *testing.T) {
	if _, err := ParseXLSX(strings.NewReader("not an xlsx")); err != ErrBadFile {
		t.Fatalf("want ErrBadFile, got %v", err)
	}
}

func TestClassify(t *testing.T) {
	promos := map[string]PromoInfo{"A": {IntegrationID: 1, Blogger: "Анна"}, "B": {IntegrationID: 2}}
	existing := map[Key]bool{{IntegrationID: 2, Day: "2026-10-10", Source: "wb"}: true}
	rows := []Row{
		{Line: 2, Day: "2026-10-11", PromoCode: "A", Source: "wb"},
		{Line: 3, Day: "2026-10-11", PromoCode: "A", Source: "wb"},   // повтор внутри файла
		{Line: 4, Day: "2026-10-11", PromoCode: "A", Source: "ozon"}, // другой источник — новая
		{Line: 5, Day: "2026-10-10", PromoCode: "B", Source: "wb"},   // уже в базе
		{Line: 6, Day: "2026-10-10", PromoCode: "ZZZ", Source: "wb"},
		{Line: 7, Err: "некорректная дата"},
	}
	p := Classify(rows, promos, existing)

	want := []string{StatusNew, StatusDuplicate, StatusNew, StatusDuplicate, StatusUnknownPromo, StatusInvalid}
	for i, w := range want {
		if p.Rows[i].Status != w {
			t.Errorf("row %d: got %s, want %s", rows[i].Line, p.Rows[i].Status, w)
		}
	}
	if p.Rows[0].Blogger != "Анна" || p.Rows[0].IntegrationID != 1 {
		t.Errorf("match info missing: %+v", p.Rows[0])
	}
	if p.Summary[StatusNew] != 2 || p.Summary[StatusDuplicate] != 2 {
		t.Errorf("summary: %+v", p.Summary)
	}
}

func TestBuildSampleRoundTrip(t *testing.T) {
	today := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	src := []sampleSource{
		{Promo: "C10-1100", Reach: 90_000, Publish: today.AddDate(0, 0, -8)},
		{Promo: "C10-1101", Reach: 30_000, Publish: today.AddDate(0, 0, -1)},
	}
	rows, err := ParseXLSX(xlsx(t, BuildSample(src, today, rand.New(rand.NewPCG(1, 2)))))
	if err != nil {
		t.Fatal(err)
	}
	// 3 новых дня у первой, 1 у второй, дубль, неизвестный промокод, битая строка.
	if len(rows) != 7 {
		t.Fatalf("want 7 rows, got %d: %+v", len(rows), rows)
	}
	invalid := 0
	for _, r := range rows {
		if r.Err != "" {
			invalid++
		}
	}
	if invalid != 1 {
		t.Errorf("want exactly one broken row, got %d", invalid)
	}
}
