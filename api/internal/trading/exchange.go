package trading

import (
	"context"

	"github.com/batuhangozegu/metron/api/internal/market"
	"github.com/shopspring/decimal"
)

// Exchange is the execution boundary; a real exchange can replace Paper without
// moving ownership, risk authorization or settlement into the engine.
type Exchange interface {
	Price(context.Context, string) (decimal.Decimal, error)
	EstimateFee(context.Context, FillRequest) (decimal.Decimal, error)
	Execute(context.Context, FillRequest) (Fill, error)
}

type FillRequest struct {
	Market, Symbol, Side, Purpose string
	Price, Quantity               decimal.Decimal
}

type Paper struct{ Tickers market.TickerSource }

func (p Paper) Price(ctx context.Context, symbol string) (decimal.Decimal, error) {
	if p.Tickers == nil {
		return decimal.Zero, ErrPrice
	}
	values, err := p.Tickers.Tickers(ctx, []string{symbol})
	if err != nil {
		return decimal.Zero, ErrPrice
	}
	price, err := decimal.NewFromString(values[symbol]["price"])
	if err != nil || !price.IsPositive() {
		return decimal.Zero, ErrPrice
	}
	return price, nil
}

func (Paper) EstimateFee(_ context.Context, request FillRequest) (decimal.Decimal, error) {
	if !request.Price.IsPositive() || !request.Quantity.IsPositive() {
		return decimal.Zero, ErrInput
	}
	rate := decimal.RequireFromString("0.001")
	if request.Market == "futures" {
		rate = decimal.RequireFromString("0.0005")
	} else if request.Market != "spot" {
		return decimal.Zero, ErrInput
	}
	return request.Price.Mul(request.Quantity).Mul(rate), nil
}

func (p Paper) Execute(ctx context.Context, request FillRequest) (Fill, error) {
	fee, err := p.EstimateFee(ctx, request)
	if err != nil {
		return Fill{}, err
	}
	return Fill{Price: request.Price.String(), Quantity: request.Quantity.String(),
		Fee: fee.String(), FeeAsset: "USDT"}, nil
}
