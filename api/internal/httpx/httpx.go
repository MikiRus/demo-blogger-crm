package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type errorBody struct {
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, errorBody{Error: msg})
}

// Decode читает JSON-тело не больше 1 МБ и отвергает лишние поля.
func Decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

// ValidationError — ошибка входных данных, отдаётся клиенту как 400.
type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

var ErrNotFound = errors.New("not found")

// Fail переводит доменную ошибку в HTTP-ответ, не раскрывая внутренности.
func Fail(w http.ResponseWriter, log *slog.Logger, err error) {
	var ve ValidationError
	switch {
	case errors.As(err, &ve):
		Error(w, http.StatusBadRequest, ve.Msg)
	case errors.Is(err, ErrNotFound):
		Error(w, http.StatusNotFound, "not found")
	default:
		log.Error("request failed", "err", err)
		Error(w, http.StatusInternalServerError, "internal error")
	}
}

func PathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, ValidationError{Msg: "invalid id"}
	}
	return id, nil
}

func QueryInt(r *http.Request, key string, def, min, max int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return def
	}
	return clamp(v, min, max)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// DateRange разбирает ?from=&to= (YYYY-MM-DD), по умолчанию — последние days дней.
func DateRange(r *http.Request, now time.Time, days int) (time.Time, time.Time, error) {
	to := now
	from := now.AddDate(0, 0, -days)
	var err error
	if v := r.URL.Query().Get("from"); v != "" {
		if from, err = time.Parse(time.DateOnly, v); err != nil {
			return from, to, ValidationError{Msg: "from: expected YYYY-MM-DD"}
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if to, err = time.Parse(time.DateOnly, v); err != nil {
			return from, to, ValidationError{Msg: "to: expected YYYY-MM-DD"}
		}
	}
	if to.Before(from) {
		return from, to, ValidationError{Msg: "to must not be before from"}
	}
	return from, to, nil
}
