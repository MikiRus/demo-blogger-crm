package report

// Статусы строки в превью.
const (
	StatusNew          = "new"           // будет записана
	StatusDuplicate    = "duplicate"     // день уже загружен — не перезаписываем
	StatusUnknownPromo = "unknown_promo" // промокод не привязан ни к одной интеграции
	StatusInvalid      = "invalid"       // строку не удалось разобрать
)

type Key struct {
	IntegrationID int64
	Day           string
	Source        string
}

type PreviewRow struct {
	Row
	Status        string `json:"status"`
	IntegrationID int64  `json:"integrationId,omitempty"`
	Blogger       string `json:"blogger,omitempty"`
}

type Preview struct {
	Rows    []PreviewRow   `json:"rows"`
	Summary map[string]int `json:"summary"`
}

type PromoInfo struct {
	IntegrationID int64
	Blogger       string
}

// Classify раскладывает строки по статусам. Повтор той же пары
// (интеграция, день, источник) внутри файла тоже считается дублем —
// иначе вторая строка молча потерялась бы при записи.
func Classify(rows []Row, promos map[string]PromoInfo, existing map[Key]bool) Preview {
	p := Preview{
		Rows: make([]PreviewRow, 0, len(rows)),
		Summary: map[string]int{
			StatusNew: 0, StatusDuplicate: 0, StatusUnknownPromo: 0, StatusInvalid: 0,
		},
	}
	seen := map[Key]bool{}
	for _, r := range rows {
		pr := PreviewRow{Row: r}
		info, known := promos[r.PromoCode]
		switch {
		case r.Err != "":
			pr.Status = StatusInvalid
		case !known:
			pr.Status = StatusUnknownPromo
		default:
			pr.IntegrationID, pr.Blogger = info.IntegrationID, info.Blogger
			k := Key{IntegrationID: info.IntegrationID, Day: r.Day, Source: r.Source}
			if existing[k] || seen[k] {
				pr.Status = StatusDuplicate
			} else {
				pr.Status = StatusNew
				seen[k] = true
			}
		}
		p.Summary[pr.Status]++
		p.Rows = append(p.Rows, pr)
	}
	return p
}

// PromoCodes — уникальные промокоды разобранных строк, для одного запроса в БД.
func PromoCodes(rows []Row) []string {
	set := map[string]bool{}
	var out []string
	for _, r := range rows {
		if r.Err == "" && r.PromoCode != "" && !set[r.PromoCode] {
			set[r.PromoCode] = true
			out = append(out, r.PromoCode)
		}
	}
	return out
}
