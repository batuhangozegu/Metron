package trading

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type wireError struct {
	ID string `json:"id"`
	EN string `json:"message_en"`
	TR string `json:"message_tr"`
}

func failure(w http.ResponseWriter, code int, id, en, tr string, check any) {
	value := map[string]any{"error": wireError{id, en, tr}}
	if check != nil {
		value["check"] = check
	}
	httpjson.Write(w, code, value)
}

// ErrorShape also covers errors emitted by the shared auth/JSON middleware.
func ErrorShape(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/v1/orders" || (strings.HasPrefix(path, "/v1/positions/") && strings.HasSuffix(path, "/close")) ||
			(strings.HasPrefix(path, "/v1/accounts/") && strings.HasSuffix(path, "/positions")) {
			w = contractWriter{w}
		}
		next.ServeHTTP(w, r)
	})
}

type contractWriter struct{ http.ResponseWriter }

func (w contractWriter) WriteError(code int, en, tr string) {
	id := "invalid_input"
	if code == 401 {
		id = "authentication_required"
	} else if code == 415 {
		id = "unsupported_media_type"
	} else if code >= 500 {
		id = "trading_unavailable"
	}
	if tr == "" {
		tr = "İstek işlenemedi."
	}
	failure(w.ResponseWriter, code, id, en, tr, nil)
}

func (p Position) MarshalJSON() ([]byte, error) {
	type plain Position
	if p.Status == "open" {
		return json.Marshal(struct {
			plain
			LastPrice     *string `json:"last_price"`
			UnrealizedPNL *string `json:"unrealized_pnl"`
		}{plain(p), p.LastPrice, p.UnrealizedPNL})
	}
	return json.Marshal(plain(p))
}

func (s Service) Mount(r chi.Router) {
	r.Post("/v1/orders", s.openHTTP)
	r.Get("/v1/accounts/{id}/positions", s.positionsHTTP)
	r.Post("/v1/positions/{id}/close", s.closeHTTP)
}

func checkJSON(e Execution) json.RawMessage {
	if e.Check == nil {
		return nil
	}
	encoded, _ := (protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}).Marshal(e.Check)
	return encoded
}

func respondError(w http.ResponseWriter, err error, e Execution) {
	code, id, en, tr := 503, "trading_unavailable", "Paper trading is temporarily unavailable.", "Kağıt işlemler geçici olarak kullanılamıyor."
	switch {
	case errors.Is(err, risk.ErrAccountNotFound):
		code, id, en, tr = 404, "account_not_found", "Account not found.", "Hesap bulunamadı."
	case errors.Is(err, ErrNotFound):
		code, id, en, tr = 404, "position_not_found", "Position not found.", "Pozisyon bulunamadı."
	case errors.Is(err, ErrClosed):
		code, id, en, tr = 409, "position_closed", "Position is already closed.", "Pozisyon zaten kapalı."
	case errors.Is(err, ErrBlocked):
		code, id, en, tr = 422, "risk_blocked", "Risk rules block this order.", "Risk kuralları bu emri engelliyor."
	case errors.Is(err, ErrQuantity):
		code, id, en, tr = 422, "quantity_exceeded", "Quantity exceeds the suggested position size.", "Miktar önerilen pozisyon büyüklüğünü aşıyor."
	case errors.Is(err, ErrBalance):
		code, id, en, tr = 422, "insufficient_balance", "Balance does not cover the position and fee.", "Bakiye pozisyon ve komisyonu karşılamıyor."
	case errors.Is(err, ErrInput):
		code, id, en, tr = 400, "invalid_input", "Invalid order input; use positive decimal strings and a USDT symbol.", "Geçersiz emir girdisi; pozitif ondalık dizeler ve USDT işlem çifti kullanın."
	case errors.Is(err, ErrPrice):
		id, en, tr = "price_unavailable", "Latest ticker price is unavailable.", "Son fiyat verisi kullanılamıyor."
	case status.Code(err) == codes.InvalidArgument || status.Code(err) == codes.OutOfRange:
		code, id, en, tr = 400, "invalid_input", "Invalid risk input, stop or leverage.", "Geçersiz risk girdisi, stop veya kaldıraç."
	case status.Code(err) == codes.NotFound:
		code, id, en, tr = 404, "symbol_not_found", "Symbol not found.", "İşlem çifti bulunamadı."
	}
	if e.Check != nil {
		failure(w, code, id, en, tr, checkJSON(e))
	} else {
		failure(w, code, id, en, tr, nil)
	}
}

func (s Service) openHTTP(w http.ResponseWriter, r *http.Request) {
	var body *Request
	if !httpjson.Decode(w, r, &body) {
		return
	}
	if body == nil {
		respondError(w, ErrInput, Execution{})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	e, err := s.Open(ctx, auth.UserFromContext(ctx).ID, *body)
	if err != nil {
		respondError(w, err, e)
		return
	}
	httpjson.Write(w, 201, map[string]any{"order": e.Order, "fill": e.Fill, "position": e.Position, "check": checkJSON(e)})
}

func (s Service) closeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	e, err := s.Close(ctx, auth.UserFromContext(ctx).ID, chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, err, e)
		return
	}
	httpjson.Write(w, 200, e)
}

func (s Service) positionsHTTP(w http.ResponseWriter, r *http.Request) {
	owner, id := auth.UserFromContext(r.Context()).ID, chi.URLParam(r, "id")
	if !risk.ValidID(id) {
		respondError(w, risk.ErrAccountNotFound, Execution{})
		return
	}
	state := r.URL.Query().Get("status")
	if state == "" {
		state = "open"
	}
	if state != "open" && state != "closed" {
		respondError(w, ErrInput, Execution{})
		return
	}
	if s.Store == nil {
		respondError(w, ErrUnavailable, Execution{})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	positions, err := s.Store.Positions(ctx, owner, id, state)
	if err != nil {
		respondError(w, err, Execution{})
		return
	}
	if positions == nil {
		positions = []Position{}
	}
	for i := range positions {
		p := &positions[i]
		if p.Status != "open" || s.Exchange == nil {
			continue
		}
		price, err := s.Exchange.Price(ctx, p.Symbol)
		if err != nil {
			continue
		}
		pnl, err := grossPNL(*p, price)
		if err != nil {
			continue
		}
		last, unrealized := price.String(), pnl.String()
		p.LastPrice, p.UnrealizedPNL = &last, &unrealized
	}
	httpjson.Write(w, 200, map[string]any{"positions": positions})
}
