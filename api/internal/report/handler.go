package report

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"crmdemo/internal/httpx"
)

const maxUpload = 5 << 20

type Handler struct {
	repo *Repository
	log  *slog.Logger
}

func NewHandler(repo *Repository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reports/sample.xlsx", h.sample)
	mux.HandleFunc("POST /api/reports/preview", h.preview)
	mux.HandleFunc("POST /api/reports/apply", h.apply)
}

func (h *Handler) sample(w http.ResponseWriter, r *http.Request) {
	src, err := h.repo.SampleIntegrations(r.Context(), 5)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	data, err := WriteXLSX(BuildSample(src, today, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 7))))
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="wb-report-demo.xlsx"`)
	_, _ = w.Write(data)
}

// classify — общий шаг превью и записи: сервер не доверяет статусам
// из браузера и при apply пересчитывает их заново.
func (h *Handler) classify(r *http.Request, rows []Row) (Preview, error) {
	promos, err := h.repo.Promos(r.Context(), PromoCodes(rows))
	if err != nil {
		return Preview{}, err
	}
	ids := make([]int64, 0, len(promos))
	for _, p := range promos {
		ids = append(ids, p.IntegrationID)
	}
	existing, err := h.repo.Existing(r.Context(), ids)
	if err != nil {
		return Preview{}, err
	}
	return Classify(rows, promos, existing), nil
}

func (h *Handler) preview(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	file, _, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "ожидается файл в поле file (до 5 МБ)")
		return
	}
	defer file.Close()

	rows, err := ParseXLSX(file)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	p, err := h.classify(r, rows)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	h.log.Info("report preview", "rows", len(rows), "new", p.Summary[StatusNew], "duplicate", p.Summary[StatusDuplicate])
	httpx.JSON(w, http.StatusOK, p)
}

type applyRequest struct {
	Rows []Row `json:"rows"`
}

func (h *Handler) apply(w http.ResponseWriter, r *http.Request) {
	var req applyRequest
	if err := httpx.Decode(w, r, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Rows) == 0 || len(req.Rows) > MaxRows {
		httpx.Error(w, http.StatusBadRequest, "rows: от 1 до 5000 строк")
		return
	}
	for i := range req.Rows {
		if err := revalidate(&req.Rows[i]); err != nil {
			httpx.Fail(w, h.log, err)
			return
		}
	}
	p, err := h.classify(r, req.Rows)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	var toInsert []PreviewRow
	for _, row := range p.Rows {
		if row.Status == StatusNew {
			toInsert = append(toInsert, row)
		}
	}
	inserted, err := h.repo.Insert(r.Context(), toInsert)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	h.log.Info("report applied", "inserted", inserted, "skipped", len(req.Rows)-inserted)
	httpx.JSON(w, http.StatusOK, map[string]int{"inserted": inserted, "skipped": len(req.Rows) - inserted})
}

// revalidate проверяет строку, пришедшую из браузера, прежде чем писать её в БД.
func revalidate(r *Row) error {
	if r.Err != "" {
		return nil // такие строки Classify пометит invalid и пропустит
	}
	if _, err := time.Parse(time.DateOnly, r.Day); err != nil {
		return httpx.ValidationError{Msg: "строка " + strconv.Itoa(r.Line) + ": некорректная дата"}
	}
	if r.Source != "wb" && r.Source != "ozon" && r.Source != "site" {
		return httpx.ValidationError{Msg: "строка " + strconv.Itoa(r.Line) + ": некорректный источник"}
	}
	if r.Orders < 0 || r.RevenueRub < 0 {
		return httpx.ValidationError{Msg: "строка " + strconv.Itoa(r.Line) + ": отрицательные значения"}
	}
	return nil
}
