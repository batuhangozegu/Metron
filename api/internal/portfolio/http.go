package portfolio

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/go-chi/chi/v5"
)

func (s Service) Mount(r chi.Router) {
	r.Get("/v1/accounts/{id}/portfolio", s.portfolioHTTP)
	r.Get("/v1/accounts/{id}/journal", s.journalHTTP)
}

func failure(w http.ResponseWriter, err error) {
	if errors.Is(err, risk.ErrAccountNotFound) {
		httpjson.Write(w, 404, map[string]string{"error": "Account not found.", "code": "account_not_found", "error_tr": "Hesap bulunamadı."})
		return
	}
	httpjson.Write(w, 503, map[string]string{"error": "portfolio_unavailable"})
}

func (s Service) portfolioHTTP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !risk.ValidID(id) {
		failure(w, risk.ErrAccountNotFound)
		return
	}
	period := "30d"
	if values, present := r.URL.Query()["period"]; present {
		if len(values) != 1 {
			httpjson.Write(w, 400, map[string]string{"error": "invalid_period"})
			return
		}
		period = values[0]
	}
	if period != "30d" && period != "90d" && period != "all" {
		httpjson.Write(w, 400, map[string]string{"error": "invalid_period"})
		return
	}
	if s.Store == nil {
		failure(w, nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	value, err := s.Portfolio(ctx, auth.UserFromContext(ctx).ID, id, period)
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, 200, value)
}

func (s Service) journalHTTP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !risk.ValidID(id) {
		failure(w, risk.ErrAccountNotFound)
		return
	}
	limit := 50
	if values, present := r.URL.Query()["limit"]; present {
		var err error
		if len(values) != 1 {
			httpjson.Write(w, 400, map[string]string{"error": "invalid_limit"})
			return
		}
		limit, err = strconv.Atoi(values[0])
		if err != nil || limit < 1 || limit > 200 {
			httpjson.Write(w, 400, map[string]string{"error": "invalid_limit"})
			return
		}
	}
	var before *time.Time
	if values, present := r.URL.Query()["before"]; present {
		if len(values) != 1 {
			httpjson.Write(w, 400, map[string]string{"error": "invalid_before"})
			return
		}
		at, err := time.Parse(time.RFC3339Nano, values[0])
		if err != nil {
			httpjson.Write(w, 400, map[string]string{"error": "invalid_before"})
			return
		}
		at = at.UTC()
		before = &at
	}
	if s.Store == nil {
		failure(w, nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	value, err := s.Store.Journal(ctx, auth.UserFromContext(ctx).ID, id, limit, before)
	if err != nil {
		failure(w, err)
		return
	}
	httpjson.Write(w, 200, value)
}
