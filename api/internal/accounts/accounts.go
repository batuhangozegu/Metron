// Package accounts manages user-owned trading accounts. Only paper accounts
// can be created until exchange credentials are implemented in step 8.
package accounts

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
)

type Account struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Type         string    `json:"type"`
	Market       string    `json:"market"`
	Name         string    `json:"name"`
	QuoteAsset   string    `json:"quote_asset"`
	PaperBalance *string   `json:"paper_balance"`
	CreatedAt    time.Time `json:"created_at"`
}

type Store interface {
	CreatePaper(context.Context, string, string, string, decimal.Decimal) (Account, error)
	List(context.Context, string) ([]Account, error)
}

type Service struct{ Store Store }

var balancePattern = regexp.MustCompile(`^[0-9]{1,8}(\.[0-9]{1,8})?$`)

func ValidateBalance(value string) (decimal.Decimal, error) {
	if !balancePattern.MatchString(value) {
		return decimal.Zero, errors.New("paper_balance must be a decimal string with at most 8 decimal places")
	}
	balance, err := decimal.NewFromString(value)
	if err != nil || !balance.IsPositive() || balance.GreaterThan(decimal.NewFromInt(10000000)) {
		return decimal.Zero, errors.New("paper_balance must be greater than 0 and at most 10000000")
	}
	return balance, nil
}

func (s Service) Create(ctx context.Context, userID, name, value, market string) (Account, error) {
	if market == "" {
		market = "spot"
	}
	if market != "spot" && market != "futures" {
		return Account{}, auth.ValidationError("market must be spot or futures")
	}
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 100 {
		return Account{}, auth.ValidationError("name must contain 1 to 100 characters")
	}
	balance, err := ValidateBalance(value)
	if err != nil {
		return Account{}, auth.ValidationError(err.Error())
	}
	return s.Store.CreatePaper(ctx, userID, name, market, balance)
}

func (s Service) Mount(r chi.Router) {
	r.Post("/v1/accounts", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name         string `json:"name"`
			PaperBalance string `json:"paper_balance"`
			Market       string `json:"market"`
		}
		if !httpjson.Decode(w, r, &body) {
			return
		}
		account, err := s.Create(r.Context(), auth.UserFromContext(r.Context()).ID, body.Name, body.PaperBalance, body.Market)
		var validation auth.ValidationError
		if errors.As(err, &validation) {
			httpjson.Error(w, 400, err.Error())
			return
		}
		if err != nil {
			httpjson.Error(w, 503, "accounts temporarily unavailable")
			return
		}
		httpjson.Write(w, 201, account)
	})
	r.Get("/v1/accounts", func(w http.ResponseWriter, r *http.Request) {
		accounts, err := s.Store.List(r.Context(), auth.UserFromContext(r.Context()).ID)
		if err != nil {
			httpjson.Error(w, 503, "accounts temporarily unavailable")
			return
		}
		if accounts == nil {
			accounts = []Account{}
		}
		httpjson.Write(w, 200, accounts)
	})
}
