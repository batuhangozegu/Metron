package accounts

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/binance"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Exchange is the small signed-exchange boundary used by live trading and fakes.
type Exchange interface {
	VerifyKeys(context.Context) error
	Balance(context.Context, string) (binance.Balance, error)
	Price(context.Context, string) (decimal.Decimal, error)
	Filters(context.Context, string) (binance.Filters, error)
	SetLeverage(context.Context, string, uint32) error
	MarketOrder(context.Context, binance.OrderRequest) (binance.OrderResult, error)
	StopOrder(context.Context, binance.StopRequest) (binance.OrderResult, error)
	CancelOrder(context.Context, string, int64) error
	QueryOrder(context.Context, string, int64) (binance.OrderResult, error)
	QueryOrderByClientID(context.Context, string, string) (binance.OrderResult, error)
}

type LiveRequest struct {
	Name        string `json:"name"`
	Market      string `json:"market"`
	Environment string `json:"environment"`
	APIKey      string `json:"api_key"`
	APISecret   string `json:"api_secret"`
}

type Credentials struct {
	Environment, Market, KeyHint string
	Key, Secret                  []byte
}
type LiveStore interface {
	CreateLive(context.Context, string, LiveRequest, Credentials, decimal.Decimal) (Account, error)
	Credentials(context.Context, string) (Credentials, error)
	DeleteLive(context.Context, string, string) error
}

var ErrDeleteConflict = errors.New("only live accounts without open positions can be deleted")
var ErrAccountNotFound = errors.New("account not found")
var ErrKeysRejected = errors.New("Binance rejected the API keys")

type cachedClient struct {
	exchange    Exchange
	environment string
}
type LiveService struct {
	Store   LiveStore
	Box     *cryptobox.Box
	Factory func(binance.Config) (Exchange, error)
	mu      sync.Mutex
	clients map[string]cachedClient
}

func (s *LiveService) client(c Credentials, key, secret string) (Exchange, error) {
	cfg := binance.Config{Env: binance.Env(c.Environment), Market: binance.Market(c.Market), Credentials: binance.Credentials{APIKey: key, Secret: secret},
		BaseURL: os.Getenv("BINANCE_" + strings.ToUpper(c.Market) + "_" + strings.ToUpper(c.Environment) + "_URL")}
	if s.Factory != nil {
		return s.Factory(cfg)
	}
	return binance.New(cfg)
}

// Client caches decrypted credentials per account. Callers authorize and lock
// the account before using this internal method.
func (s *LiveService) Client(ctx context.Context, id string) (Exchange, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.clients[id]; ok {
		return c.exchange, c.environment, nil
	}
	if s.Store == nil || s.Box == nil {
		return nil, "", binance.ErrUnavailable
	}
	c, err := s.Store.Credentials(ctx, id)
	if err != nil {
		return nil, "", err
	}
	key, err := s.Box.Open(c.Key, "binance-api-key")
	if err != nil {
		return nil, "", err
	}
	secret, err := s.Box.Open(c.Secret, "binance-api-secret")
	if err != nil {
		return nil, "", err
	}
	exchange, err := s.client(c, string(key), string(secret))
	if err != nil {
		return nil, "", err
	}
	if s.clients == nil {
		s.clients = map[string]cachedClient{}
	}
	s.clients[id] = cachedClient{exchange, c.Environment}
	return exchange, c.Environment, nil
}

func (s *LiveService) Create(ctx context.Context, owner string, r LiveRequest) (Account, error) {
	r.Name = strings.TrimSpace(r.Name)
	if !utf8.ValidString(r.Name) || utf8.RuneCountInString(r.Name) < 1 || utf8.RuneCountInString(r.Name) > 100 ||
		(r.Market != "spot" && r.Market != "futures") || (r.Environment != "testnet" && r.Environment != "mainnet") ||
		len(r.APIKey) < 4 || len(r.APIKey) > 512 || len(r.APISecret) < 1 || len(r.APISecret) > 512 || strings.ContainsAny(r.APIKey+r.APISecret, " \r\n\t") {
		return Account{}, auth.ValidationError("Invalid live account name, market, environment or keys. / Canlı hesap adı, piyasa, ortam veya anahtarlar geçersiz.")
	}
	if s.Store == nil || s.Box == nil {
		return Account{}, binance.ErrUnavailable
	}
	c := Credentials{Environment: r.Environment, Market: r.Market, KeyHint: "•••" + r.APIKey[len(r.APIKey)-4:]}
	exchange, err := s.client(c, r.APIKey, r.APISecret)
	if err != nil {
		return Account{}, err
	}
	if err = exchange.VerifyKeys(ctx); err != nil {
		return Account{}, ErrKeysRejected
	}
	balance, err := exchange.Balance(ctx, "USDT")
	if err != nil {
		return Account{}, err
	}
	c.Key, err = s.Box.Seal([]byte(r.APIKey), "binance-api-key")
	if err != nil {
		return Account{}, err
	}
	c.Secret, err = s.Box.Seal([]byte(r.APISecret), "binance-api-secret")
	if err != nil {
		return Account{}, err
	}
	return s.Store.CreateLive(ctx, owner, r, c, balance.Total)
}

func (s *LiveService) Delete(ctx context.Context, owner, id string) error {
	if s.Store == nil {
		return binance.ErrUnavailable
	}
	if err := s.Store.DeleteLive(ctx, owner, id); err != nil {
		return err
	}
	// Invalidate after the database lock is released, avoiding a lock inversion
	// with trades which lock the account before obtaining a cached client.
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, id)
	return nil
}

func (s Service) mountLive(r chi.Router) {
	r.Post("/v1/accounts/live", func(w http.ResponseWriter, r *http.Request) {
		var body LiveRequest
		if !httpjson.Decode(w, r, &body) {
			return
		}
		if s.Live == nil {
			httpjson.Error(w, 503, "accounts temporarily unavailable")
			return
		}
		a, err := s.Live.Create(r.Context(), auth.UserFromContext(r.Context()).ID, body)
		var validation auth.ValidationError
		switch {
		case errors.As(err, &validation):
			httpjson.Error(w, 400, err.Error())
		case errors.Is(err, ErrKeysRejected):
			httpjson.Error(w, 422, ErrKeysRejected.Error())
		case err != nil:
			httpjson.Error(w, 503, "accounts temporarily unavailable")
		default:
			httpjson.Write(w, 201, a)
		}
	})
	r.Delete("/v1/accounts/{id}", func(w http.ResponseWriter, r *http.Request) {
		if s.Live == nil {
			httpjson.Error(w, 503, "accounts temporarily unavailable")
			return
		}
		err := s.Live.Delete(r.Context(), auth.UserFromContext(r.Context()).ID, chi.URLParam(r, "id"))
		switch {
		case errors.Is(err, ErrAccountNotFound):
			httpjson.Error(w, 404, "Account not found / Hesap bulunamadı")
		case errors.Is(err, ErrDeleteConflict):
			httpjson.Error(w, 409, "Live account has open positions or account is paper / Canlı hesabın açık pozisyonları var veya hesap kağıt hesabıdır")
		case err != nil:
			httpjson.Error(w, 503, "accounts temporarily unavailable")
		default:
			w.WriteHeader(204)
		}
	})
}

func (p Postgres) CreateLive(ctx context.Context, owner string, r LiveRequest, c Credentials, balance decimal.Decimal) (Account, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return Account{}, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO app.accounts(user_id,type,name,market,starting_balance) VALUES($1::uuid,'live',$2,$3,$4::numeric) RETURNING id::text`, owner, r.Name, r.Market, balance.String()).Scan(&id)
	if err != nil {
		return Account{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.exchange_credentials(account_id,environment,key_hint,api_key_enc,api_secret_enc) VALUES($1::uuid,$2,$3,$4,$5)`, id, c.Environment, c.KeyHint, c.Key, c.Secret)
	if err != nil {
		return Account{}, err
	}
	a, err := scanAccount(tx.QueryRow(ctx, `SELECT `+columns+` FROM app.accounts WHERE id=$1::uuid`, id))
	if err != nil {
		return Account{}, err
	}
	return a, tx.Commit(ctx)
}
func (p Postgres) Credentials(ctx context.Context, id string) (Credentials, error) {
	var c Credentials
	err := p.Pool.QueryRow(ctx, `SELECT c.environment,a.market,c.key_hint,c.api_key_enc,c.api_secret_enc FROM app.exchange_credentials c JOIN app.accounts a ON a.id=c.account_id WHERE a.id=$1::uuid AND a.type='live'`, id).Scan(&c.Environment, &c.Market, &c.KeyHint, &c.Key, &c.Secret)
	return c, err
}
func (p Postgres) DeleteLive(ctx context.Context, owner, id string) error {
	if !validAccountID(id) {
		return ErrAccountNotFound
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	var kind string
	err = tx.QueryRow(ctx, `SELECT type FROM app.accounts WHERE id=$1::uuid AND user_id=$2::uuid FOR UPDATE`, id, owner).Scan(&kind)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}
	var open bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app.positions WHERE account_id=$1::uuid AND status='open')`, id).Scan(&open); err != nil {
		return err
	}
	if kind != "live" || open {
		return ErrDeleteConflict
	}
	if _, err = tx.Exec(ctx, `DELETE FROM app.accounts WHERE id=$1::uuid`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func validAccountID(id string) bool { // UUID validation without database casts on untrusted paths.
	if len(id) != 36 {
		return false
	}
	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}
