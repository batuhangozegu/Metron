package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

const testKey = "test-api-key-private"
const testSecret = "test-secret-private"

var testNow = time.UnixMilli(1700000000000)

func newTestClient(t *testing.T, market Market, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := New(Config{Env: "testnet", Market: market,
		Credentials: Credentials{APIKey: testKey, Secret: testSecret},
		BaseURL:     server.URL, HTTP: server.Client(), Now: func() time.Time { return testNow }})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func checkSigned(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("X-MBX-APIKEY") != testKey {
		t.Error("missing API key header")
	}
	query := r.URL.Query()
	signature := query.Get("signature")
	query.Del("signature")
	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = mac.Write([]byte(query.Encode()))
	if !hmac.Equal([]byte(signature), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		t.Error("invalid request signature")
	}
	if query.Get("recvWindow") != "5000" || query.Get("timestamp") == "" {
		t.Error("missing recvWindow or timestamp")
	}
}

func checkParams(t *testing.T, r *http.Request, want map[string]string) {
	t.Helper()
	for name, value := range want {
		if got := r.URL.Query().Get(name); got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
}

func checkDecimal(t *testing.T, got decimal.Decimal, want string) {
	t.Helper()
	if !got.Equal(decimal.RequireFromString(want)) {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNew(t *testing.T) {
	for _, test := range []struct {
		env    Env
		market Market
		url    string
	}{
		{"mainnet", "spot", "https://api.binance.com"},
		{"testnet", "spot", "https://testnet.binance.vision"},
		{"mainnet", "futures", "https://fapi.binance.com"},
		{"testnet", "futures", "https://testnet.binancefuture.com"},
	} {
		t.Run(string(test.env)+"/"+string(test.market), func(t *testing.T) {
			client, err := New(Config{Env: test.env, Market: test.market})
			if err != nil {
				t.Fatal(err)
			}
			if client.baseURL != test.url || client.http == nil || client.now == nil {
				t.Fatalf("incorrect defaults: %+v", client)
			}
		})
	}
	for _, cfg := range []Config{
		{Env: "bad", Market: "spot"}, {Env: "testnet", Market: "bad"},
		{Env: "testnet", Market: "spot", BaseURL: "://invalid"},
		{Env: "testnet", Market: "spot", BaseURL: "ftp://example.com"},
		{Env: "testnet", Market: "spot", BaseURL: "https://user:password@example.com"},
	} {
		if _, err := New(cfg); err == nil {
			t.Errorf("New accepted invalid config")
		}
	}
}

func TestSpotAccountAndPrice(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		switch r.URL.Path {
		case "/api/v3/account":
			checkSigned(t, r)
			checkParams(t, r, map[string]string{"timestamp": "1700000000000"})
			fmt.Fprint(w, `{"balances":[{"asset":"BTC","free":"1","locked":"2"},{"asset":"USDT","free":"100.15","locked":"25.2"}]}`)
		case "/api/v3/ticker/price":
			checkParams(t, r, map[string]string{"symbol": "BTCUSDT"})
			if r.URL.Query().Has("signature") || r.Header.Get("X-MBX-APIKEY") != "" {
				t.Error("public request contains credentials")
			}
			fmt.Fprint(w, `{"symbol":"BTCUSDT","price":"65000.12345678"}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	ctx := context.Background()
	if err := client.VerifyKeys(ctx); err != nil {
		t.Fatal(err)
	}
	balance, err := client.Balance(ctx, "usdt")
	if err != nil {
		t.Fatal(err)
	}
	checkDecimal(t, balance.Free, "100.15")
	checkDecimal(t, balance.Total, "125.35")
	missing, err := client.Balance(ctx, "MISSING")
	if err != nil || !missing.Total.IsZero() || !missing.Free.IsZero() {
		t.Fatalf("missing balance: %+v, %v", missing, err)
	}
	price, err := client.Price(ctx, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	checkDecimal(t, price, "65000.12345678")
	if err := client.SetLeverage(ctx, "BTCUSDT", 10); err != nil {
		t.Fatal(err)
	}
}

func TestFuturesAccountAndLeverage(t *testing.T) {
	client := newTestClient(t, "futures", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/ticker/price" {
			checkSigned(t, r)
		}
		switch r.URL.Path {
		case "/fapi/v2/account":
			fmt.Fprint(w, `{"totalWalletBalance":"125"}`)
		case "/fapi/v2/balance":
			fmt.Fprint(w, `[{"asset":"USDT","balance":"125.35","availableBalance":"100.15"},{"asset":"BTC","walletBalance":"2","availableBalance":"1.5"}]`)
		case "/fapi/v1/leverage":
			if r.Method != http.MethodPost {
				t.Errorf("method = %s", r.Method)
			}
			checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "leverage": "12"})
			fmt.Fprint(w, `{"leverage":12}`)
		case "/fapi/v1/ticker/price":
			checkParams(t, r, map[string]string{"symbol": "BTCUSDT"})
			fmt.Fprint(w, `{"price":"65001.2"}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	ctx := context.Background()
	if err := client.VerifyKeys(ctx); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ asset, free, total string }{{"USDT", "100.15", "125.35"}, {"BTC", "1.5", "2"}, {"MISSING", "0", "0"}} {
		balance, err := client.Balance(ctx, test.asset)
		if err != nil {
			t.Fatal(err)
		}
		checkDecimal(t, balance.Free, test.free)
		checkDecimal(t, balance.Total, test.total)
	}
	if err := client.SetLeverage(ctx, "BTCUSDT", 12); err != nil {
		t.Fatal(err)
	}
	price, err := client.Price(ctx, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	checkDecimal(t, price, "65001.2")
}

const exchangeInfo = `{"symbols":[{"symbol":"BTCUSDT","filters":[
	{"filterType":"MARKET_LOT_SIZE","stepSize":"0.005","minQty":"0.01"},
	{"filterType":"LOT_SIZE","stepSize":"0.001","minQty":"0.001"},
	{"filterType":"PRICE_FILTER","tickSize":"0.1"},
	{"filterType":"MIN_NOTIONAL","minNotional":"5"},
	{"filterType":"NOTIONAL","minNotional":"10"}]},
	{"symbol":"ETHUSDT","filters":[{"filterType":"LOT_SIZE","stepSize":"0.01","minQty":"0.02"},
	{"filterType":"MARKET_LOT_SIZE","stepSize":"0","minQty":"0"},
	{"filterType":"MIN_NOTIONAL","notional":"20"}]}]}`

func TestFiltersCacheAndRounding(t *testing.T) {
	for _, market := range []Market{"spot", "futures"} {
		t.Run(string(market), func(t *testing.T) {
			var calls atomic.Int32
			client := newTestClient(t, market, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				prefix := "/api/v3/"
				if market == "futures" {
					prefix = "/fapi/v1/"
				}
				if r.URL.Path != prefix+"exchangeInfo" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				fmt.Fprint(w, exchangeInfo)
			})
			now := testNow
			client.now = func() time.Time { return now }
			ctx := context.Background()
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					f, err := client.Filters(ctx, "BTCUSDT")
					if err != nil {
						t.Error(err)
						return
					}
					checkDecimal(t, f.StepSize, "0.005")
					checkDecimal(t, f.TickSize, "0.1")
					checkDecimal(t, f.MinQty, "0.01")
					checkDecimal(t, f.MinNotional, "10")
					checkDecimal(t, f.RoundQuantity(decimal.RequireFromString("1.234999")), "1.23")
					checkDecimal(t, f.RoundPrice(decimal.RequireFromString("123.49")), "123.4")
				}()
			}
			wg.Wait()
			f, err := client.Filters(ctx, "ETHUSDT")
			if err != nil {
				t.Fatal(err)
			}
			checkDecimal(t, f.StepSize, "0.01")
			checkDecimal(t, f.MinQty, "0.02")
			checkDecimal(t, f.MinNotional, "20")
			if _, err := client.Filters(ctx, "MISSING"); !errors.Is(err, ErrRejected) {
				t.Fatalf("unknown symbol error = %v", err)
			}
			now = now.Add(filterTTL - time.Nanosecond)
			if _, err := client.Filters(ctx, "BTCUSDT"); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 {
				t.Fatalf("cache made %d requests", calls.Load())
			}
			now = now.Add(time.Nanosecond)
			if _, err := client.Filters(ctx, "BTCUSDT"); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 2 {
				t.Fatalf("cache did not expire: %d requests", calls.Load())
			}
		})
	}
}

func TestFloorPrecision(t *testing.T) {
	f := Filters{StepSize: decimal.RequireFromString("0.00000001"), TickSize: decimal.RequireFromString("0.05")}
	checkDecimal(t, f.RoundQuantity(decimal.RequireFromString("1.99999999999999999999999999999")), "1.99999999")
	checkDecimal(t, f.RoundQuantity(decimal.RequireFromString("0.000000009")), "0")
	checkDecimal(t, f.RoundQuantity(decimal.RequireFromString("-0.000000009")), "-0.00000001")
	checkDecimal(t, f.RoundPrice(decimal.RequireFromString("1.249")), "1.2")
	checkDecimal(t, f.RoundPrice(decimal.RequireFromString("1.25")), "1.25")
	checkDecimal(t, (Filters{}).RoundQuantity(decimal.RequireFromString("1.123")), "1.123")
	checkDecimal(t, (Filters{}).RoundPrice(decimal.RequireFromString("1.123")), "1.123")
}

func TestErrorMapping(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   error
		code   int
	}{
		{"invalid key format", 400, `{"code":-2014,"msg":"bad key"}`, ErrInvalidKeys, 0},
		{"invalid permissions", 400, `{"code":-2015,"msg":"bad permissions"}`, ErrInvalidKeys, 0},
		{"invalid signature", 400, `{"code":-1022,"msg":"bad signature"}`, ErrInvalidKeys, 0},
		{"unauthorized", 401, `not JSON`, ErrInvalidKeys, 0},
		{"rejected", 400, `{"code":-1013,"msg":"Filter failure: LOT_SIZE"}`, ErrRejected, -1013},
		{"JSON error on success", 200, `{"code":-1111,"msg":"precision"}`, ErrRejected, -1111},
		{"rate limited", 429, `{"code":-1003,"msg":"too many requests"}`, ErrRejected, -1003},
		{"non JSON rejection", 403, `<html>forbidden</html>`, ErrRejected, 0},
		{"unavailable", 503, `{"code":-1000,"msg":"internal"}`, ErrUnavailable, 0},
		{"server auth error", 500, `{"code":-2015,"msg":"internal"}`, ErrUnavailable, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
				checkSigned(t, r)
				w.WriteHeader(test.status)
				fmt.Fprint(w, test.body)
			})
			err := client.VerifyKeys(context.Background())
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if test.want == ErrRejected {
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.HTTPStatus != test.status || apiErr.Code != test.code {
					t.Fatalf("incorrect APIError: %v", err)
				}
			}
		})
	}
}

func TestErrorRedaction(t *testing.T) {
	var signature string
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		signature = r.URL.Query().Get("signature")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": -1013,
			"msg": testKey + " " + testSecret + " " + signature + " " + r.URL.String()})
	})
	err := client.VerifyKeys(context.Background())
	if !errors.Is(err, ErrRejected) {
		t.Fatal(err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected APIError")
	}
	for _, sensitive := range []string{testKey, testSecret, signature} {
		if strings.Contains(err.Error(), sensitive) || strings.Contains(apiErr.Msg, sensitive) {
			t.Error("sensitive data in error")
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTransportErrorsRedacted(t *testing.T) {
	client, err := New(Config{Env: "testnet", Market: "spot", Credentials: Credentials{testKey, testSecret},
		HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("transport: %s %s %s", r.URL, testKey, testSecret)
		})}})
	if err != nil {
		t.Fatal(err)
	}
	err = client.VerifyKeys(context.Background())
	if !errors.Is(err, ErrUnavailable) || strings.Contains(err.Error(), testKey) || strings.Contains(err.Error(), testSecret) || strings.Contains(err.Error(), "signature") {
		t.Fatalf("transport error not sanitized: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.VerifyKeys(ctx); !errors.Is(err, ErrUnavailable) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
	client.credentials = Credentials{}
	if err := client.VerifyKeys(context.Background()); !errors.Is(err, ErrInvalidKeys) {
		t.Fatalf("missing keys error = %v", err)
	}
}

func TestServerTimeRetry(t *testing.T) {
	for _, market := range []Market{"spot", "futures"} {
		for _, alwaysReject := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/reject=%t", market, alwaysReject), func(t *testing.T) {
				var accountCalls, timeCalls int
				client := newTestClient(t, market, func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.URL.Path, "/time") {
						timeCalls++
						wantPath := "/api/v3/time"
						if market == "futures" {
							wantPath = "/fapi/v1/time"
						}
						if r.URL.Path != wantPath || r.URL.Query().Has("signature") {
							t.Error("invalid time request")
						}
						fmt.Fprint(w, `{"serverTime":1700000123456}`)
						return
					}
					checkSigned(t, r)
					accountCalls++
					if accountCalls == 1 {
						checkParams(t, r, map[string]string{"timestamp": "1700000000000"})
					} else {
						checkParams(t, r, map[string]string{"timestamp": "1700000123456"})
					}
					if accountCalls == 1 || alwaysReject {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, `{"code":-1021,"msg":"timestamp outside recvWindow"}`)
						return
					}
					fmt.Fprint(w, `{}`)
				})
				err := client.VerifyKeys(context.Background())
				if alwaysReject {
					if !errors.Is(err, ErrRejected) {
						t.Fatal(err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if accountCalls != 2 || timeCalls != 1 {
					t.Fatalf("calls: account=%d time=%d", accountCalls, timeCalls)
				}
				if !alwaysReject {
					if err := client.VerifyKeys(context.Background()); err != nil {
						t.Fatal(err)
					}
					if timeCalls != 1 {
						t.Error("offset was not retained")
					}
				}
			})
		}
	}
}

func TestMalformedResponse(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"price":"not-a-decimal"}`)
	})
	if _, err := client.Price(context.Background(), "BTCUSDT"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("malformed response error = %v", err)
	}
}

func TestVerifyKeysTradingPermission(t *testing.T) {
	for _, market := range []Market{"spot", "futures"} {
		t.Run(string(market), func(t *testing.T) {
			client := newTestClient(t, market, func(w http.ResponseWriter, r *http.Request) {
				checkSigned(t, r)
				fmt.Fprint(w, `{"canTrade":false}`)
			})
			if err := client.VerifyKeys(context.Background()); !errors.Is(err, ErrInvalidKeys) {
				t.Fatalf("disabled trading permission error = %v", err)
			}
		})
	}
}

func TestRedirectDoesNotForwardCredentials(t *testing.T) {
	var forwarded atomic.Bool
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded.Store(true)
	}))
	defer destination.Close()
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	})
	if err := client.VerifyKeys(context.Background()); !errors.Is(err, ErrRejected) {
		t.Fatalf("redirect error = %v", err)
	}
	if forwarded.Load() {
		t.Error("followed redirect with credentials")
	}
}

func TestReadFailure(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		fmt.Fprint(w, `{}`)
	})
	if err := client.VerifyKeys(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("read error = %v", err)
	}
}

func TestTimeRefreshFailure(t *testing.T) {
	for _, response := range []string{`{"serverTime":0}`, `invalid JSON`} {
		t.Run(response, func(t *testing.T) {
			client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v3/time" {
					fmt.Fprint(w, response)
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"code":-1021,"msg":"timestamp"}`)
			})
			if err := client.VerifyKeys(context.Background()); !errors.Is(err, ErrUnavailable) {
				t.Fatalf("time refresh error = %v", err)
			}
		})
	}
}

func TestFiltersFailureIsNotCached(t *testing.T) {
	calls := 0
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, exchangeInfo)
	})
	if _, err := client.Filters(context.Background(), "BTCUSDT"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("first error = %v", err)
	}
	if _, err := client.Filters(context.Background(), "BTCUSDT"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("requests = %d", calls)
	}
}
