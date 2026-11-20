package risk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
)

const accountID = "11111111-1111-4111-8111-111111111111"
const foreignID = "22222222-2222-4222-8222-222222222222"
const token = "sssssssssssssssssssssssssssssssssssssssssss"
const validBody = `{"account_id":"` + accountID + `","symbol":"BTCUSDT","side":"long","entry_price":"100.12345678"}`

type fakeAuth struct{ auth.Store }

func (fakeAuth) ResolveSession(_ context.Context, hash []byte, pending bool, _ time.Time) (auth.User, error) {
	if pending || !bytes.Equal(hash, auth.TokenHash(token)) {
		return auth.User{}, auth.ErrUnauthorized
	}
	return auth.User{ID: "user-1"}, nil
}

type fakeAccounts struct {
	market        string
	user, account string
	calls         int
	err           error
	activity      *enginev1.Activity
	holding       string
	profile       *Profile
	saved         SavedProfile
	recorded      int
	recordBody    Request
	recordErr     error
	limit         int
}

func (f *fakeAccounts) RiskAccount(_ context.Context, user, account, _ string) (Account, error) {
	f.user, f.account = user, account
	f.calls++
	if f.err != nil {
		return Account{}, f.err
	}
	if user != "user-1" || account != accountID {
		return Account{}, ErrAccountNotFound
	}
	market := f.market
	if market == "" {
		market = "spot"
	}
	activity, holding := f.activity, f.holding
	if activity == nil {
		activity = &enginev1.Activity{OpenRiskPercent: "0", DailyLossPercent: "0", WeeklyLossPercent: "0"}
	}
	if holding == "" {
		holding = "0"
	}
	return Account{Equity: "12345.12345678", Market: market, CurrentClassHolding: holding, Activity: activity, Profile: f.profile}, nil
}

func (f *fakeAccounts) Profile(_ context.Context, owner, id string) (SavedProfile, error) {
	if owner != "user-1" || id != accountID {
		return SavedProfile{}, ErrAccountNotFound
	}
	if f.err != nil {
		return SavedProfile{}, f.err
	}
	if f.saved.UpdatedAt == nil {
		return SavedProfile{Profile: DefaultProfile()}, nil
	}
	return f.saved, nil
}

func (f *fakeAccounts) SaveProfile(ctx context.Context, owner, id string, p Profile) (SavedProfile, error) {
	if _, err := f.Profile(ctx, owner, id); err != nil {
		return SavedProfile{}, err
	}
	now := time.Now().UTC()
	f.saved = SavedProfile{Profile: p, UpdatedAt: &now}
	f.profile = &f.saved.Profile
	return f.saved, nil
}

func (f *fakeAccounts) RecordCheck(_ context.Context, _, _ string, body Request, _ *enginev1.CheckResponse) error {
	f.recorded++
	f.recordBody = body
	return f.recordErr
}

func (f *fakeAccounts) Checks(ctx context.Context, owner, id string, limit int) ([]CheckRecord, error) {
	if _, err := f.Profile(ctx, owner, id); err != nil {
		return nil, err
	}
	f.limit = limit
	return nil, nil
}

type fakeEngine struct {
	request  *enginev1.CheckRequest
	calls    int
	err      error
	deadline bool
	blocked  bool
}

func (f *fakeEngine) Check(ctx context.Context, request *enginev1.CheckRequest, _ ...grpc.CallOption) (*enginev1.CheckResponse, error) {
	f.request = request
	f.calls++
	_, f.deadline = ctx.Deadline()
	if f.err != nil {
		return nil, f.err
	}
	result := &enginev1.CheckResponse{
		Market: request.Market, Leverage: max(1, request.GetLeverage()),
		Symbol: request.Symbol, Side: request.Side, AccountEquity: request.AccountEquity,
		PositionSize: "1.12345678", RiskAmount: "123.4512345678", ActualRisk: "123.45123456", Notional: "112.345678",
		Verdict: "caution", Allowed: true, Passed: 5, Cautions: 1,
		Checks: []*enginev1.RuleCheck{{RuleId: "class_risk", Status: "caution", Values: map[string]string{"effective_percent": "0.5"}, MessageEn: "Risk is capped.", MessageTr: "Risk sınırlandırılır."}},
	}
	if request.Market == "futures" {
		liquidation, margin := "90.45226130653266331658291457", "11.2345678"
		result.LiquidationPrice, result.InitialMargin = &liquidation, &margin
	}
	if f.blocked {
		result.Allowed, result.Verdict = false, "fail"
		result.Failed, result.Cautions = 1, 0
		result.Checks = []*enginev1.RuleCheck{{RuleId: "daily_loss", Status: "fail", MessageEn: "Daily loss limit reached.", MessageTr: "Günlük zarar sınırına ulaşıldı."}}
	}
	return result, nil
}

func router(t *testing.T, accounts *fakeAccounts, engine *fakeEngine) http.Handler {
	t.Helper()
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	h := auth.HTTP{Service: &auth.Service{Store: fakeAuth{}, Box: box, Now: time.Now}}
	h.Mount(r, Service{Accounts: accounts, Store: accounts, Engine: engine}.Mount)
	return r
}

func post(handler http.Handler, body, contentType, cookie string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/v1/risk/check", strings.NewReader(body))
	r.Header.Set("Content-Type", contentType)
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: cookie})
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestCheckForwardsOnlyOwnedBalanceAndDecimalStrings(t *testing.T) {
	for _, extra := range []string{"", `,"manual_stop":"90.12345678","risk_profile":{"risk_per_trade_percent":"0.5","target_risk_reward":"2.5"}`} {
		a, e := &fakeAccounts{}, &fakeEngine{}
		h := router(t, a, e)
		body := strings.TrimSuffix(validBody, "}") + extra + "}"
		w := post(h, body, "application/json; charset=utf-8", token)
		if w.Code != 200 {
			t.Fatalf("%d: %s", w.Code, w.Body)
		}
		if a.user != "user-1" || a.account != accountID || e.request.AccountEquity != "12345.12345678" || e.request.EntryPrice != "100.12345678" || !e.deadline {
			t.Fatalf("wrong engine input: %+v", e.request)
		}
		if extra == "" {
			if e.request.ManualStop != nil || e.request.RiskProfile.GetRiskPerTradePercent() != "2" {
				t.Fatal("missing defaults")
			}
		} else if e.request.GetManualStop() != "90.12345678" || e.request.RiskProfile.RiskPerTradePercent != "0.5" || e.request.RiskProfile.TargetRiskReward != "2.5" {
			t.Fatal("override lost")
		}
		var value map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		if value["risk_amount"] != "123.4512345678" || value["position_size"] != "1.12345678" || value["failed"] != float64(0) || value["allowed"] != true {
			t.Fatalf("wrong response: %s", w.Body)
		}
		if !strings.Contains(w.Body.String(), "message_tr") || strings.Contains(w.Body.String(), "AccountEquity") {
			t.Fatal("invalid wire format")
		}
		if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Type") != "application/json" || len(w.Result().Cookies()) == 0 {
			t.Fatal("missing response/session policy")
		}
	}
}

func TestCheckRejectsInvalidRequestsBeforeLookup(t *testing.T) {
	for _, body := range []string{
		`null`, `[]`, `{`, validBody + `{}`, strings.TrimSuffix(validBody, "}") + `,"equity":"99999"}`,
		strings.TrimSuffix(validBody, "}") + `,"risk_profile":{"unknown":"1"}}`,
		strings.Replace(validBody, `"100.12345678"`, `100.12345678`, 1),
		strings.Replace(validBody, `"100.12345678"`, `"1e2"`, 1),
		strings.Replace(validBody, `"100.12345678"`, `"0"`, 1),
		strings.Replace(validBody, `"100.12345678"`, `"-1"`, 1),
		strings.Replace(validBody, `"100.12345678"`, `"1.123456789"`, 1),
		strings.Replace(validBody, `"100.12345678"`, `"1000000000001"`, 1),
		strings.Replace(validBody, `"long"`, `"buy"`, 1),
		strings.Replace(validBody, `"BTCUSDT"`, `"BTC/USDT"`, 1),
		strings.TrimSuffix(validBody, "}") + `,"manual_stop":"0"}`,
		strings.TrimSuffix(validBody, "}") + `,"risk_profile":{"risk_per_trade_percent":"101"}}`,
		strings.TrimSuffix(validBody, "}") + `,"risk_profile":{"target_risk_reward":"0"}}`,
		strings.TrimSuffix(validBody, "}") + `,"manual_stop":"` + strings.Repeat("1", 4096) + `"}`,
	} {
		t.Run(body[:min(len(body), 100)], func(t *testing.T) {
			a, e := &fakeAccounts{}, &fakeEngine{}
			w := post(router(t, a, e), body, "application/json", token)
			if w.Code != 400 || a.calls != 0 || e.calls != 0 {
				t.Fatalf("%d: %s; calls %d %d", w.Code, w.Body, a.calls, e.calls)
			}
			if !strings.Contains(w.Body.String(), "error_tr") {
				t.Fatal("missing Turkish error")
			}
		})
	}
}

func TestCheckAccountIsolationAndUnavailableDependencies(t *testing.T) {
	for _, tc := range []struct {
		name, id            string
		storeErr, engineErr error
		code, engineCalls   int
	}{
		{"foreign", foreignID, nil, nil, 404, 0},
		{"unknown", "33333333-3333-4333-8333-333333333333", nil, nil, 404, 0},
		{"malformed", "not-a-uuid", nil, nil, 404, 0},
		{"database", accountID, errors.New("private database detail"), nil, 503, 0},
		{"engine", accountID, nil, status.Error(codes.Unavailable, "private engine detail"), 503, 1},
		{"timeout", accountID, nil, context.DeadlineExceeded, 503, 1},
		{"symbol", accountID, nil, status.Error(codes.NotFound, "private symbol detail"), 404, 1},
		{"wrong stop", accountID, nil, status.Error(codes.InvalidArgument, "wrong_stop_side"), 400, 1},
		{"zero stop", accountID, nil, status.Error(codes.InvalidArgument, "zero_stop_distance"), 400, 1},
		{"stable", accountID, nil, status.Error(codes.InvalidArgument, "stablecoin_unsupported"), 400, 1},
		{"invalid", accountID, nil, status.Error(codes.InvalidArgument, "private detail"), 400, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, e := &fakeAccounts{err: tc.storeErr}, &fakeEngine{err: tc.engineErr}
			w := post(router(t, a, e), strings.Replace(validBody, accountID, tc.id, 1), "application/json", token)
			if w.Code != tc.code || e.calls != tc.engineCalls {
				t.Fatalf("%d: %s; engine calls %d", w.Code, w.Body, e.calls)
			}
			if !strings.Contains(w.Body.String(), "error_tr") || strings.Contains(w.Body.String(), "private") {
				t.Fatal("unsafe or untranslated failure")
			}
		})
	}
}

func TestCheckSessionAndJSONPolicy(t *testing.T) {
	for _, tc := range []struct {
		cookie, media string
		code          int
	}{
		{"", "application/json", 401},
		{strings.Repeat("p", 43), "application/json", 401}, // pending/expired/unknown token
		{token, "text/plain", 415},
		{token, "", 415},
	} {
		a, e := &fakeAccounts{}, &fakeEngine{}
		w := post(router(t, a, e), validBody, tc.media, tc.cookie)
		if w.Code != tc.code || a.calls != 0 || e.calls != 0 {
			t.Fatalf("%d %s", w.Code, w.Body)
		}
		if !strings.Contains(w.Body.String(), "error_tr") {
			t.Fatal("missing translation")
		}
	}
}

func TestMarketLeverageAndTrustedActivity(t *testing.T) {
	for _, tc := range []struct {
		market, leverage string
		code             int
	}{
		{"spot", "", 200}, {"spot", `,"leverage":1`, 200}, {"spot", `,"leverage":2`, 400},
		{"futures", "", 400}, {"futures", `,"leverage":1`, 200}, {"futures", `,"leverage":125`, 200},
		{"futures", `,"leverage":0`, 400}, {"futures", `,"leverage":126`, 400},
		{"futures", `,"leverage":2.5`, 400}, {"futures", `,"leverage":"10"`, 400},
	} {
		a, e := &fakeAccounts{market: tc.market}, &fakeEngine{}
		w := post(router(t, a, e), strings.TrimSuffix(validBody, "}")+tc.leverage+"}", "application/json", token)
		if w.Code != tc.code {
			t.Fatalf("%s %s: %d %s", tc.market, tc.leverage, w.Code, w.Body)
		}
		if tc.code == 200 {
			r := e.request
			if r.Market != tc.market || r.CurrentClassHolding != "0" || r.Activity == nil || r.Activity.OpenRiskPercent != "0" || r.Activity.DailyLossPercent != "0" || r.Activity.WeeklyLossPercent != "0" || r.Activity.OpenPositions != 0 || r.Activity.LosingTradesToday != 0 || r.Activity.StoppedTradesThisWeek != 0 {
				t.Fatalf("untrusted activity or market: %+v", r)
			}
			var result map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result["market"] != tc.market || result["leverage"] != float64(max(1, r.GetLeverage())) {
				t.Fatalf("response: %s %v", w.Body, err)
			}
			if tc.market == "futures" {
				if result["liquidation_price"] != "90.45226130653266331658291457" || result["initial_margin"] != "11.2345678" {
					t.Fatalf("lost futures decimal strings: %s", w.Body)
				}
			} else if _, present := result["liquidation_price"]; present {
				t.Fatalf("unexpected spot liquidation: %s", w.Body)
			}
		} else if e.calls != 0 {
			t.Fatal("invalid leverage reached engine")
		}
	}
	for _, extra := range []string{
		`,"market":"futures"`, `,"activity":{"open_positions":0}`, `,"open_positions":0`, `,"open_risk_percent":"0"`,
		`,"losing_trades_today":0`, `,"daily_loss_percent":"0"`, `,"stopped_trades_this_week":0`, `,"weekly_loss_percent":"0"`, `,"current_class_holding":"0"`,
		`,"risk_profile":{"max_open_positions":0}`, `,"risk_profile":{"max_losing_trades_today":-1}`, `,"risk_profile":{"daily_loss_limit_percent":"101"}`,
	} {
		a, e := &fakeAccounts{}, &fakeEngine{}
		w := post(router(t, a, e), strings.TrimSuffix(validBody, "}")+extra+"}", "application/json", token)
		if w.Code != 400 || a.calls != 0 || e.calls != 0 {
			t.Fatalf("untrusted %s: %d %s", extra, w.Code, w.Body)
		}
	}
}

func TestDisciplineProfileOverrides(t *testing.T) {
	a, e := &fakeAccounts{market: "futures"}, &fakeEngine{}
	body := strings.TrimSuffix(validBody, "}") + `,"leverage":10,"risk_profile":{"total_open_risk_limit_percent":"4.5","max_open_positions":4,"max_losing_trades_today":2,"daily_loss_limit_percent":"3.5","max_stopped_trades_this_week":4,"weekly_loss_limit_percent":"7.5"}}`
	w := post(router(t, a, e), body, "application/json", token)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	p := e.request.RiskProfile
	if p.TotalOpenRiskLimitPercent != "4.5" || p.GetMaxOpenPositions() != 4 || p.GetMaxLosingTradesToday() != 2 || p.DailyLossLimitPercent != "3.5" || p.GetMaxStoppedTradesThisWeek() != 4 || p.WeeklyLossLimitPercent != "7.5" {
		t.Fatalf("lost profile: %+v", p)
	}
}

func TestCheckForwardsAuthoritativeActivity(t *testing.T) {
	a := &fakeAccounts{holding: "1200.123456789", activity: &enginev1.Activity{
		OpenRiskPercent: "3.123456789", OpenPositions: 4, LosingTradesToday: 2, DailyLossPercent: "1.25", StoppedTradesThisWeek: 3, WeeklyLossPercent: "2.5"}}
	e := &fakeEngine{}
	w := post(router(t, a, e), validBody, "application/json", token)
	if w.Code != 200 || e.request.Activity != a.activity || e.request.CurrentClassHolding != a.holding {
		t.Fatalf("wrong trusted snapshot %d %s %+v", w.Code, w.Body, e.request)
	}
}
