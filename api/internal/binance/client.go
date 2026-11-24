// Package binance implements the Binance spot and USDT-M futures REST APIs.
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
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Env string
type Market string

type Credentials struct{ APIKey, Secret string }

type Config struct {
	Env         Env
	Market      Market
	Credentials Credentials
	BaseURL     string
	HTTP        *http.Client
	Now         func() time.Time
}

type Balance struct{ Free, Total decimal.Decimal }

type OrderRequest struct {
	Symbol, Side  string
	Quantity      decimal.Decimal
	ReduceOnly    bool
	ClientOrderID string
}

type StopRequest struct {
	Symbol, Side        string
	Quantity, StopPrice decimal.Decimal
	ClientOrderID       string
}

type OrderResult struct {
	OrderID                    int64
	ClientOrderID, Status      string
	ExecutedQty, AvgPrice, Fee decimal.Decimal
	FeeAsset                   string
}

type Filters struct{ StepSize, TickSize, MinQty, MinNotional decimal.Decimal }

func (f Filters) RoundQuantity(q decimal.Decimal) decimal.Decimal { return floor(q, f.StepSize) }
func (f Filters) RoundPrice(p decimal.Decimal) decimal.Decimal    { return floor(p, f.TickSize) }

func floor(value, increment decimal.Decimal) decimal.Decimal {
	if !increment.IsPositive() {
		return value
	}
	// QuoRem avoids rounding a ratio up at decimal.DivisionPrecision.
	quotient, remainder := value.QuoRem(increment, 0)
	if remainder.IsNegative() {
		quotient = quotient.Sub(decimal.NewFromInt(1))
	}
	return quotient.Mul(increment)
}

var (
	ErrInvalidKeys = errors.New("binance: invalid API keys or permissions")
	ErrRejected    = errors.New("binance: request rejected")
	ErrUnavailable = errors.New("binance: exchange unavailable")
)

type APIError struct {
	HTTPStatus, Code int
	Msg              string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("binance: request rejected (HTTP %d, code %d): %s", e.HTTPStatus, e.Code, e.Msg)
}

func (e *APIError) Unwrap() error { return ErrRejected }

const filterTTL = 5 * time.Minute

type Client struct {
	market      Market
	credentials Credentials
	baseURL     string
	http        *http.Client
	now         func() time.Time
	timeMu      sync.RWMutex
	offset      int64
	filterMu    sync.Mutex
	filters     map[string]Filters
	filterUntil time.Time
}

func New(cfg Config) (*Client, error) {
	if cfg.Env != "testnet" && cfg.Env != "mainnet" {
		return nil, errors.New("binance: invalid environment")
	}
	if cfg.Market != "spot" && cfg.Market != "futures" {
		return nil, errors.New("binance: invalid market")
	}
	base := cfg.BaseURL
	if base == "" {
		if cfg.Market == "futures" {
			base = "https://fapi.binance.com"
			if cfg.Env == "testnet" {
				base = "https://testnet.binancefuture.com"
			}
		} else {
			base = "https://api.binance.com"
			if cfg.Env == "testnet" {
				base = "https://testnet.binance.vision"
			}
		}
	}
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("binance: invalid base URL")
	}
	if cfg.HTTP == nil {
		cfg.HTTP = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Client{market: cfg.Market, credentials: cfg.Credentials, baseURL: strings.TrimRight(base, "/"), http: cfg.HTTP, now: cfg.Now}, nil
}

func (c *Client) path(resource string) string {
	if c.market == "futures" {
		return "/fapi/v1/" + resource
	}
	return "/api/v3/" + resource
}

func (c *Client) request(ctx context.Context, method, path string, params url.Values, signed bool, result any) error {
	for attempt := 0; attempt < 2; attempt++ {
		err := c.requestOnce(ctx, method, path, params, signed, result)
		var apiErr *APIError
		if !signed || attempt != 0 || !errors.As(err, &apiErr) || apiErr.Code != -1021 {
			return err
		}
		var response struct {
			ServerTime int64 `json:"serverTime"`
		}
		if err := c.requestOnce(ctx, http.MethodGet, c.path("time"), nil, false, &response); err != nil {
			return err
		}
		if response.ServerTime <= 0 {
			return ErrUnavailable
		}
		c.timeMu.Lock()
		c.offset = response.ServerTime - c.now().UnixMilli()
		c.timeMu.Unlock()
	}
	return ErrUnavailable
}

func (c *Client) requestOnce(ctx context.Context, method, path string, params url.Values, signed bool, result any) error {
	values := make(url.Values, len(params)+3)
	for key, entries := range params {
		values[key] = append([]string(nil), entries...)
	}
	signature := ""
	if signed {
		if c.credentials.APIKey == "" || c.credentials.Secret == "" {
			return ErrInvalidKeys
		}
		c.timeMu.RLock()
		timestamp := c.now().UnixMilli() + c.offset
		c.timeMu.RUnlock()
		values.Set("timestamp", strconv.FormatInt(timestamp, 10))
		values.Set("recvWindow", "5000")
		mac := hmac.New(sha256.New, []byte(c.credentials.Secret))
		_, _ = mac.Write([]byte(values.Encode()))
		signature = hex.EncodeToString(mac.Sum(nil))
		values.Set("signature", signature)
	}
	endpoint := c.baseURL + path
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return ErrUnavailable
	}
	if signed {
		req.Header.Set("X-MBX-APIKEY", c.credentials.APIKey)
	}
	// Do not follow redirects carrying signed URLs or API key headers.
	client := *c.http
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(req)
	if err != nil {
		// Transport errors can embed the full signed URL; never return them.
		if ctx.Err() != nil {
			return errors.Join(ErrUnavailable, ctx.Err())
		}
		return ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		return ErrInvalidKeys
	}
	if response.StatusCode >= 500 {
		return ErrUnavailable
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return ErrUnavailable
	}
	var apiErr struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	_ = json.Unmarshal(body, &apiErr)
	if apiErr.Code == -2014 || apiErr.Code == -2015 || apiErr.Code == -1022 {
		return ErrInvalidKeys
	}
	if apiErr.Code != 0 || response.StatusCode < 200 || response.StatusCode >= 300 {
		msg := apiErr.Msg
		for _, sensitive := range []string{signature, c.credentials.APIKey, c.credentials.Secret} {
			if sensitive != "" {
				msg = strings.ReplaceAll(msg, sensitive, "[redacted]")
			}
		}
		return &APIError{HTTPStatus: response.StatusCode, Code: apiErr.Code, Msg: msg}
	}
	if result != nil && json.Unmarshal(body, result) != nil {
		return ErrUnavailable
	}
	return nil
}
