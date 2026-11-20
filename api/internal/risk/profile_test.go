package risk

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const profileBody = `{"risk_per_trade_percent":"1.12345678","target_risk_reward":"2","total_open_risk_limit_percent":"6","max_open_positions":5,"max_losing_trades_today":3,"daily_loss_limit_percent":"5","max_stopped_trades_this_week":5,"weekly_loss_limit_percent":"10"}`

func TestProfileValidationAndDefaults(t *testing.T) {
	p := DefaultProfile()
	if !p.Valid(true) || p.RiskPerTradePercent != "2" || p.TargetRiskReward != "2" || p.TotalOpenRiskLimitPercent != "6" ||
		*p.MaxOpenPositions != 5 || *p.MaxLosingTradesToday != 3 || p.DailyLossLimitPercent != "5" ||
		*p.MaxStoppedTradesThisWeek != 5 || p.WeeklyLossLimitPercent != "10" {
		t.Fatalf("wrong defaults: %+v", p)
	}
	// Each required field independently rejects absence and invalid limits.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(profileBody), &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		invalid := []string{`null`, `0`, `-1`, `1.5`, `4294967296`, `"1"`}
		valid := `4294967295`
		if strings.HasSuffix(name, "percent") || name == "target_risk_reward" {
			invalid = []string{`null`, `""`, `"0"`, `"100.00000001"`, `"1.123456789"`, `"-1"`, `"+1"`, `"1e1"`, `" 1"`, `1`}
			valid = `"100.00000000"`
		}
		original := fields[name]
		for _, value := range append(invalid, valid) {
			fields[name] = json.RawMessage(value)
			encoded, _ := json.Marshal(fields)
			var p Profile
			err := json.Unmarshal(encoded, &p)
			if got := err == nil && p.Valid(true); got != (value == valid) {
				t.Fatalf("%s=%s: valid=%t err=%v", name, value, got, err)
			}
		}
		delete(fields, name)
		encoded, _ := json.Marshal(fields)
		var p Profile
		if err := json.Unmarshal(encoded, &p); err != nil || p.Valid(true) || !p.Valid(false) {
			t.Fatalf("required vs preview field %s: %+v %v", name, p, err)
		}
		fields[name] = original
	}
}

func TestResolvePreviewKeepsStoredLimits(t *testing.T) {
	stored := DefaultProfile()
	stored.DailyLossLimitPercent, stored.MaxOpenPositions = "0.5", count(2)
	preview := Profile{RiskPerTradePercent: "0.1"}
	effective := ResolveProfile(&stored, &preview)
	if effective.RiskPerTradePercent != "0.1" || effective.DailyLossLimitPercent != "0.5" || *effective.MaxOpenPositions != 2 || stored.RiskPerTradePercent != "2" {
		t.Fatalf("lost stored limits or mutated profile: %+v %+v", effective, stored)
	}
	*effective.MaxOpenPositions = 10
	if *stored.MaxOpenPositions != 2 || !reflect.DeepEqual(ResolveProfile(nil, nil), DefaultProfile()) {
		t.Fatal("defaults or profile isolation failed")
	}
}

func profileRequest(h http.Handler, method, path, body, cookie string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: "metron_session", Value: cookie})
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestProfileHandlers(t *testing.T) {
	f := &fakeAccounts{}
	h := router(t, f, &fakeEngine{})
	path := "/v1/accounts/" + accountID + "/risk-profile"
	w := profileRequest(h, "GET", path, "", token)
	var saved SavedProfile
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil || w.Code != 200 || saved.UpdatedAt != nil || !reflect.DeepEqual(saved.Profile, DefaultProfile()) {
		t.Fatalf("default response: %d %s %v", w.Code, w.Body, err)
	}
	w = profileRequest(h, "PUT", path, profileBody, token)
	if err := json.Unmarshal(w.Body.Bytes(), &saved); err != nil || w.Code != 200 || saved.UpdatedAt == nil || saved.RiskPerTradePercent != "1.12345678" {
		t.Fatalf("saved response: %d %s %v", w.Code, w.Body, err)
	}
	w = profileRequest(h, "GET", path, "", token)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"risk_per_trade_percent":"1.12345678"`) || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("get saved profile: %d %s", w.Code, w.Body)
	}
	for _, body := range []string{`null`, `{}`, `[]`, profileBody + `{}`, strings.TrimSuffix(profileBody, "}") + `,"updated_at":null}`, strings.TrimSuffix(profileBody, "}") + `,"unknown":1}`, strings.Repeat(" ", 4096) + profileBody} {
		w := profileRequest(h, "PUT", path, body, token)
		if w.Code != 400 || !strings.Contains(w.Body.String(), `"code":"invalid_input"`) || !strings.Contains(w.Body.String(), "error_tr") {
			t.Fatalf("strict profile body: %d %s", w.Code, w.Body)
		}
	}
	for _, method := range []string{"GET", "PUT"} {
		for _, id := range []string{foreignID, "unknown"} {
			w := profileRequest(h, method, strings.Replace(path, accountID, id, 1), profileBody, token)
			if w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"account_not_found"`) {
				t.Fatalf("owner filter: %d %s", w.Code, w.Body)
			}
		}
		if w := profileRequest(h, method, path, profileBody, ""); w.Code != 401 {
			t.Fatalf("session required: %d %s", w.Code, w.Body)
		}
	}
	f.err = errors.New("private detail")
	w = profileRequest(h, "GET", path, "", token)
	if w.Code != 503 || strings.Contains(w.Body.String(), "private") {
		t.Fatalf("unsafe failure: %d %s", w.Code, w.Body)
	}
}

func TestCheckUsesStoredAndPreviewProfilesAndRecords(t *testing.T) {
	stored := DefaultProfile()
	stored.DailyLossLimitPercent = "0.5"
	f, e := &fakeAccounts{profile: &stored}, &fakeEngine{}
	h := router(t, f, e)
	for _, preview := range []bool{false, true} {
		body, source, expectedRisk := validBody, "stored", "2"
		if preview {
			body = strings.TrimSuffix(body, "}") + `,"risk_profile":{"risk_per_trade_percent":"0.1"}}`
			source, expectedRisk = "preview", "0.1"
		}
		w := post(h, body, "application/json", token)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"profile_source":"`+source+`"`) ||
			e.request.RiskProfile.DailyLossLimitPercent != "0.5" || e.request.RiskProfile.RiskPerTradePercent != expectedRisk || ProfileSource(f.recordBody) != source {
			t.Fatalf("profile/check: %d %s %+v", w.Code, w.Body, e.request.RiskProfile)
		}
	}
	if f.recorded != 2 || stored.RiskPerTradePercent != "2" {
		t.Fatal("preview mutated stored profile or check not recorded")
	}
	f.recordErr = errors.New("private recording failure")
	w := post(h, validBody, "application/json", token)
	if w.Code != 503 || strings.Contains(w.Body.String(), "private") {
		t.Fatalf("audit write failure: %d %s", w.Code, w.Body)
	}
}

func TestRiskHistoryHandler(t *testing.T) {
	f := &fakeAccounts{}
	h := router(t, f, &fakeEngine{})
	path := "/v1/accounts/" + accountID + "/risk-checks"
	for _, tc := range []struct {
		query       string
		code, limit int
	}{
		{"", 200, 20}, {"?limit=1", 200, 1}, {"?limit=100", 200, 100},
		{"?limit=0", 400, 0}, {"?limit=101", 400, 0}, {"?limit=-1", 400, 0}, {"?limit=", 400, 0},
		{"?limit=1.5", 400, 0}, {"?limit=1&limit=2", 400, 0}, {"?limit=abc", 400, 0},
	} {
		w := profileRequest(h, "GET", path+tc.query, "", token)
		if w.Code != tc.code || (tc.code == 200 && (f.limit != tc.limit || strings.TrimSpace(w.Body.String()) != `{"checks":[]}`)) {
			t.Fatalf("history %s: %d %s", tc.query, w.Code, w.Body)
		}
	}
	for _, id := range []string{foreignID, "unknown"} {
		w := profileRequest(h, "GET", strings.Replace(path, accountID, id, 1), "", token)
		if w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"account_not_found"`) {
			t.Fatalf("history isolation: %d %s", w.Code, w.Body)
		}
	}
	if w := profileRequest(h, "GET", path, "", ""); w.Code != 401 {
		t.Fatalf("history session: %d %s", w.Code, w.Body)
	}
}
