package ozon

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"crmdemo/internal/httpx"
)

type Handler struct {
	repo *Repository
	log  *slog.Logger
	now  func() time.Time
}

func NewHandler(repo *Repository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log, now: time.Now}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ozon/reports", h.list)
	mux.HandleFunc("GET /api/ozon/reports/{day}", h.lines)
	mux.HandleFunc("POST /api/ozon/reports/today", h.refreshToday)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, total, err := h.repo.List(r.Context(), httpx.QueryInt(r, "limit", 30, 1, 100), httpx.QueryInt(r, "offset", 0, 0, 100_000))
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if items == nil {
		items = []Report{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (h *Handler) lines(w http.ResponseWriter, r *http.Request) {
	day, err := time.Parse(time.DateOnly, r.PathValue("day"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "day: ожидается YYYY-MM-DD")
		return
	}
	lines, err := h.repo.Lines(r.Context(), day)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if lines == nil {
		lines = []Line{}
	}
	httpx.JSON(w, http.StatusOK, lines)
}

// refreshToday — ручной запуск: сегодняшний отчёт скачается вне расписания.
func (h *Handler) refreshToday(w http.ResponseWriter, r *http.Request) {
	y, m, d := h.now().Date()
	err := Enqueue(r.Context(), h.repo.pool, time.Date(y, m, d, 0, 0, 0, 0, time.UTC), "manual")
	if errors.Is(err, ErrBusy) {
		httpx.Error(w, http.StatusConflict, "отчёт за сегодня уже загружается")
		return
	}
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}
