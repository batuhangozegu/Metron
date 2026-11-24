package accounts

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/batuhangozegu/metron/api/internal/binance"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type liveAccountStore struct {
	credentials Credentials
	kind        string
	open        bool
	deleted     bool
	reads       int
}

func (f *liveAccountStore) CreateLive(_ context.Context, owner string, r LiveRequest, c Credentials, _ decimal.Decimal) (Account, error) {
	f.credentials = c
	f.kind = "live"
	return Account{ID: "11111111-1111-4111-8111-111111111111", UserID: owner, Type: "live", Market: r.Market, Name: r.Name, QuoteAsset: "USDT", Environment: &c.Environment, KeyHint: &c.KeyHint}, nil
}
func (f *liveAccountStore) Credentials(context.Context, string) (Credentials, error) {
	f.reads++
	if f.deleted {
		return Credentials{}, ErrAccountNotFound
	}
	return f.credentials, nil
}
func (f *liveAccountStore) DeleteLive(_ context.Context, owner, id string) error {
	if owner != "owner" || id != "11111111-1111-4111-8111-111111111111" {
		return ErrAccountNotFound
	}
	if f.kind != "live" || f.open {
		return ErrDeleteConflict
	}
	f.deleted = true
	return nil
}

type keysExchange struct {
	Exchange
	verifyError error
}

func (f keysExchange) VerifyKeys(context.Context) error { return f.verifyError }
func (f keysExchange) Balance(context.Context, string) (binance.Balance, error) {
	return binance.Balance{Free: decimal.NewFromInt(500), Total: decimal.NewFromInt(500)}, nil
}
func liveAccountFixture(t *testing.T) (*LiveService, *liveAccountStore) {
	t.Helper()
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	store := &liveAccountStore{}
	return &LiveService{Store: store, Box: box, Factory: func(binance.Config) (Exchange, error) { return keysExchange{}, nil }}, store
}
func liveBody() LiveRequest {
	return LiveRequest{Name: " Binance ", Market: "spot", Environment: "testnet", APIKey: "private-key-WXYZ", APISecret: "private-secret"}
}

func TestLiveCredentialsEncryptedAndCached(t *testing.T) {
	s, store := liveAccountFixture(t)
	a, err := s.Create(context.Background(), "owner", liveBody())
	if err != nil {
		t.Fatal(err)
	}
	if a.Type != "live" || a.Name != "Binance" || a.PaperBalance != nil || *a.Environment != "testnet" || *a.KeyHint != "•••WXYZ" {
		t.Fatalf("bad account %+v", a)
	}
	encoded, _ := json.Marshal(a)
	if strings.Contains(string(encoded), "private") {
		t.Fatal("credentials in account response")
	}
	key, err := s.Box.Open(store.credentials.Key, "binance-api-key")
	if err != nil || string(key) != liveBody().APIKey {
		t.Fatal("key encryption mismatch", err)
	}
	if _, err = s.Box.Open(store.credentials.Secret, "binance-api-key"); err == nil {
		t.Fatal("purpose not bound")
	}
	factories := 0
	s.Factory = func(cfg binance.Config) (Exchange, error) {
		factories++
		if cfg.Credentials.APIKey != liveBody().APIKey || cfg.Credentials.Secret != liveBody().APISecret || cfg.Env != "testnet" || cfg.Market != "spot" {
			t.Fatal("wrong decrypted config")
		}
		return keysExchange{}, nil
	}
	for i := 0; i < 2; i++ {
		if _, _, err = s.Client(context.Background(), a.ID); err != nil {
			t.Fatal(err)
		}
	}
	if factories != 1 || store.reads != 1 {
		t.Fatal("client not cached")
	}
	if err = s.Delete(context.Background(), "owner", a.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = s.Client(context.Background(), a.ID); !errors.Is(err, ErrAccountNotFound) {
		t.Fatal("deleted client remains cached", err)
	}
}

func TestLiveAccountValidationAndKeyRejection(t *testing.T) {
	for _, mutate := range []func(*LiveRequest){func(r *LiveRequest) { r.Market = "margin" }, func(r *LiveRequest) { r.Environment = "prod" }, func(r *LiveRequest) { r.Name = "" }, func(r *LiveRequest) { r.APIKey = "abc" }, func(r *LiveRequest) { r.APISecret = "" }} {
		s, _ := liveAccountFixture(t)
		r := liveBody()
		mutate(&r)
		if _, err := s.Create(context.Background(), "owner", r); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
	s, store := liveAccountFixture(t)
	s.Factory = func(binance.Config) (Exchange, error) { return keysExchange{verifyError: binance.ErrInvalidKeys}, nil }
	if _, err := s.Create(context.Background(), "owner", liveBody()); !errors.Is(err, ErrKeysRejected) || len(store.credentials.Key) != 0 {
		t.Fatal("bad keys persisted", err)
	}
	router := chi.NewRouter()
	(Service{Live: s}).Mount(router)
	body, _ := json.Marshal(liveBody())
	request := httptest.NewRequest("POST", "/v1/accounts/live", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	if w.Code != 422 || !strings.Contains(w.Body.String(), `"error":"Binance rejected the API keys"`) {
		t.Fatalf("wrong HTTP key failure %d %s", w.Code, w.Body)
	}
}

func TestLiveDeleteOwnershipAndOpenPositions(t *testing.T) {
	s, store := liveAccountFixture(t)
	a, err := s.Create(context.Background(), "owner", liveBody())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(context.Background(), "foreign", a.ID); !errors.Is(err, ErrAccountNotFound) {
		t.Fatal(err)
	}
	store.open = true
	if err = s.Delete(context.Background(), "owner", a.ID); !errors.Is(err, ErrDeleteConflict) {
		t.Fatal(err)
	}
	store.open = false
	store.kind = "paper"
	if err = s.Delete(context.Background(), "owner", a.ID); !errors.Is(err, ErrDeleteConflict) {
		t.Fatal(err)
	}
	if store.deleted {
		t.Fatal("invalid deletion reached storage")
	}
}
