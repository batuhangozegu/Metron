package alerts

import (
	"context"
	"log/slog"
	"time"

	"github.com/batuhangozegu/metron/api/internal/market"
	"github.com/shopspring/decimal"
)

type Checker struct {
	Store   Store
	Tickers market.TickerSource
	Log     *slog.Logger
}

func (c Checker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
			if err := c.Check(checkCtx); err != nil && ctx.Err() == nil {
				c.logger().Warn("price alert check failed", "err", err)
			}
			cancel()
		}
	}
}

func (c Checker) logger() *slog.Logger {
	if c.Log != nil {
		return c.Log
	}
	return slog.Default()
}

func (c Checker) Check(ctx context.Context) error {
	if c.Store == nil || c.Tickers == nil {
		return nil
	}
	symbols, err := c.Store.ActiveSymbols(ctx)
	if err != nil || len(symbols) == 0 {
		return err
	}
	tickers, err := c.Tickers.Tickers(ctx, symbols)
	if err != nil {
		return err
	}
	for _, symbol := range symbols {
		price, err := decimal.NewFromString(tickers[symbol]["price"])
		if err != nil || !price.IsPositive() {
			continue
		}
		triggered, err := c.Store.Trigger(ctx, symbol, price)
		if err != nil {
			return err
		}
		for _, alert := range triggered {
			c.logger().Info("price alert triggered", "alert_id", alert.ID, "symbol", symbol, "price", price.String())
		}
	}
	return nil
}
