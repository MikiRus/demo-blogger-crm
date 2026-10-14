package campaign

import (
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
	mux.HandleFunc("GET /api/campaigns", h.list)
	mux.HandleFunc("GET /api/campaigns/{id}", h.get)
	mux.HandleFunc("POST /api/campaigns", h.create)
	mux.HandleFunc("POST /api/campaigns/{id}/integrations", h.addIntegration)
}

func (h *Handler) today() time.Time {
	y, m, d := h.now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" && !statuses[status] {
		httpx.Error(w, http.StatusBadRequest, "status: planned | active | finished")
		return
	}
	items, err := h.repo.List(r.Context(), status)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if items == nil {
		items = []ListItem{}
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.PathID(r)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	d, err := h.repo.Get(r.Context(), id)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if d.Integrations == nil {
		d.Integrations = []Integration{}
	}
	if d.Daily == nil {
		d.Daily = []DayPoint{}
	}
	httpx.JSON(w, http.StatusOK, d)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	start, end, err := in.Validate()
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	id, err := h.repo.Create(r.Context(), in, start, end, StatusFor(start, end, h.today()))
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *Handler) addIntegration(w http.ResponseWriter, r *http.Request) {
	campaignID, err := httpx.PathID(r)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	var in AddIntegrationInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	publish, err := in.Validate()
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	status := "planned"
	if !publish.After(h.today()) {
		status = "published"
	}
	id, promo, err := h.repo.AddIntegration(r.Context(), campaignID, in, publish, status)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": id, "promoCode": promo})
}
