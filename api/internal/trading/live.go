package trading

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/binance"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ExchangeStop struct {
	Status  string  `json:"status"`
	OrderID *string `json:"order_id"`
	Error   *string `json:"error"`
}
type LiveProvider interface {
	Client(context.Context, string) (accounts.Exchange, string, error)
}

// LiveExchange is satisfied by *binance.Client; execution does not depend on
// credential storage or the concrete HTTP implementation.
type LiveExchange interface {
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
type stopTransaction interface {
	UpdateStop(context.Context, Position) error
}

// Signed request URLs can occur in transport errors. Keep their causes for
// errors.Is, but never expose them through the stop consumer's logger.
type liveExecutionError struct{ cause error }

func (e liveExecutionError) Error() string { return "live exchange execution unavailable" }
func (e liveExecutionError) Unwrap() error { return e.cause }

func isLive(tx Transaction) bool   { v, ok := tx.(interface{ IsLive() bool }); return ok && v.IsLive() }
func (p *postgresTx) IsLive() bool { return p.live }
func stopOrderID(p Position) *string {
	if p.ExchangeStop == nil {
		return nil
	}
	return p.ExchangeStop.OrderID
}
func stopStatus(p Position) *string {
	if p.ExchangeStop == nil {
		return nil
	}
	return &p.ExchangeStop.Status
}
func stopError(p Position) *string {
	if p.ExchangeStop == nil {
		return nil
	}
	return p.ExchangeStop.Error
}
func pair(symbol string) string {
	if strings.HasSuffix(symbol, "USDT") {
		return symbol
	}
	return symbol + "USDT"
}
func opposite(side string) string {
	if side == "long" {
		return "sell"
	}
	return "buy"
}
func entrySide(side string) string {
	if side == "long" {
		return "buy"
	}
	return "sell"
}
func (s Service) logger() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}
func (s Service) liveClient(ctx context.Context, id string) (LiveExchange, string, error) {
	if s.Live == nil {
		return nil, "", ErrUnavailable
	}
	return s.Live.Client(ctx, id)
}
func liveFill(r binance.OrderResult) (Fill, error) {
	if !r.ExecutedQty.IsPositive() || !r.AvgPrice.IsPositive() {
		return Fill{}, ErrUnavailable
	}
	asset := r.FeeAsset
	if asset == "" {
		asset = "USDT"
	}
	return Fill{Price: r.AvgPrice.String(), Quantity: r.ExecutedQty.String(), Fee: r.Fee.String(), FeeAsset: asset}, nil
}

func (s Service) openLive(ctx context.Context, tx Transaction, body Request) (Execution, error) {
	var result Execution
	exchange, env, err := s.liveClient(ctx, body.AccountID)
	if err != nil {
		return result, err
	}
	if env == "mainnet" && !body.ConfirmLive {
		return result, ErrConfirmation
	}
	// A spot "short" would sell coins the user already holds on Binance.
	if body.Side == "short" && tx.Account().Market == "spot" {
		return result, ErrSpotShort
	}
	symbol := pair(body.Symbol)
	price, err := exchange.Price(ctx, symbol)
	if err != nil {
		return result, err
	}
	balance, err := exchange.Balance(ctx, "USDT")
	if err != nil {
		return result, err
	}
	// The transaction's activity aggregates must use the same exchange equity
	// supplied to the engine, including percentage-based discipline rules.
	if v, ok := tx.(interface{ SetEquity(string) }); ok {
		v.SetEquity(balance.Total.String())
	}
	account, err := tx.Snapshot(ctx, body.Symbol)
	if err != nil {
		return result, err
	}
	account.Equity = balance.Total.String()
	check, err := s.Risk.Evaluate(ctx, body.riskRequest(price.String()), account)
	if err != nil {
		return result, err
	}
	result.Check = check
	suggested, err := decimal.NewFromString(check.PositionSize)
	if err != nil {
		return result, ErrUnavailable
	}
	quantity := suggested
	if body.Quantity != nil {
		quantity, _ = decimal.NewFromString(*body.Quantity)
	}
	stop, err := decimal.NewFromString(check.Stop)
	if err != nil || !stop.IsPositive() {
		return result, ErrUnavailable
	}
	leverage := check.Leverage
	if leverage < 1 || leverage > 125 {
		return result, ErrUnavailable
	}
	now := s.now()
	stopText := stop.String()
	order := Order{ID: newID(), AccountID: body.AccountID, Symbol: body.Symbol, Side: body.Side, Kind: "market", Purpose: "open", Quantity: quantity.String(), Leverage: &leverage, Stop: &stopText, Status: "rejected", CreatedAt: now}
	reject := func(reason error) (Execution, error) {
		if !quantity.IsPositive() {
			order.Quantity = "0.00000001"
		}
		if err := tx.Reject(ctx, order); err != nil {
			return result, err
		}
		if err := tx.RecordCheck(ctx, body.riskRequest(price.String()), check, order.ID); err != nil {
			return result, err
		}
		if err := tx.Commit(ctx); err != nil {
			return result, err
		}
		return result, reason
	}
	blocked := !check.Allowed
	for _, rule := range check.Checks {
		if rule.GetStatus() == "fail" {
			blocked = true
		}
	}
	if blocked {
		return reject(ErrBlocked)
	}
	if quantity.GreaterThan(suggested) {
		return reject(ErrQuantity)
	}
	filters, err := exchange.Filters(ctx, symbol)
	if err != nil {
		return result, err
	}
	quantity = filters.RoundQuantity(quantity)
	stop = filters.RoundPrice(stop)
	stopText = stop.String()
	order.Stop = &stopText
	order.Quantity = quantity.String()
	if !quantity.IsPositive() || quantity.LessThan(filters.MinQty) || quantity.Mul(price).LessThan(filters.MinNotional) || !stop.IsPositive() {
		return reject(ErrMinimum)
	}
	required := quantity.Mul(price)
	if account.Market == "futures" {
		required = required.Div(decimal.NewFromInt(int64(leverage)))
	}
	if required.GreaterThan(balance.Free) {
		return reject(ErrBalance)
	}
	if account.Market == "futures" {
		if err = exchange.SetLeverage(ctx, symbol, leverage); err != nil {
			return result, err
		}
	}
	actual, err := exchange.MarketOrder(ctx, binance.OrderRequest{Symbol: symbol, Side: entrySide(body.Side), Quantity: quantity, ClientOrderID: order.ID})
	if errors.Is(err, binance.ErrUnavailable) {
		// The response was lost (timeout, 5xx) but the order may have filled.
		// Look it up by our client order id before giving up on it.
		lookup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		recovered, qerr := exchange.QueryOrderByClientID(lookup, symbol, order.ID)
		cancel()
		if qerr == nil && recovered.ExecutedQty.IsPositive() {
			actual, err = recovered, nil
		} else if qerr == nil || !errors.Is(qerr, binance.ErrRejected) {
			s.logger().Error("LIVE ORDER STATE UNKNOWN; CHECK THE EXCHANGE", "account_id", body.AccountID, "client_order_id", order.ID)
		}
	}
	if err != nil {
		return result, err
	}
	fill, err := liveFill(actual)
	if err != nil {
		s.logger().Error("LIVE MARKET FILL COULD NOT BE RECORDED", "account_id", body.AccountID, "order_id", actual.OrderID)
		return result, err
	}
	// Record fees in the quote asset. A spot buy pays its fee in the base coin,
	// so the coins actually held (and the stop/close size) are net of that fee.
	held := actual.ExecutedQty
	fee, feeAsset := actual.Fee, actual.FeeAsset
	base := strings.TrimSuffix(symbol, "USDT")
	switch {
	case !fee.IsPositive() || feeAsset == "" || feeAsset == "USDT":
	case feeAsset == base:
		if account.Market == "spot" {
			held = filters.RoundQuantity(held.Sub(fee))
		}
		fee = fee.Mul(actual.AvgPrice)
	default:
		if rate, e := exchange.Price(ctx, feeAsset+"USDT"); e == nil && rate.IsPositive() {
			fee = fee.Mul(rate)
		} else {
			s.logger().Warn("live fee asset could not be priced; fee recorded as zero", "asset", feeAsset)
			fee = decimal.Zero
		}
	}
	if !held.IsPositive() {
		s.logger().Error("LIVE FILL LEFT NO TRADABLE QUANTITY", "account_id", body.AccountID, "order_id", actual.OrderID)
		return result, ErrUnavailable
	}
	fill.Quantity, fill.Fee, fill.FeeAsset = held.String(), fee.String(), "USDT"
	id := newID()
	order.Status = "filled"
	order.PositionID = &id
	order.Quantity = fill.Quantity
	fill.ID, fill.OrderID, fill.CreatedAt = newID(), order.ID, now
	pending := "Exchange stop placement pending / Borsa stop emri bekleniyor"
	p := Position{ID: id, AccountID: body.AccountID, Symbol: body.Symbol, Side: body.Side, Quantity: fill.Quantity, EntryPrice: fill.Price, Stop: stopText, Leverage: &leverage, CoinClass: check.CoinClass,
		InitialRisk: actual.AvgPrice.Sub(stop).Abs().Mul(held).String(), Status: "open", OpenedAt: now, ExchangeEntryOrderID: &actual.OrderID, ExchangeStop: &ExchangeStop{Status: "failed", Error: &pending}}
	result.Order, result.Fill, result.Position = order, fill, p
	if err = tx.Open(ctx, result); err != nil {
		s.logger().Error("LIVE FILL DATABASE WRITE FAILED", "account_id", body.AccountID, "order_id", actual.OrderID)
		return result, err
	}
	if err = tx.RecordCheck(ctx, body.riskRequest(price.String()), check, order.ID); err != nil {
		return result, err
	}
	// Commit the real fill before placing the protective order. A placement or
	// request failure cannot roll the filled position out of local history.
	if err = tx.Commit(ctx); err != nil {
		s.logger().Error("LIVE FILL COMMIT FAILED", "account_id", body.AccountID, "order_id", actual.OrderID)
		return result, err
	}
	// The account owner is resolved internally after the original lock is released.
	retryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if updated, e := s.retryStop(retryCtx, nil, id); e == nil {
		result.Position = updated
	} else {
		s.logger().Error("LIVE POSITION HAS NO CONFIRMED EXCHANGE STOP", "position_id", id, "err", e)
	}
	s.snapshot(ctx, body.AccountID)
	return result, nil
}

func (s Service) placeStop(ctx context.Context, tx Transaction, p Position, exchange LiveExchange) (Position, error) {
	writer, ok := tx.(stopTransaction)
	if !ok {
		return p, ErrUnavailable
	}
	quantity, err := decimal.NewFromString(p.Quantity)
	if err != nil {
		return p, err
	}
	stop, err := decimal.NewFromString(p.Stop)
	if err != nil {
		return p, err
	}
	r, err := exchange.StopOrder(ctx, binance.StopRequest{Symbol: pair(p.Symbol), Side: opposite(p.Side), Quantity: quantity, StopPrice: stop, ClientOrderID: "s-" + strings.ReplaceAll(p.ID, "-", "")})
	p.ExchangeStop = &ExchangeStop{Status: "placed"}
	if err != nil {
		// Do not persist raw exchange errors: transport errors can contain signed URLs.
		message := "Exchange stop placement failed; retry required / Borsa stop emri başarısız; yeniden deneyin"
		p.ExchangeStop.Status = "failed"
		p.ExchangeStop.Error = &message
		s.logger().Error("LIVE EXCHANGE STOP PLACEMENT FAILED", "position_id", p.ID)
	} else {
		id := strconv.FormatInt(r.OrderID, 10)
		p.ExchangeStop.OrderID = &id
	}
	if e := writer.UpdateStop(ctx, p); e != nil {
		return p, e
	}
	if e := tx.Commit(ctx); e != nil {
		return p, e
	}
	return p, nil
}
func (s Service) RetryStop(ctx context.Context, owner, id string) (Position, error) {
	if owner == "" {
		return Position{}, ErrNotFound
	}
	return s.retryStop(ctx, &owner, id)
}
func (s Service) retryStop(ctx context.Context, owner *string, id string) (Position, error) {
	if !risk.ValidID(id) {
		return Position{}, ErrNotFound
	}
	if s.Store == nil {
		return Position{}, ErrUnavailable
	}
	account, user, err := s.Store.PositionAccount(ctx, owner, id)
	if err != nil {
		return Position{}, err
	}
	tx, err := s.Store.Begin(ctx, user, account)
	if err != nil {
		return Position{}, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	p, err := tx.Position(ctx, id)
	if err != nil {
		return p, err
	}
	if p.Status != "open" {
		return p, ErrClosed
	}
	if !isLive(tx) || p.ExchangeStop == nil || p.ExchangeStop.Status != "failed" {
		return p, ErrStopState
	}
	exchange, _, err := s.liveClient(ctx, account)
	if err != nil {
		return p, err
	}
	return s.placeStop(ctx, tx, p, exchange)
}

func (s Service) closeLive(ctx context.Context, tx Transaction, p Position, reason string) (Execution, error) {
	exchange, _, err := s.liveClient(ctx, p.AccountID)
	if err != nil {
		return Execution{}, err
	}
	var stopID int64
	if p.ExchangeStop != nil && p.ExchangeStop.OrderID != nil {
		stopID, err = strconv.ParseInt(*p.ExchangeStop.OrderID, 10, 64)
		if err != nil {
			return Execution{}, err
		}
	}
	if stopID != 0 && p.ExchangeStop.Status == "placed" {
		r, err := exchange.QueryOrder(ctx, pair(p.Symbol), stopID)
		if err != nil {
			return Execution{}, err
		}
		if r.Status == "FILLED" {
			return s.settleStop(ctx, tx, p, r)
		}
		if r.ExecutedQty.IsPositive() {
			return Execution{}, ErrUnavailable
		}
		if r.Status == "CANCELED" || r.Status == "EXPIRED" || r.Status == "EXPIRED_IN_MATCH" || r.Status == "REJECTED" {
			message := "Exchange stop is no longer active; retry required / Borsa stop emri etkin değil; yeniden deneyin"
			p.ExchangeStop = &ExchangeStop{Status: "failed", Error: &message}
			writer, ok := tx.(stopTransaction)
			if !ok {
				return Execution{}, ErrUnavailable
			}
			if err = writer.UpdateStop(ctx, p); err != nil {
				return Execution{}, err
			}
			if reason == "reconcile" {
				return Execution{Position: p}, tx.Commit(ctx)
			}
		} else {
			// Engine events only reconcile exchange-owned stops; never submit a duplicate close.
			if reason == "stop" || reason == "reconcile" {
				return Execution{Position: p}, nil
			}
			if err = exchange.CancelOrder(ctx, pair(p.Symbol), stopID); err != nil {
				return Execution{}, err
			}
			// A stop may have filled between the first query and cancel acknowledgement.
			r, err = exchange.QueryOrder(ctx, pair(p.Symbol), stopID)
			if err != nil {
				return Execution{}, err
			}
			if r.Status == "FILLED" {
				return s.settleStop(ctx, tx, p, r)
			}
			if r.ExecutedQty.IsPositive() || (r.Status != "CANCELED" && r.Status != "EXPIRED" && r.Status != "EXPIRED_IN_MATCH" && r.Status != "REJECTED") {
				return Execution{}, ErrUnavailable
			}
		}
	}
	if reason == "reconcile" {
		return Execution{Position: p}, nil
	}
	quantity, err := decimal.NewFromString(p.Quantity)
	if err != nil {
		return Execution{}, err
	}
	r, err := exchange.MarketOrder(ctx, binance.OrderRequest{Symbol: pair(p.Symbol), Side: opposite(p.Side), Quantity: quantity, ReduceOnly: tx.Account().Market == "futures", ClientOrderID: "c-" + strings.ReplaceAll(p.ID, "-", "")})
	if err != nil {
		return Execution{}, err
	}
	fill, err := liveFill(r)
	if err != nil {
		return Execution{}, err
	}
	if !r.ExecutedQty.Equal(quantity) {
		s.logger().Error("LIVE CLOSE PARTIAL FILL REQUIRES RECONCILIATION", "position_id", p.ID)
		return Execution{}, ErrUnavailable
	}
	if p.ExchangeStop != nil {
		p.ExchangeStop = &ExchangeStop{Status: "cancelled", OrderID: p.ExchangeStop.OrderID}
	}
	return s.settle(ctx, tx, p, fill, r.AvgPrice, reason)
}
func (s Service) settleStop(ctx context.Context, tx Transaction, p Position, r binance.OrderResult) (Execution, error) {
	fill, err := liveFill(r)
	if err != nil {
		return Execution{}, err
	}
	if !r.ExecutedQty.Equal(decimal.RequireFromString(p.Quantity)) {
		return Execution{}, ErrUnavailable
	}
	p.ExchangeStop = &ExchangeStop{Status: "filled", OrderID: p.ExchangeStop.OrderID}
	return s.settle(ctx, tx, p, fill, r.AvgPrice, "exchange_stop")
}
func (s Service) Reconcile(ctx context.Context, id string) error {
	_, err := s.close(ctx, nil, id, "reconcile")
	if errors.Is(err, ErrClosed) || errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}
func (s Service) RunReconciler(ctx context.Context) {
	source, ok := s.Store.(interface {
		OpenLivePositions(context.Context) ([]string, error)
	})
	if !ok {
		return
	}
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ids, err := source.OpenLivePositions(ctx)
			if err != nil {
				s.logger().Warn("live stop reconciliation unavailable")
				continue
			}
			for _, id := range ids {
				if ctx.Err() != nil {
					return
				}
				call, cancel := context.WithTimeout(ctx, 10*time.Second)
				err = s.Reconcile(call, id)
				cancel()
				if err != nil {
					s.logger().Warn("live stop reconcile failed", "position_id", id)
				}
			}
		}
	}
}

func loadTradingAccount(ctx context.Context, q risk.Querier, owner, id string, lock bool) (risk.Account, bool, error) {
	query := `SELECT COALESCE(paper_balance,0)::text,market,type FROM app.accounts WHERE user_id=$1::uuid AND id=$2::uuid AND quote_asset='USDT' AND type IN ('paper','live')`
	if lock {
		query += ` FOR UPDATE`
	}
	var a risk.Account
	var kind string
	err := q.QueryRow(ctx, query, owner, id).Scan(&a.Equity, &a.Market, &kind)
	if errors.Is(err, pgx.ErrNoRows) {
		err = risk.ErrAccountNotFound
	}
	if err == nil {
		var balance decimal.Decimal
		balance, err = decimal.NewFromString(a.Equity)
		a.Equity = balance.String()
	}
	return a, kind == "live", err
}
func (p *postgresTx) SetEquity(equity string) { p.account.Equity = equity }

// RiskAccounts adds live previews without changing the paper risk package.
type RiskAccounts struct {
	Store Postgres
	Live  LiveProvider
}

func (p RiskAccounts) RiskAccount(ctx context.Context, owner, id, symbol string) (risk.Account, error) {
	tx, err := p.Store.Begin(ctx, owner, id)
	if err != nil {
		return risk.Account{}, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	if isLive(tx) {
		if p.Live == nil {
			return risk.Account{}, ErrUnavailable
		}
		exchange, _, err := p.Live.Client(ctx, id)
		if err != nil {
			return risk.Account{}, err
		}
		balance, err := exchange.Balance(ctx, "USDT")
		if err != nil {
			return risk.Account{}, err
		}
		tx.(*postgresTx).SetEquity(balance.Total.String())
	}
	return tx.Snapshot(ctx, symbol)
}
func (p *postgresTx) UpdateStop(ctx context.Context, v Position) error {
	_, err := p.Exec(ctx, `UPDATE app.positions SET exchange_stop_order_id=$3::bigint,exchange_stop_status=$4,exchange_stop_error=$5 WHERE id=$1::uuid AND account_id=$2::uuid`, v.ID, p.accountID, stopOrderID(v), stopStatus(v), stopError(v))
	return err
}
func (p Postgres) OpenLivePositions(ctx context.Context) ([]string, error) {
	rows, err := p.Pool.Query(ctx, `SELECT p.id::text FROM app.positions p JOIN app.accounts a ON a.id=p.account_id WHERE a.type='live' AND p.status='open' AND p.exchange_stop_status='placed'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
