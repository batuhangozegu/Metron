// Package alerts manages user-owned price alerts and their background checker.
package alerts

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
	"github.com/batuhangozegu/metron/api/internal/market"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type Alert struct {
	ID             string     `json:"id"`
	Symbol         string     `json:"symbol"`
	Direction      string     `json:"direction"`
	Price          string     `json:"price"`
	Note           *string    `json:"note"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	TriggeredAt    *time.Time `json:"triggered_at"`
	TriggeredPrice *string    `json:"triggered_price"`
}

type Request struct {
	Symbol    string  `json:"symbol"`
	Direction string  `json:"direction"`
	Price     string  `json:"price"`
	Note      *string `json:"note"`
}

var (
	ErrLimit     = errors.New("active alert limit reached")
	ErrNotFound  = errors.New("alert not found")
	pricePattern = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,8})?$`)
	idPattern    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

type Store interface {
	Create(context.Context, string, Request) (Alert, error)
	List(context.Context, string, *time.Time) ([]Alert, error)
	Cancel(context.Context, string, string) error
	ActiveSymbols(context.Context) ([]string, error)
	Trigger(context.Context, string, decimal.Decimal) ([]Alert, error)
}

type Service struct {
	Store   Store
	Symbols market.SymbolSource
}

func (s Service) Create(ctx context.Context, userID string, r Request) (Alert, error) {
	if r.Direction != "above" && r.Direction != "below" {
		return Alert{}, auth.ValidationError("direction must be above or below / Yön above veya below olmalıdır")
	}
	price, err := decimal.NewFromString(r.Price)
	if !pricePattern.MatchString(r.Price) || err != nil || !price.IsPositive() {
		return Alert{}, auth.ValidationError("price must be positive with at most 8 decimal places / Fiyat pozitif ve en fazla 8 ondalık basamaklı olmalıdır")
	}
	if r.Note != nil && (!utf8.ValidString(*r.Note) || utf8.RuneCountInString(*r.Note) > 200) {
		return Alert{}, auth.ValidationError("note must contain at most 200 characters / Not en fazla 200 karakter olmalıdır")
	}
	if s.Symbols == nil || s.Store == nil {
		return Alert{}, errors.New("alerts unavailable")
	}
	symbols, err := s.Symbols.Symbols(ctx)
	if err != nil {
		return Alert{}, err
	}
	for _, symbol := range symbols {
		if symbol.Symbol == r.Symbol {
			r.Price = price.String()
			return s.Store.Create(ctx, userID, r)
		}
	}
	return Alert{}, auth.ValidationError("unknown symbol / Bilinmeyen sembol")
}

func (s Service) Mount(r chi.Router) {
	r.Post("/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		var body Request
		if !httpjson.Decode(w, r, &body) {
			return
		}
		alert, err := s.Create(r.Context(), auth.UserFromContext(r.Context()).ID, body)
		if err != nil {
			writeError(w, err)
			return
		}
		httpjson.Write(w, 201, alert)
	})
	r.Get("/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		var since *time.Time
		if values, ok := r.URL.Query()["triggered_since"]; ok {
			if len(values) != 1 {
				writeError(w, auth.ValidationError("invalid triggered_since / Geçersiz triggered_since"))
				return
			}
			stamp, err := time.Parse(time.RFC3339Nano, values[0])
			if err != nil {
				writeError(w, auth.ValidationError("triggered_since must be RFC3339 / triggered_since RFC3339 biçiminde olmalıdır"))
				return
			}
			since = &stamp
		}
		if s.Store == nil {
			writeError(w, errors.New("unavailable"))
			return
		}
		list, err := s.Store.List(r.Context(), auth.UserFromContext(r.Context()).ID, since)
		if err != nil {
			writeError(w, err)
			return
		}
		if list == nil {
			list = []Alert{}
		}
		httpjson.Write(w, 200, list)
	})
	r.Delete("/v1/alerts/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !idPattern.MatchString(id) {
			writeError(w, ErrNotFound)
			return
		}
		if s.Store == nil {
			writeError(w, errors.New("unavailable"))
			return
		}
		if err := s.Store.Cancel(r.Context(), auth.UserFromContext(r.Context()).ID, id); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(204)
	})
}

func writeError(w http.ResponseWriter, err error) {
	var validation auth.ValidationError
	status, message := 503, "alerts temporarily unavailable / Alarmlar geçici olarak kullanılamıyor"
	switch {
	case errors.As(err, &validation):
		status, message = 400, err.Error()
	case errors.Is(err, ErrLimit):
		status, message = 409, "maximum 50 active alerts / En fazla 50 aktif alarm olabilir"
	case errors.Is(err, ErrNotFound):
		status, message = 404, "alert not found / Alarm bulunamadı"
	}
	// Keep the shared httpjson.Error wire shape, including its bilingual fields.
	en, tr, _ := strings.Cut(message, " / ")
	value := map[string]string{"error": en, "error_tr": tr}
	if status == 400 {
		value["code"] = "invalid_input"
	}
	if status >= 500 {
		value["code"] = "unavailable"
	}
	httpjson.Write(w, status, value)
}
