package risk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseSession struct {
	auth.Store
	owner string
}

func (s databaseSession) ResolveSession(context.Context, []byte, bool, time.Time) (auth.User, error) {
	return auth.User{ID: s.owner}, nil
}

func TestPostgresProfilesAndChecks(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test Postgres database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var owner, id string
	if err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES(gen_random_uuid()::text || '@risk.example.com','test-only') RETURNING id::text`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, owner); err != nil {
			t.Error(err)
		}
	}()
	if err := pool.QueryRow(ctx, `INSERT INTO app.accounts(user_id,type,name,paper_balance) VALUES($1::uuid,'paper','Risk profile',1000) RETURNING id::text`, owner).Scan(&id); err != nil {
		t.Fatal(err)
	}
	store := Postgres{Pool: pool}
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	h := chi.NewRouter()
	engine := &fakeEngine{}
	auth.HTTP{Service: &auth.Service{Store: databaseSession{owner: owner}, Box: box, Now: time.Now}}.Mount(h, Service{Accounts: store, Store: store, Engine: engine}.Mount)
	path := "/v1/accounts/" + id + "/risk-profile"
	w := profileRequest(h, "GET", path, "", token)
	var saved SavedProfile
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil || w.Code != 200 || saved.UpdatedAt != nil || !reflect.DeepEqual(saved.Profile, DefaultProfile()) {
		t.Fatalf("defaults: %d %s %v", w.Code, w.Body, err)
	}
	w = profileRequest(h, "PUT", path, profileBody, token)
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil || w.Code != 200 || saved.UpdatedAt == nil || saved.RiskPerTradePercent != "1.12345678" {
		t.Fatalf("save: %d %s %v", w.Code, w.Body, err)
	}
	got, err := store.Profile(ctx, owner, id)
	if err != nil || !reflect.DeepEqual(got, saved) {
		t.Fatalf("round trip: %+v %+v %v", got, saved, err)
	}
	// Upserts preserve uint32's full range, which requires bigint storage.
	p := saved.Profile
	p.MaxOpenPositions = count(^uint32(0))
	updated, err := store.SaveProfile(ctx, owner, id, p)
	if err != nil || *updated.MaxOpenPositions != ^uint32(0) || updated.UpdatedAt.Before(*saved.UpdatedAt) {
		t.Fatalf("update: %+v %v", updated, err)
	}
	foreignHTTP := chi.NewRouter()
	auth.HTTP{Service: &auth.Service{Store: databaseSession{owner: foreignID}, Box: box, Now: time.Now}}.Mount(foreignHTTP, Service{Accounts: store, Store: store, Engine: engine}.Mount)
	for _, endpoint := range []struct{ method, suffix, body string }{
		{"GET", "risk-profile", ""}, {"PUT", "risk-profile", profileBody}, {"GET", "risk-checks", ""},
	} {
		w := profileRequest(foreignHTTP, endpoint.method, "/v1/accounts/"+id+"/"+endpoint.suffix, endpoint.body, token)
		if w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"account_not_found"`) {
			t.Fatalf("foreign account HTTP: %d %s", w.Code, w.Body)
		}
	}
	for _, foreign := range []string{foreignID, "33333333-3333-4333-8333-333333333333"} {
		for _, suffix := range []string{"risk-profile", "risk-checks"} {
			w := profileRequest(h, "GET", "/v1/accounts/"+foreign+"/"+suffix, "", token)
			if w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"account_not_found"`) {
				t.Fatalf("missing account: %d %s", w.Code, w.Body)
			}
		}
		if _, err := store.Profile(ctx, foreign, id); !errors.Is(err, ErrAccountNotFound) {
			t.Fatalf("foreign profile: %v", err)
		}
		if _, err := store.SaveProfile(ctx, foreign, id, p); !errors.Is(err, ErrAccountNotFound) {
			t.Fatalf("foreign save: %v", err)
		}
		if _, err := store.Checks(ctx, foreign, id, 20); !errors.Is(err, ErrAccountNotFound) {
			t.Fatalf("foreign history: %v", err)
		}
	}
	// Both successful and blocked previews persist their exact protobuf result.
	for _, preview := range []bool{false, true} {
		engine.blocked = preview
		body := strings.Replace(validBody, accountID, id, 1)
		if preview {
			body = strings.TrimSuffix(body, "}") + `,"risk_profile":{"risk_per_trade_percent":"0.1"}}`
		}
		w := post(h, body, "application/json", token)
		if w.Code != 200 {
			t.Fatalf("check: %d %s", w.Code, w.Body)
		}
	}
	checks, err := store.Checks(ctx, owner, id, 1)
	if err != nil || len(checks) != 1 || checks[0].ProfileSource != "preview" || checks[0].Source != "check" || checks[0].OrderID != nil || checks[0].Allowed || checks[0].Verdict != "fail" ||
		checks[0].EntryPrice != "100.12345678" || !reflect.DeepEqual(checks[0].FailedRules, []string{"daily_loss"}) || len(checks[0].CautionRules) != 0 {
		t.Fatalf("history: %+v %v", checks, err)
	}
	var result []byte
	if err := pool.QueryRow(ctx, `SELECT result FROM app.risk_checks WHERE id=$1::uuid`, checks[0].ID).Scan(&result); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result), `"position_size": "1.12345678"`) || !strings.Contains(string(result), `"message_tr"`) {
		t.Fatalf("incomplete protojson result: %s", result)
	}
	w = profileRequest(h, http.MethodGet, "/v1/accounts/"+id+"/risk-checks?limit=100", "", token)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"failed_rules":[]`) {
		t.Fatalf("history wire format: %d %s", w.Code, w.Body)
	}
}
