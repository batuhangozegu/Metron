package auth_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/batuhangozegu/metron/api/internal/risk"
)

// Requires a dedicated test database with db/app migrations applied. Only this
// test's unique users are deleted; it never drops schemas or truncates tables.
func TestPostgresAccountsAndSignIn(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test Postgres database")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	var migrated bool
	if err = pool.QueryRow(ctx, `SELECT to_regclass('app.users') IS NOT NULL`).Scan(&migrated); err != nil || !migrated {
		t.Fatalf("apply db/app goose migrations first: %v", err)
	}
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	store := auth.Postgres{Pool: pool}
	service, err := auth.New(store, box)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(30 * time.Second)
	service.Now = func() time.Time { return now }
	router := chi.NewRouter()
	auth.HTTP{Service: service}.Mount(router, accounts.Service{Store: accounts.Postgres{Pool: pool}}.Mount)
	request := func(method, path, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "127.0.0.1:4567"
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	want := func(res *httptest.ResponseRecorder, status int) {
		t.Helper()
		if res.Code != status {
			t.Fatalf("HTTP %d want %d: %s", res.Code, status, res.Body.String())
		}
	}
	cookie := func(res *httptest.ResponseRecorder, name string) *http.Cookie {
		t.Helper()
		for _, c := range res.Result().Cookies() {
			if c.Name == name && c.Value != "" {
				return c
			}
		}
		t.Fatalf("missing cookie %s", name)
		return nil
	}
	random := make([]byte, 8)
	if _, err = rand.Read(random); err != nil {
		t.Fatal(err)
	}
	email := "integration-" + hex.EncodeToString(random) + "@example.com"
	credentials := fmt.Sprintf(`{"email":%q,"password":"integration password!"}`, email)
	res := request("POST", "/v1/auth/register", credentials)
	want(res, 201)
	var user auth.User
	if err = json.Unmarshal(res.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(ctx, `DELETE FROM app.users WHERE id=$1`, user.ID); err != nil {
			t.Error(err)
		}
	}()
	full := cookie(res, auth.CookieName)
	var storedHash []byte
	if err = pool.QueryRow(ctx, `SELECT token_hash FROM app.sessions WHERE user_id=$1`, user.ID).Scan(&storedHash); err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(storedHash) != hex.EncodeToString(auth.TokenHash(full.Value)) {
		t.Fatal("session hash does not match")
	}
	want(request("POST", "/v1/auth/register", strings.Replace(credentials, email, strings.ToUpper(email), 1)), 409)
	want(request("POST", "/v1/auth/login", strings.Replace(credentials, "integration password!", "incorrect password!", 1)), 401)
	want(request("POST", "/v1/auth/login", strings.Replace(credentials, email, "unknown-"+email, 1)), 401)
	want(request("GET", "/v1/me", "", full), 200)
	want(request("POST", "/v1/accounts", `{"name":"Paper","paper_balance":100}`, full), 400)
	res = request("POST", "/v1/accounts", `{"name":"Paper","paper_balance":"1234.12345678"}`, full)
	want(res, 201)
	var account accounts.Account
	if err = json.Unmarshal(res.Body.Bytes(), &account); err != nil {
		t.Fatal(err)
	}
	if account.Type != "paper" || account.Market != "spot" || account.PaperBalance == nil || *account.PaperBalance != "1234.12345678" {
		t.Fatalf("precision/account: %+v", account)
	}
	res = request("GET", "/v1/accounts", "", full)
	want(res, 200)
	var list []accounts.Account
	if err = json.Unmarshal(res.Body.Bytes(), &list); err != nil || len(list) != 1 {
		t.Fatalf("list: %s %v", res.Body.String(), err)
	}
	want(request("POST", "/v1/accounts", `{"name":"Invalid","paper_balance":"100","market":"margin"}`, full), 400)
	res = request("POST", "/v1/accounts", `{"name":"Futures","paper_balance":"2000.00000001","market":"futures"}`, full)
	want(res, 201)
	var futures accounts.Account
	if err = json.Unmarshal(res.Body.Bytes(), &futures); err != nil || futures.Market != "futures" || futures.PaperBalance == nil || *futures.PaperBalance != "2000.00000001" {
		t.Fatalf("futures account: %s %v", res.Body.String(), err)
	}
	res = request("GET", "/v1/accounts", "", full)
	want(res, 200)
	if err = json.Unmarshal(res.Body.Bytes(), &list); err != nil || len(list) != 2 || list[0].Market != "spot" || list[1].Market != "futures" {
		t.Fatalf("market list: %s %v", res.Body.String(), err)
	}
	riskStore := risk.Postgres{Pool: pool}
	for _, a := range list {
		got, err := riskStore.RiskAccount(ctx, user.ID, a.ID)
		if err != nil || got.Market != a.Market || got.Equity != *a.PaperBalance {
			t.Fatalf("risk account: %+v %v", got, err)
		}
	}
	// Migration defaults old-style inserts to spot and constrains markets.
	var defaultMarket string
	if err = pool.QueryRow(ctx, `INSERT INTO app.accounts(user_id,type,name,paper_balance) VALUES($1,'paper','default',100) RETURNING market`, user.ID).Scan(&defaultMarket); err != nil || defaultMarket != "spot" {
		t.Fatalf("market default: %s %v", defaultMarket, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO app.accounts(user_id,type,name,paper_balance,market) VALUES($1,'paper','invalid',100,'margin')`, user.ID); err == nil {
		t.Fatal("database accepted invalid market")
	}
	// Isolation: another user's list cannot expose the first account.
	res = request("POST", "/v1/auth/register", strings.Replace(credentials, email, "other-"+email, 1))
	want(res, 201)
	var other auth.User
	if err = json.Unmarshal(res.Body.Bytes(), &other); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(ctx, `DELETE FROM app.users WHERE id=$1`, other.ID); err != nil {
			t.Error(err)
		}
	}()
	otherCookie := cookie(res, auth.CookieName)
	if _, err = riskStore.RiskAccount(ctx, other.ID, futures.ID); err != risk.ErrAccountNotFound {
		t.Fatalf("risk account isolation: %v", err)
	}
	res = request("GET", "/v1/accounts", "", otherCookie)
	want(res, 200)
	if strings.TrimSpace(res.Body.String()) != "[]" {
		t.Fatal("cross-user account leak")
	}
	// DB constraints also reject a paper balance missing or over the cap.
	for _, value := range []string{"NULL", "0", "10000000.00000001", "0.000000001"} {
		if _, err = pool.Exec(ctx, `INSERT INTO app.accounts(user_id,type,name,paper_balance) VALUES($1,'paper','invalid',`+value+`)`, user.ID); err == nil {
			t.Fatal("database accepted invalid paper balance", value)
		}
	}
	res = request("POST", "/v1/auth/2fa/setup", `{}`, full)
	want(res, 200)
	var setup struct {
		Secret string `json:"secret"`
		URL    string `json:"otpauth_url"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &setup); err != nil {
		t.Fatal(err)
	}
	code := func() string {
		t.Helper()
		value, err := totp.GenerateCode(setup.Secret, now)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	body := func() string { return fmt.Sprintf(`{"code":%q}`, code()) }
	want(request("POST", "/v1/auth/2fa/enable", body(), full), 200)
	res = request("POST", "/v1/auth/login", credentials)
	want(res, 200)
	if !strings.Contains(res.Body.String(), `"two_factor_required":true`) {
		t.Fatal("2FA bypassed")
	}
	pending := cookie(res, auth.PendingCookieName)
	want(request("GET", "/v1/me", "", pending), 401)
	// A pending token is never accepted even if placed in the full-session cookie.
	want(request("GET", "/v1/me", "", &http.Cookie{Name: auth.CookieName, Value: pending.Value}), 401)
	want(request("POST", "/v1/auth/login/2fa", body(), pending), 401) // enable code replay
	now = now.Add(30 * time.Second)
	res = request("POST", "/v1/auth/login/2fa", body(), pending)
	want(res, 200)
	signedIn := cookie(res, auth.CookieName)
	want(request("GET", "/v1/me", "", signedIn), 200)
	want(request("POST", "/v1/auth/login/2fa", body(), pending), 401) // consumed pending token
	// Two concurrent attempts using different pending tokens and the same code
	// must still consume the user's counter exactly once.
	_, p1, _, err := service.Login(ctx, email, "integration password!", "127.0.0.2", "test")
	if err != nil {
		t.Fatal(err)
	}
	_, p2, _, err := service.Login(ctx, email, "integration password!", "127.0.0.2", "test")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(30 * time.Second)
	currentCode := code()
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, token := range []string{p1, p2} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			_, _, err := service.CompleteLogin(ctx, token, currentCode, "127.0.0.2", "test")
			results <- err
		}(token)
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent TOTP successes = %d", successes)
	}
	now = now.Add(30 * time.Second)
	disable := fmt.Sprintf(`{"code":%q,"password":"integration password!"}`, code())
	want(request("POST", "/v1/auth/2fa/disable", disable, signedIn), 200)
	want(request("GET", "/v1/me", "", full), 401) // revoked at disable
	// Expired sessions stay expired even though resolution normally slides.
	if _, err = pool.Exec(ctx, `UPDATE app.sessions SET expires_at=$2 WHERE token_hash=$1`, auth.TokenHash(signedIn.Value), now); err != nil {
		t.Fatal(err)
	}
	want(request("GET", "/v1/me", "", signedIn), 401)
	res = request("POST", "/v1/auth/login", credentials)
	want(res, 200)
	last := cookie(res, auth.CookieName)
	want(request("POST", "/v1/auth/logout", `{}`, last), 200)
	want(request("GET", "/v1/me", "", last), 401)
}
