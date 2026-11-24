package trading

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/binance"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/shopspring/decimal"
)

type fakeLiveExchange struct {
	balance                     binance.Balance
	filters                     binance.Filters
	stopErr, errorBalance       error
	marketErr                   error
	marketRequests              []binance.OrderRequest
	stops                       []binance.StopRequest
	queries, cancels, leverages int
	stopResult                  binance.OrderResult
	afterCancel                 *binance.OrderResult
	recovered                   *binance.OrderResult
	marketFee                   *binance.OrderResult
}

func (f *fakeLiveExchange) VerifyKeys(context.Context) error { return nil }
func (f *fakeLiveExchange) Balance(context.Context, string) (binance.Balance, error) {
	return f.balance, f.errorBalance
}
func (f *fakeLiveExchange) Price(context.Context, string) (decimal.Decimal, error) {
	return decimal.NewFromInt(100), nil
}
func (f *fakeLiveExchange) Filters(context.Context, string) (binance.Filters, error) {
	return f.filters, nil
}
func (f *fakeLiveExchange) SetLeverage(context.Context, string, uint32) error {
	f.leverages++
	return nil
}
func (f *fakeLiveExchange) MarketOrder(_ context.Context, r binance.OrderRequest) (binance.OrderResult, error) {
	f.marketRequests = append(f.marketRequests, r)
	price := int64(101)
	if len(f.marketRequests) > 1 {
		price = 110
	}
	result := binance.OrderResult{OrderID: int64(len(f.marketRequests)), Status: "FILLED", AvgPrice: decimal.NewFromInt(price), ExecutedQty: r.Quantity, Fee: decimal.RequireFromString("0.23"), FeeAsset: "USDT"}
	if f.marketFee != nil {
		result.Fee, result.FeeAsset = f.marketFee.Fee, f.marketFee.FeeAsset
	}
	return result, f.marketErr
}
func (f *fakeLiveExchange) StopOrder(_ context.Context, r binance.StopRequest) (binance.OrderResult, error) {
	f.stops = append(f.stops, r)
	return binance.OrderResult{OrderID: 42, Status: "NEW"}, f.stopErr
}
func (f *fakeLiveExchange) CancelOrder(context.Context, string, int64) error {
	f.cancels++
	f.stopResult.Status = "CANCELED"
	if f.afterCancel != nil {
		f.stopResult = *f.afterCancel
	}
	return nil
}
func (f *fakeLiveExchange) QueryOrder(context.Context, string, int64) (binance.OrderResult, error) {
	f.queries++
	return f.stopResult, nil
}
func (f *fakeLiveExchange) QueryOrderByClientID(context.Context, string, string) (binance.OrderResult, error) {
	if f.recovered == nil {
		return binance.OrderResult{}, binance.ErrRejected
	}
	return *f.recovered, nil
}

type fakeLiveProvider struct {
	exchange *fakeLiveExchange
	env      string
}

func (f fakeLiveProvider) Client(context.Context, string) (accounts.Exchange, string, error) {
	return f.exchange, f.env, nil
}

type liveMemoryStore struct{ *memoryStore }
type liveMemoryTx struct {
	*memoryTx
	equity string
}

func (s liveMemoryStore) Begin(ctx context.Context, owner, id string) (Transaction, error) {
	tx, err := s.memoryStore.Begin(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	return &liveMemoryTx{memoryTx: tx.(*memoryTx)}, nil
}
func (tx *liveMemoryTx) IsLive() bool           { return true }
func (tx *liveMemoryTx) SetEquity(value string) { tx.equity = value }
func (tx *liveMemoryTx) UpdateStop(_ context.Context, p Position) error {
	tx.positions[p.ID] = p
	return nil
}
func (tx *liveMemoryTx) Close(ctx context.Context, e Execution, _ decimal.Decimal) error {
	return tx.memoryTx.Open(ctx, e)
}

func liveFixture(t *testing.T) (Service, *memoryStore, *fakeLiveExchange, *testEngine) {
	s, store, _, engine, _ := fixture(t)
	exchange := &fakeLiveExchange{balance: binance.Balance{Free: decimal.NewFromInt(5000), Total: decimal.NewFromInt(6000)},
		filters: binance.Filters{StepSize: decimal.RequireFromString("0.1"), TickSize: decimal.RequireFromString("0.01"), MinQty: decimal.RequireFromString("0.1"), MinNotional: decimal.NewFromInt(10)}, stopResult: binance.OrderResult{OrderID: 42, Status: "NEW"}}
	s.Store = liveMemoryStore{store}
	s.Live = fakeLiveProvider{exchange, "testnet"}
	return s, store, exchange, engine
}
func liveRequest() Request { return Request{AccountID: testAccount, Symbol: "BTCUSDT", Side: "long"} }

func TestLiveOpenAndCloseRealFills(t *testing.T) {
	s, store, exchange, engine := liveFixture(t)
	engine.size = "2.37"
	result, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	p := result.Position
	if p.EntryPrice != "101" || p.Quantity != "2.3" || result.Fill.Fee != "0.23" || p.ExchangeEntryOrderID == nil || p.ExchangeStop.Status != "placed" || *p.ExchangeStop.OrderID != "42" {
		t.Fatalf("real fill not stored: %+v %+v", p, result.Fill)
	}
	if engine.requests[0].AccountEquity != "6000" || exchange.stops[0].Side != "sell" || exchange.stops[0].Quantity.String() != "2.3" {
		t.Fatalf("risk/stop mismatch: %+v %+v", engine.requests[0], exchange.stops)
	}
	closed, err := s.Close(context.Background(), "owner", p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *closed.Position.ExitPrice != "110" || *closed.Position.RealizedPNL != "20.24" || closed.Position.ExchangeStop.Status != "cancelled" || exchange.cancels != 1 || len(exchange.marketRequests) != 2 || exchange.marketRequests[1].Side != "sell" {
		t.Fatalf("bad live settlement: %+v", closed)
	}
	if store.balance.String() != "1000" {
		t.Fatal("live settlement altered paper balance")
	}
}

func TestLiveStopFailureRetryAndEngineFallback(t *testing.T) {
	s, _, exchange, _ := liveFixture(t)
	exchange.stopErr = errors.New("signed URL must never escape")
	opened, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	if opened.Position.ExchangeStop.Status != "failed" || *opened.Position.ExchangeStop.Error == exchange.stopErr.Error() {
		t.Fatal("failed stop lost or raw error exposed")
	}
	if _, err = s.RetryStop(context.Background(), "other", opened.Position.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign retry: %v", err)
	}
	exchange.stopErr = nil
	p, err := s.RetryStop(context.Background(), "owner", opened.Position.ID)
	if err != nil || p.ExchangeStop.Status != "placed" {
		t.Fatalf("retry %+v %v", p, err)
	}
	if _, err = s.RetryStop(context.Background(), "owner", p.ID); !errors.Is(err, ErrStopState) {
		t.Fatalf("duplicate retry: %v", err)
	}
	if err = s.Stop(context.Background(), &enginev1.StopHit{PositionId: p.ID}); err != nil {
		t.Fatal(err)
	}
	if len(exchange.marketRequests) != 1 || exchange.queries != 1 {
		t.Fatal("engine market-closed protected position")
	}

	s, _, exchange, _ = liveFixture(t)
	exchange.stopErr = errors.New("rejected")
	opened, err = s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Stop(context.Background(), &enginev1.StopHit{PositionId: opened.Position.ID}); err != nil {
		t.Fatal(err)
	}
	if len(exchange.marketRequests) != 2 {
		t.Fatal("failed stop did not fall back to market close")
	}
}

func TestLiveReconcileFilledStopIdempotent(t *testing.T) {
	s, store, exchange, _ := liveFixture(t)
	opened, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	exchange.stopResult = binance.OrderResult{OrderID: 42, Status: "FILLED", ExecutedQty: decimal.NewFromInt(2), AvgPrice: decimal.NewFromInt(94), Fee: decimal.RequireFromString("0.11"), FeeAsset: "USDT"}
	for i := 0; i < 2; i++ {
		if err = s.Reconcile(context.Background(), opened.Position.ID); err != nil {
			t.Fatal(err)
		}
	}
	p := store.positions[opened.Position.ID]
	if p.Status != "closed" || *p.CloseReason != "exchange_stop" || *p.ExitPrice != "94" || *p.RealizedPNL != "-14.34" || p.ExchangeStop.Status != "filled" || len(exchange.marketRequests) != 1 {
		t.Fatalf("reconcile %+v", p)
	}
}

func TestLiveMainnetConfirmationAndMinimum(t *testing.T) {
	s, _, exchange, _ := liveFixture(t)
	s.Live = fakeLiveProvider{exchange, "mainnet"}
	if _, err := s.Open(context.Background(), "owner", liveRequest()); !errors.Is(err, ErrConfirmation) || len(exchange.marketRequests) != 0 {
		t.Fatalf("unconfirmed order: %v", err)
	}
	r := liveRequest()
	r.ConfirmLive = true
	if _, err := s.Open(context.Background(), "owner", r); err != nil {
		t.Fatal(err)
	}
	s, _, exchange, _ = liveFixture(t)
	exchange.filters.MinNotional = decimal.NewFromInt(1000)
	if _, err := s.Open(context.Background(), "owner", liveRequest()); !errors.Is(err, ErrMinimum) || len(exchange.marketRequests) != 0 {
		t.Fatalf("minimum: %v", err)
	}
}

func TestLiveFuturesLeverageAndReduceOnly(t *testing.T) {
	s, store, exchange, _ := liveFixture(t)
	store.market = "futures"
	r := liveRequest()
	leverage := uint32(5)
	r.Leverage = &leverage
	r.Side = "short"
	opened, err := s.Open(context.Background(), "owner", r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Close(context.Background(), "owner", opened.Position.ID); err != nil {
		t.Fatal(err)
	}
	if exchange.leverages != 1 || exchange.marketRequests[0].Side != "sell" || exchange.marketRequests[1].Side != "buy" || !exchange.marketRequests[1].ReduceOnly || exchange.stops[0].Side != "buy" {
		t.Fatal("futures routing mismatch")
	}
}

func TestLiveExchangeErrorHasNoMarketOrder(t *testing.T) {
	s, _, exchange, _ := liveFixture(t)
	exchange.errorBalance = binance.ErrUnavailable
	if _, err := s.Open(context.Background(), "owner", liveRequest()); !errors.Is(err, binance.ErrUnavailable) || len(exchange.marketRequests) != 0 {
		t.Fatal(err)
	}
}

func TestLiveStopFilledDuringCancelDoesNotMarketClose(t *testing.T) {
	s, _, exchange, _ := liveFixture(t)
	opened, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	exchange.afterCancel = &binance.OrderResult{OrderID: 42, Status: "FILLED", ExecutedQty: decimal.NewFromInt(2), AvgPrice: decimal.NewFromInt(94), Fee: decimal.RequireFromString("0.11"), FeeAsset: "USDT"}
	result, err := s.Close(context.Background(), "owner", opened.Position.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *result.Position.CloseReason != "exchange_stop" || len(exchange.marketRequests) != 1 {
		t.Fatal("stop cancellation race submitted a duplicate market order")
	}
}

func TestLiveCloseTransportErrorIsSafeToLog(t *testing.T) {
	s, _, exchange, _ := liveFixture(t)
	opened, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	exchange.marketErr = errors.New("https://exchange/order?signature=private")
	_, err = s.Close(context.Background(), "owner", opened.Position.ID)
	if err == nil || err.Error() == exchange.marketErr.Error() || !errors.Is(err, exchange.marketErr) {
		t.Fatal("transport cause not sanitized")
	}
}

func TestLiveBaseSymbolAndNoPaperExchange(t *testing.T) {
	s, _, exchange, _ := liveFixture(t)
	s.Exchange = nil
	r := liveRequest()
	r.Symbol = "BTC"
	opened, err := s.Open(context.Background(), "owner", r)
	if err != nil {
		t.Fatal(err)
	}
	if opened.Position.Symbol != "BTCUSDT" || exchange.marketRequests[0].Symbol != "BTCUSDT" {
		t.Fatal("base asset not normalized")
	}
	if _, err = s.Close(context.Background(), "owner", opened.Position.ID); err != nil {
		t.Fatal(err)
	}
}

func TestLiveHTTPErrorContract(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{ErrConfirmation, 400, "live_confirmation_required"},
		{ErrMinimum, 422, "exchange_minimum"},
		{ErrStopState, 409, "exchange_stop_not_failed"},
	} {
		w := httptest.NewRecorder()
		respondError(w, tc.err, Execution{})
		assertError(t, w, tc.status, tc.code)
	}
}

func TestLiveSpotBaseFeeShrinksHeldQuantity(t *testing.T) {
	s, _, exchange, engine := liveFixture(t)
	engine.size = "2.37"
	// Binance charges a spot buy's fee in the bought coin: 2.3 BTC bought, 0.0023 BTC fee.
	exchange.marketFee = &binance.OrderResult{Fee: decimal.RequireFromString("0.0023"), FeeAsset: "BTC"}
	result, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	// 2.2977 rounded down to the 0.1 step is 2.2; the fee is 0.0023 * 101 USDT.
	if result.Position.Quantity != "2.2" || exchange.stops[0].Quantity.String() != "2.2" || result.Fill.Fee != "0.2323" || result.Fill.FeeAsset != "USDT" {
		t.Fatalf("net quantity or quote fee wrong: %+v %+v %+v", result.Position, result.Fill, exchange.stops)
	}
}

func TestLiveLostOrderResponseIsRecovered(t *testing.T) {
	s, _, exchange, engine := liveFixture(t)
	engine.size = "1"
	exchange.marketErr = binance.ErrUnavailable
	exchange.recovered = &binance.OrderResult{OrderID: 77, Status: "FILLED", AvgPrice: decimal.NewFromInt(102), ExecutedQty: decimal.NewFromInt(1), Fee: decimal.RequireFromString("0.1"), FeeAsset: "USDT"}
	result, err := s.Open(context.Background(), "owner", liveRequest())
	if err != nil {
		t.Fatal(err)
	}
	if result.Position.EntryPrice != "102" || *result.Position.ExchangeEntryOrderID != 77 || result.Position.ExchangeStop.Status != "placed" {
		t.Fatalf("lost fill not recovered: %+v", result.Position)
	}
}

func TestLiveLostOrderNotFoundStaysAnError(t *testing.T) {
	s, _, exchange, engine := liveFixture(t)
	engine.size = "1"
	exchange.marketErr = binance.ErrUnavailable
	if _, err := s.Open(context.Background(), "owner", liveRequest()); !errors.Is(err, binance.ErrUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestLiveSpotShortNeverSells(t *testing.T) {
	s, _, exchange, engine := liveFixture(t)
	engine.size = "1"
	request := liveRequest()
	request.Side = "short"
	if _, err := s.Open(context.Background(), "owner", request); !errors.Is(err, ErrSpotShort) || len(exchange.marketRequests) != 0 {
		t.Fatalf("err = %v, orders = %d", err, len(exchange.marketRequests))
	}
}
