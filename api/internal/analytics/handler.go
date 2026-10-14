package analytics

import (
	"log/slog"
	"net/http"
	"time"

	"crmdemo/internal/httpx"
)

type Handler struct {
	repo *Repository
	log  *slog.Logger
}

func NewHandler(repo *Repository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/analytics/summary", h.summary)
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	from, to, err := httpx.DateRange(r, time.Now(), 180)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	s, err := h.repo.Summary(r.Context(), from, to)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if s.TopBloggers == nil {
		s.TopBloggers = []TopBlogger{}
	}
	if s.Heatmap == nil {
		s.Heatmap = []HeatCell{}
	}
	if s.BySource == nil {
		s.BySource = []SourceRow{}
	}
	httpx.JSON(w, http.StatusOK, s)
}
