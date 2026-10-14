package blogger

import (
	"log/slog"
	"net/http"
	"strings"

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
	mux.HandleFunc("GET /api/platforms", h.platforms)
	mux.HandleFunc("GET /api/bloggers", h.list)
	mux.HandleFunc("GET /api/bloggers/{id}", h.get)
	mux.HandleFunc("POST /api/bloggers", h.create)
}

func (h *Handler) platforms(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.Platforms(r.Context())
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilter{
		Platform: q.Get("platform"),
		Query:    strings.TrimSpace(q.Get("q")),
		Sort:     q.Get("sort"),
		Limit:    httpx.QueryInt(r, "limit", 20, 1, 100),
		Offset:   httpx.QueryInt(r, "offset", 0, 0, 1_000_000),
	}
	items, total, err := h.repo.List(r.Context(), f)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if items == nil {
		items = []ListItem{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
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
	if d.History == nil {
		d.History = []IntegrationRow{}
	}
	httpx.JSON(w, http.StatusOK, d)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := in.Validate(); err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	id, err := h.repo.Create(r.Context(), in)
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]int64{"id": id})
}
