// Package risk exposes the engine's preview behind the existing session policy.
// Risk decisions and sizing remain in Rust; Go supplies account aggregates.
package risk

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/batuhangozegu/metron/api/internal/auth"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
)

var ErrAccountNotFound = errors.New("account not found")

type Accounts interface {
	// RiskAccount must filter by BOTH owner and account, and only accept USDT accounts.
	RiskAccount(context.Context, string, string, string) (Account, error)
}

type Account struct {
	Equity              string
	Market              string
	Activity            *enginev1.Activity
	CurrentClassHolding string
	Profile             *Profile
	Snapshot            *ActivitySnapshot
}

type Engine interface {
	Check(context.Context, *enginev1.CheckRequest, ...grpc.CallOption) (*enginev1.CheckResponse, error)
}

type Service struct {
	Accounts Accounts
	Engine   Engine
	Store    Store
	Prices   Prices
}

type Profile struct {
	RiskPerTradePercent       string  `json:"risk_per_trade_percent"`
	TargetRiskReward          string  `json:"target_risk_reward"`
	TotalOpenRiskLimitPercent string  `json:"total_open_risk_limit_percent"`
	MaxOpenPositions          *uint32 `json:"max_open_positions"`
	MaxLosingTradesToday      *uint32 `json:"max_losing_trades_today"`
	DailyLossLimitPercent     string  `json:"daily_loss_limit_percent"`
	MaxStoppedTradesThisWeek  *uint32 `json:"max_stopped_trades_this_week"`
	WeeklyLossLimitPercent    string  `json:"weekly_loss_limit_percent"`
}

type Request struct {
	AccountID   string   `json:"account_id"`
	Symbol      string   `json:"symbol"`
	Side        string   `json:"side"`
	EntryPrice  string   `json:"entry_price"`
	ManualStop  *string  `json:"manual_stop"`
	RiskProfile *Profile `json:"risk_profile"`
	Leverage    *uint32  `json:"leverage"`
}

var (
	uuid         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	symbol       = regexp.MustCompile(`^[A-Z0-9]{1,26}USDT$`)
	plainDecimal = regexp.MustCompile(`^[0-9]{1,13}(\.[0-9]{1,8})?$`)
)

func validDecimal(value string, max int64) bool {
	if !plainDecimal.MatchString(value) {
		return false
	}
	d, err := decimal.NewFromString(value)
	return err == nil && d.IsPositive() && d.LessThanOrEqual(decimal.NewFromInt(max))
}

func ValidID(value string) bool       { return uuid.MatchString(value) }
func ValidQuantity(value string) bool { return validDecimal(value, 1000000000000) }

func (r Request) Valid() bool {
	if r.Leverage != nil && (*r.Leverage < 1 || *r.Leverage > 125) {
		return false
	}
	if !symbol.MatchString(r.Symbol) || (r.Side != "long" && r.Side != "short") || !validDecimal(r.EntryPrice, 1000000000000) {
		return false
	}
	if r.ManualStop != nil && !validDecimal(*r.ManualStop, 1000000000000) {
		return false
	}
	if r.RiskProfile != nil && !r.RiskProfile.Valid(false) {
		return false
	}
	return true
}

func failure(w http.ResponseWriter, code int, id, en, tr string) {
	httpjson.Write(w, code, map[string]string{"code": id, "error": en, "error_tr": tr})
}

func engineFailure(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument, codes.OutOfRange:
		messages := map[string][2]string{
			"invalid_leverage":       {"Spot permits leverage 1 or absent; futures requires integer leverage from 1 to 125.", "Spot için kaldıraç 1 olmalı veya verilmemelidir; vadeli işlemler için 1 ile 125 arasında tam sayı kaldıraç gereklidir."},
			"wrong_stop_side":        {"Stop must be below entry for a long and above entry for a short.", "Stop, uzun pozisyonda girişin altında, kısa pozisyonda girişin üstünde olmalıdır."},
			"zero_stop_distance":     {"Stop must differ from entry.", "Stop, giriş fiyatından farklı olmalıdır."},
			"stablecoin_unsupported": {"Stablecoins are cash and do not support ATR risk sizing.", "Sabit coinler nakit olarak değerlendirilir; ATR risk hesabı desteklenmez."},
			"non_positive":           {"Prices, equity, risk and the suggested stop must be positive.", "Fiyatlar, öz sermaye, risk ve önerilen stop pozitif olmalıdır."},
			"arithmetic_range":       {"Values exceed the exact calculation range.", "Değerler kesin hesaplama aralığını aşıyor."},
		}
		id := status.Convert(err).Message()
		if message, ok := messages[id]; ok {
			failure(w, 400, id, message[0], message[1])
		} else {
			failure(w, 400, "invalid_input", "Invalid risk check input.", "Risk kontrolü girdisi geçersiz.")
		}
	case codes.NotFound:
		failure(w, 404, "symbol_not_found", "Symbol not found.", "İşlem çifti bulunamadı.")
	default:
		failure(w, 503, "risk_unavailable", "Risk check unavailable; fresh market data and the engine are required.", "Risk kontrolü kullanılamıyor; güncel piyasa verisi ve hesaplama motoru gereklidir.")
	}
}

// Mount must be called in auth.HTTP.Mount's protected group.
func (s Service) Mount(r chi.Router) {
	r.Post("/v1/risk/check", s.check)
	r.Get("/v1/accounts/{id}/risk-profile", s.profileHTTP)
	r.Put("/v1/accounts/{id}/risk-profile", s.profileHTTP)
	r.Get("/v1/accounts/{id}/risk-checks", s.checksHTTP)
	r.Get("/v1/accounts/{id}/risk-status", s.statusHTTP)
}

func (s Service) check(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserFromContext(r.Context()).ID
	if userID == "" {
		failure(w, 401, "authentication_required", "Authentication required.", "Oturum açmanız gerekiyor.")
		return
	}
	var body *Request
	if !httpjson.Decode(w, r, &body) {
		return
	}
	if body == nil {
		failure(w, 400, "invalid_input", "A JSON object is required.", "Bir JSON nesnesi gereklidir.")
		return
	}
	if !uuid.MatchString(body.AccountID) {
		failure(w, 404, "account_not_found", "Account not found.", "Hesap bulunamadı.")
		return
	}
	if !body.Valid() {
		failure(w, 400, "invalid_input", "Use a USDT symbol, long/short side and positive plain decimal strings with up to 8 decimal places.", "USDT işlem çifti, long/short yönü ve en fazla 8 ondalık basamaklı pozitif ondalık dizeler kullanın.")
		return
	}
	if s.Accounts == nil || s.Engine == nil || s.Store == nil {
		engineFailure(w, status.Error(codes.Unavailable, ""))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	account, err := s.Accounts.RiskAccount(ctx, userID, body.AccountID, body.Symbol)
	if errors.Is(err, ErrAccountNotFound) {
		failure(w, 404, "account_not_found", "Account not found.", "Hesap bulunamadı.")
		return
	}
	if err != nil {
		engineFailure(w, err)
		return
	}
	result, err := s.Evaluate(ctx, *body, account)
	if err != nil {
		engineFailure(w, err)
		return
	}
	if result == nil {
		engineFailure(w, status.Error(codes.Unavailable, ""))
		return
	}
	// Preserve the contract's snake_case names and decimal strings, including
	// zero values, without exporting Go protobuf implementation details.
	encoded, err := checkResponse(result, ProfileSource(*body))
	if err != nil {
		engineFailure(w, err)
		return
	}
	if err := s.Store.RecordCheck(ctx, userID, body.AccountID, *body, result); err != nil {
		storeFailure(w, err)
		return
	}
	httpjson.Write(w, 200, encoded)
}

// Evaluate is shared by previews and orders. Orders supply a snapshot protected
// by the account lock; clients cannot supply holdings or activity.
func (s Service) Evaluate(ctx context.Context, body Request, account Account) (*enginev1.CheckResponse, error) {
	if !body.Valid() {
		return nil, status.Error(codes.InvalidArgument, "invalid_input")
	}
	if s.Engine == nil || (account.Market != "spot" && account.Market != "futures") || account.Activity == nil || account.CurrentClassHolding == "" {
		return nil, status.Error(codes.Unavailable, "risk activity unavailable")
	}
	if (account.Market == "spot" && body.Leverage != nil && *body.Leverage != 1) || (account.Market == "futures" && body.Leverage == nil) {
		return nil, status.Error(codes.InvalidArgument, "invalid_leverage")
	}
	request := &enginev1.CheckRequest{Symbol: body.Symbol, Side: body.Side, EntryPrice: body.EntryPrice, ManualStop: body.ManualStop,
		AccountEquity: account.Equity, Market: account.Market, Leverage: body.Leverage,
		CurrentClassHolding: account.CurrentClassHolding, Activity: account.Activity}
	p := ResolveProfile(account.Profile, body.RiskProfile)
	request.RiskProfile = &enginev1.RiskProfile{RiskPerTradePercent: p.RiskPerTradePercent, TargetRiskReward: p.TargetRiskReward,
		TotalOpenRiskLimitPercent: p.TotalOpenRiskLimitPercent, MaxOpenPositions: p.MaxOpenPositions, MaxLosingTradesToday: p.MaxLosingTradesToday,
		DailyLossLimitPercent: p.DailyLossLimitPercent, MaxStoppedTradesThisWeek: p.MaxStoppedTradesThisWeek, WeeklyLossLimitPercent: p.WeeklyLossLimitPercent}
	result, err := s.Engine.Check(ctx, request)
	if err == nil && result == nil {
		err = status.Error(codes.Unavailable, "empty risk response")
	}
	return result, err
}
