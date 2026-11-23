// Package trading owns order execution and settlement. Risk rules remain in the engine.
package trading

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/shopspring/decimal"
)

var (
	ErrNotFound          = errors.New("position not found")
	ErrClosed            = errors.New("position closed")
	ErrBlocked           = errors.New("risk blocked")
	ErrQuantity          = errors.New("quantity exceeds suggestion")
	ErrBalance           = errors.New("insufficient balance")
	ErrInput             = errors.New("invalid input")
	ErrPrice             = errors.New("price unavailable")
	ErrUnavailable       = errors.New("trading unavailable")
	ErrProfileNotAllowed = errors.New("client risk profile not allowed")
)

type Order struct {
	ID         string    `json:"id"`
	AccountID  string    `json:"account_id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Kind       string    `json:"kind"`
	Purpose    string    `json:"purpose"`
	Quantity   string    `json:"quantity"`
	Leverage   *uint32   `json:"leverage"`
	Stop       *string   `json:"stop"`
	Status     string    `json:"status"`
	PositionID *string   `json:"position_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Fill struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Price     string    `json:"price"`
	Quantity  string    `json:"quantity"`
	Fee       string    `json:"fee"`
	FeeAsset  string    `json:"fee_asset"`
	CreatedAt time.Time `json:"created_at"`
}

type Position struct {
	ID            string     `json:"id"`
	AccountID     string     `json:"account_id"`
	Symbol        string     `json:"symbol"`
	Side          string     `json:"side"`
	Quantity      string     `json:"quantity"`
	EntryPrice    string     `json:"entry_price"`
	Stop          string     `json:"stop"`
	Leverage      *uint32    `json:"leverage"`
	CoinClass     string     `json:"coin_class"`
	InitialRisk   string     `json:"initial_risk"`
	Status        string     `json:"status"`
	OpenedAt      time.Time  `json:"opened_at"`
	ClosedAt      *time.Time `json:"closed_at"`
	ExitPrice     *string    `json:"exit_price"`
	RealizedPNL   *string    `json:"realized_pnl"`
	CloseReason   *string    `json:"close_reason"`
	LastPrice     *string    `json:"last_price,omitempty"`
	UnrealizedPNL *string    `json:"unrealized_pnl,omitempty"`
}

// Execution is persisted atomically. Check is encoded with protobuf JSON by HTTP.
type Execution struct {
	Order    Order                   `json:"order"`
	Fill     Fill                    `json:"fill"`
	Position Position                `json:"position"`
	Check    *enginev1.CheckResponse `json:"-"`
}

type Request struct {
	AccountID   string          `json:"account_id"`
	Symbol      string          `json:"symbol"`
	Side        string          `json:"side"`
	ManualStop  *string         `json:"manual_stop"`
	RiskProfile json.RawMessage `json:"risk_profile"`
	Leverage    *uint32         `json:"leverage"`
	Quantity    *string         `json:"quantity"`
}

func (r Request) riskRequest(price string) risk.Request {
	return risk.Request{AccountID: r.AccountID, Symbol: r.Symbol, Side: r.Side, EntryPrice: price, ManualStop: r.ManualStop, Leverage: r.Leverage}
}

type Store interface {
	Begin(context.Context, string, string) (Transaction, error)
	// A nil owner is reserved for the internal stop consumer.
	PositionAccount(context.Context, *string, string) (string, string, error)
	Positions(context.Context, string, string, string) ([]Position, error)
}

type Transaction interface {
	Account() risk.Account
	Snapshot(context.Context, string) (risk.Account, error)
	Position(context.Context, string) (Position, error)
	OpeningFee(context.Context, string) (decimal.Decimal, error)
	Capital(context.Context) ([]Capital, error)
	Reject(context.Context, Order) error
	Open(context.Context, Execution) error
	Close(context.Context, Execution, decimal.Decimal) error
	RecordCheck(context.Context, risk.Request, *enginev1.CheckResponse, string) error
	Commit(context.Context) error
	Rollback(context.Context) error
}

type Service struct {
	Store     Store
	Exchange  Exchange
	Risk      risk.Service
	Now       func() time.Time
	Snapshots interface {
		Snapshot(context.Context, string) error
	}
	Log *slog.Logger
}

func (s Service) snapshot(ctx context.Context, accountID string) {
	if s.Snapshots == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.Snapshots.Snapshot(ctx, accountID); err != nil {
		log := s.Log
		if log == nil {
			log = slog.Default()
		}
		log.Warn("equity snapshot skipped after trade", "account_id", accountID, "err", err)
	}
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func newID() string {
	var b [16]byte
	// crypto/rand.Read is guaranteed to fill the buffer or terminate the process.
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (s Service) Open(ctx context.Context, owner string, body Request) (Execution, error) {
	var result Execution
	if len(body.RiskProfile) != 0 && string(body.RiskProfile) != "null" {
		return result, ErrProfileNotAllowed
	}
	if !risk.ValidID(body.AccountID) {
		return result, risk.ErrAccountNotFound
	}
	if owner == "" || !body.riskRequest("1").Valid() || (body.Quantity != nil && !risk.ValidQuantity(*body.Quantity)) {
		return result, ErrInput
	}
	if s.Store == nil || s.Exchange == nil {
		return result, ErrUnavailable
	}
	tx, err := s.Store.Begin(ctx, owner, body.AccountID)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	price, err := s.Exchange.Price(ctx, body.Symbol)
	if err != nil {
		return result, err
	}
	account, err := tx.Snapshot(ctx, body.Symbol)
	if err != nil {
		return result, err
	}
	check, err := s.Risk.Evaluate(ctx, body.riskRequest(price.String()), account)
	if err != nil {
		return result, err
	}
	result.Check = check
	quantity, err := decimal.NewFromString(check.PositionSize)
	if err != nil {
		return result, ErrUnavailable
	}
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
	order := Order{ID: newID(), AccountID: body.AccountID, Symbol: body.Symbol, Side: body.Side, Kind: "market", Purpose: "open",
		Quantity: quantity.String(), Leverage: &leverage, Stop: &stopText, Status: "rejected", CreatedAt: now}
	reject := func(reason error) (Execution, error) {
		// The schema requires positive rejected quantities, including a zero-size
		// risk failure. Record the minimum executable base unit in that case.
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
	suggested, err := decimal.NewFromString(check.PositionSize)
	if err != nil || !suggested.IsPositive() {
		return result, ErrUnavailable
	}
	if quantity.GreaterThan(suggested) {
		return reject(ErrQuantity)
	}
	if !quantity.IsPositive() || !quantity.Mod(decimal.RequireFromString("0.00000001")).IsZero() {
		return result, ErrInput
	}
	fillRequest := FillRequest{Market: account.Market, Symbol: body.Symbol, Side: body.Side, Purpose: "open", Price: price, Quantity: quantity}
	fee, err := s.Exchange.EstimateFee(ctx, fillRequest)
	if err != nil {
		return result, err
	}
	notional := price.Mul(quantity)
	capital, err := tx.Capital(ctx)
	if err != nil {
		return result, ErrUnavailable
	}
	balance, err := decimal.NewFromString(tx.Account().Equity)
	if err != nil {
		return result, ErrUnavailable
	}
	if !affordable(balance, account.Market, capital, notional, fee, leverage) {
		return reject(ErrBalance)
	}
	fill, err := s.Exchange.Execute(ctx, fillRequest)
	if err != nil {
		return result, err
	}
	id := newID()
	order.Status, order.PositionID = "filled", &id
	fill.ID, fill.OrderID, fill.CreatedAt = newID(), order.ID, now
	result.Order, result.Fill = order, fill
	result.Position = Position{ID: id, AccountID: body.AccountID, Symbol: body.Symbol, Side: body.Side,
		Quantity: quantity.String(), EntryPrice: price.String(), Stop: stopText, Leverage: &leverage, CoinClass: check.CoinClass,
		InitialRisk: price.Sub(stop).Abs().Mul(quantity).String(), Status: "open", OpenedAt: now}
	if err := tx.Open(ctx, result); err != nil {
		return Execution{}, err
	}
	if err := tx.RecordCheck(ctx, body.riskRequest(price.String()), check, order.ID); err != nil {
		return Execution{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Execution{}, err
	}
	s.snapshot(ctx, body.AccountID)
	return result, nil
}

func grossPNL(p Position, price decimal.Decimal) (decimal.Decimal, error) {
	entry, err := decimal.NewFromString(p.EntryPrice)
	if err != nil {
		return decimal.Zero, err
	}
	quantity, err := decimal.NewFromString(p.Quantity)
	if err != nil {
		return decimal.Zero, err
	}
	delta := price.Sub(entry).Mul(quantity)
	if p.Side == "short" {
		delta = delta.Neg()
	}
	return delta, nil
}

func (s Service) Close(ctx context.Context, owner, id string) (Execution, error) {
	if owner == "" {
		return Execution{}, ErrNotFound
	}
	return s.close(ctx, &owner, id, "manual")
}

// Stop is idempotent and uses the stored stop, never the ticker or event price.
func (s Service) Stop(ctx context.Context, hit *enginev1.StopHit) error {
	if hit == nil {
		return ErrInput
	}
	_, err := s.close(ctx, nil, hit.PositionId, "stop")
	if errors.Is(err, ErrClosed) || errors.Is(err, ErrNotFound) || errors.Is(err, risk.ErrAccountNotFound) {
		return nil
	}
	return err
}

func (s Service) close(ctx context.Context, owner *string, id, reason string) (Execution, error) {
	var result Execution
	if !risk.ValidID(id) {
		return result, ErrNotFound
	}
	if s.Store == nil || s.Exchange == nil {
		return result, ErrUnavailable
	}
	accountID, userID, err := s.Store.PositionAccount(ctx, owner, id)
	if err != nil {
		return result, err
	}
	tx, err := s.Store.Begin(ctx, userID, accountID)
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	p, err := tx.Position(ctx, id)
	if err != nil {
		return result, err
	}
	if p.Status != "open" {
		return result, ErrClosed
	}
	var price decimal.Decimal
	if reason == "stop" {
		price, err = decimal.NewFromString(p.Stop)
	} else {
		price, err = s.Exchange.Price(ctx, p.Symbol)
	}
	if err != nil {
		return result, err
	}
	quantity, err := decimal.NewFromString(p.Quantity)
	if err != nil {
		return result, err
	}
	fill, err := s.Exchange.Execute(ctx, FillRequest{Market: tx.Account().Market, Symbol: p.Symbol, Side: p.Side, Purpose: "close", Price: price, Quantity: quantity})
	if err != nil {
		return result, err
	}
	openingFee, err := tx.OpeningFee(ctx, id)
	if err != nil {
		return result, err
	}
	closingFee, err := decimal.NewFromString(fill.Fee)
	if err != nil {
		return result, err
	}
	pnl, err := grossPNL(p, price)
	if err != nil {
		return result, err
	}
	pnl = pnl.Sub(openingFee).Sub(closingFee)
	now, exit, net := s.now(), price.String(), pnl.String()
	p.Status, p.ClosedAt, p.ExitPrice, p.RealizedPNL, p.CloseReason = "closed", &now, &exit, &net, &reason
	stop := p.Stop
	order := Order{ID: newID(), AccountID: accountID, Symbol: p.Symbol, Side: p.Side, Kind: "market", Purpose: "close", Quantity: p.Quantity,
		Leverage: p.Leverage, Stop: &stop, Status: "filled", PositionID: &id, CreatedAt: now}
	fill.ID, fill.OrderID, fill.CreatedAt = newID(), order.ID, now
	result = Execution{Order: order, Fill: fill, Position: p}
	if err := tx.Close(ctx, result, pnl); err != nil {
		return Execution{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Execution{}, err
	}
	s.snapshot(ctx, accountID)
	return result, nil
}
