package risk

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
	"github.com/go-chi/chi/v5"
)

func storeFailure(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrAccountNotFound) {
		failure(w, 404, "account_not_found", "Account not found.", "Hesap bulunamadı.")
		return
	}
	failure(w, 503, "risk_unavailable", "Risk settings and history are temporarily unavailable.", "Risk ayarları ve geçmişi geçici olarak kullanılamıyor.")
}

func (s Service) profileHTTP(w http.ResponseWriter, r *http.Request) {
	owner, id := auth.UserFromContext(r.Context()).ID, chi.URLParam(r, "id")
	if !ValidID(id) {
		storeFailure(w, ErrAccountNotFound)
		return
	}
	var body *Profile
	if r.Method == http.MethodPut {
		if !httpjson.Decode(w, r, &body) {
			return
		}
		if body == nil || !body.Valid(true) {
			failure(w, 400, "invalid_input", "All eight risk profile fields are required and must contain positive limits.", "Sekiz risk profili alanının tamamı gereklidir ve pozitif sınırlar içermelidir.")
			return
		}
	}
	if s.Store == nil {
		storeFailure(w, nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	var value SavedProfile
	var err error
	if body != nil {
		value, err = s.Store.SaveProfile(ctx, owner, id, *body)
	} else {
		value, err = s.Store.Profile(ctx, owner, id)
	}
	if err != nil {
		storeFailure(w, err)
		return
	}
	httpjson.Write(w, 200, value)
}

func (s Service) checksHTTP(w http.ResponseWriter, r *http.Request) {
	owner, id := auth.UserFromContext(r.Context()).ID, chi.URLParam(r, "id")
	if !ValidID(id) {
		storeFailure(w, ErrAccountNotFound)
		return
	}
	limit := 20
	if values, present := r.URL.Query()["limit"]; present {
		var err error
		if len(values) != 1 {
			failure(w, 400, "invalid_input", "Limit must be an integer from 1 to 100.", "Limit 1 ile 100 arasında bir tam sayı olmalıdır.")
			return
		}
		limit, err = strconv.Atoi(values[0])
		if err != nil || limit < 1 || limit > 100 {
			failure(w, 400, "invalid_input", "Limit must be an integer from 1 to 100.", "Limit 1 ile 100 arasında bir tam sayı olmalıdır.")
			return
		}
	}
	if s.Store == nil {
		storeFailure(w, nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	checks, err := s.Store.Checks(ctx, owner, id, limit)
	if err != nil {
		storeFailure(w, err)
		return
	}
	if checks == nil {
		checks = []CheckRecord{}
	}
	httpjson.Write(w, 200, map[string]any{"checks": checks})
}
