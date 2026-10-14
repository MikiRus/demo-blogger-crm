package report

import (
	"bytes"
	"math/rand/v2"
	"time"

	"github.com/xuri/excelize/v2"

	"crmdemo/internal/seed"
)

// BuildSample собирает строки демо-отчёта WB за дни после уже загруженных
// (seed.WBLoadedThrough): несколько новых дней,
// один уже загруженный день, неизвестный промокод и битую строку —
// чтобы в превью были видны все статусы.
func BuildSample(src []sampleSource, today time.Time, rnd *rand.Rand) [][]any {
	out := [][]any{{"Дата", "Промокод", "Источник", "Заказы", "Выручка"}}
	dupAdded := false
	for _, s := range src {
		for back := seed.WBLoadedThrough - 1; back >= 1; back-- {
			day := today.AddDate(0, 0, -back)
			if day.Before(s.Publish) {
				continue
			}
			orders := rnd.IntN(max(2, s.Reach/15_000)) + 1
			revenue := orders * (6_000 + rnd.IntN(14_000))
			out = append(out, []any{day.Format("02.01.2006"), s.Promo, "WB", orders, revenue})
		}
		dupDay := today.AddDate(0, 0, -seed.WBLoadedThrough)
		if !dupAdded && !dupDay.Before(s.Publish) {
			out = append(out, []any{dupDay.Format("02.01.2006"), s.Promo, "WB", 3, 27_480})
			dupAdded = true
		}
	}
	yesterday := today.AddDate(0, 0, -1).Format("02.01.2006")
	out = append(out,
		[]any{yesterday, "SUMMER-2024", "WB", 2, 15_980},
		[]any{yesterday, "BROKEN-ROW", "WB", "много", "12 000"},
	)
	return out
}

func WriteXLSX(rows [][]any) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Отчёт WB"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			return nil, err
		}
	}
	_ = f.SetColWidth(sheet, "A", "E", 16)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
