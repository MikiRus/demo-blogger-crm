// Package report — загрузка отчёта маркетплейса (XLSX) с превью перед записью.
package report

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const MaxRows = 5000

// Row — строка отчёта после разбора. Err заполнен, если строку прочитать не удалось.
type Row struct {
	Line       int    `json:"line"`
	Day        string `json:"day"`
	PromoCode  string `json:"promoCode"`
	Source     string `json:"source"`
	Orders     int    `json:"orders"`
	RevenueRub int64  `json:"revenueRub"`
	Err        string `json:"error,omitempty"`
}

var headerAliases = map[string]string{
	"дата": "day", "date": "day", "день": "day",
	"промокод": "promo", "promo": "promo", "promo_code": "promo", "promocode": "promo",
	"источник": "source", "source": "source", "площадка": "source",
	"заказы": "orders", "orders": "orders", "кол-во заказов": "orders",
	"выручка": "revenue", "revenue": "revenue", "сумма": "revenue",
}

var sources = map[string]string{
	"wb": "wb", "wildberries": "wb", "вб": "wb",
	"ozon": "ozon", "озон": "ozon",
	"site": "site", "сайт": "site", "crm": "site",
}

var ErrBadFile = errors.New("не удалось прочитать XLSX")

// ParseXLSX читает первый лист. Колонки ищутся по заголовку, поэтому их порядок
// не важен; без колонки «Источник» все строки считаются отчётом WB.
func ParseXLSX(r io.Reader) ([]Row, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, ErrBadFile
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrBadFile
	}
	raw, err := f.GetRows(sheets[0], excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, ErrBadFile
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("в файле нет строк с данными")
	}

	cols := map[string]int{}
	for i, h := range raw[0] {
		if key, ok := headerAliases[strings.ToLower(strings.TrimSpace(h))]; ok {
			cols[key] = i
		}
	}
	for _, need := range []string{"day", "promo", "orders", "revenue"} {
		if _, ok := cols[need]; !ok {
			return nil, fmt.Errorf("нет обязательной колонки: %s", map[string]string{
				"day": "Дата", "promo": "Промокод", "orders": "Заказы", "revenue": "Выручка",
			}[need])
		}
	}
	if len(raw)-1 > MaxRows {
		return nil, fmt.Errorf("слишком много строк: максимум %d", MaxRows)
	}

	rows := make([]Row, 0, len(raw)-1)
	for i, rec := range raw[1:] {
		cell := func(key string) string {
			idx, ok := cols[key]
			if !ok || idx >= len(rec) {
				return ""
			}
			return strings.TrimSpace(rec[idx])
		}
		if isBlank(rec) {
			continue
		}
		rows = append(rows, parseRow(i+2, cell))
	}
	return rows, nil
}

func parseRow(line int, cell func(string) string) Row {
	row := Row{Line: line, PromoCode: strings.ToUpper(cell("promo")), Source: "wb"}

	if src := cell("source"); src != "" {
		s, ok := sources[strings.ToLower(src)]
		if !ok {
			row.Err = "неизвестный источник: " + src
			return row
		}
		row.Source = s
	}

	day, err := parseDay(cell("day"))
	if err != nil {
		row.Err = "некорректная дата: " + cell("day")
		return row
	}
	row.Day = day.Format(time.DateOnly)

	if row.PromoCode == "" {
		row.Err = "пустой промокод"
		return row
	}
	orders, err := strconv.Atoi(cell("orders"))
	if err != nil || orders < 0 {
		row.Err = "некорректное число заказов: " + cell("orders")
		return row
	}
	row.Orders = orders

	revenue, err := parseMoney(cell("revenue"))
	if err != nil || revenue < 0 {
		row.Err = "некорректная выручка: " + cell("revenue")
		return row
	}
	row.RevenueRub = revenue
	return row
}

func parseDay(s string) (time.Time, error) {
	for _, layout := range []string{time.DateOnly, "02.01.2006", "02.01.06"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	// Ячейка с форматом даты приходит как серийный номер Excel.
	if serial, err := strconv.ParseFloat(s, 64); err == nil && serial > 0 {
		return excelize.ExcelDateToTime(serial, false)
	}
	return time.Time{}, fmt.Errorf("bad date %q", s)
}

// parseMoney понимает «12 345,67», «12345.67» и «12 345 ₽», округляет до рубля.
func parseMoney(s string) (int64, error) {
	s = strings.NewReplacer(" ", "", " ", "", "₽", "", "руб.", "", "руб", "", ",", ".").Replace(s)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("bad money %q", s)
	}
	return int64(math.Round(v)), nil
}

func isBlank(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
