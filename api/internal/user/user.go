// Package user — сотрудники, у которых есть доступ к CRM, и их роли.
package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"crmdemo/internal/httpx"
)

var roles = map[string]bool{"admin": true, "marketer": true, "manager": true, "analyst": true, "finance": true}

type User struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
}

type InviteInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (in *InviteInput) Validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	switch {
	case in.Name == "" || len(in.Name) > 120:
		return httpx.ValidationError{Msg: "name: от 1 до 120 символов"}
	case !validEmail(in.Email):
		return httpx.ValidationError{Msg: "email: некорректный адрес"}
	case !roles[in.Role]:
		return httpx.ValidationError{Msg: "role: неизвестная роль"}
	}
	return nil
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s
}

type Handler struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewHandler(pool *pgxpool.Pool, log *slog.Logger) *Handler {
	return &Handler{pool: pool, log: log}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/users", h.list)
	mux.HandleFunc("POST /api/users", h.invite)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.all(r.Context())
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	if items == nil {
		items = []User{}
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) all(ctx context.Context) ([]User, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT id, name, email, role, status, last_seen_at FROM users
		ORDER BY status = 'blocked', last_seen_at DESC NULLS LAST, id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (User, error) {
		var u User
		err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.LastSeenAt)
		return u, err
	})
}

// invite заводит сотрудника в статусе invited; письмо в демо не отправляется.
func (h *Handler) invite(w http.ResponseWriter, r *http.Request) {
	var in InviteInput
	if err := httpx.Decode(w, r, &in); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := in.Validate(); err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	var id int64
	err := h.pool.QueryRow(r.Context(), `
		INSERT INTO users (name, email, role, status) VALUES ($1, $2, $3, 'invited') RETURNING id`,
		in.Name, in.Email, in.Role).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		httpx.Error(w, http.StatusBadRequest, "email: такой сотрудник уже есть")
		return
	}
	if err != nil {
		httpx.Fail(w, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]int64{"id": id})
}
