package binance

import (
	"context"
	"net/http"

	"github.com/shopspring/decimal"
)

func (c *Client) Filters(ctx context.Context, symbol string) (Filters, error) {
	c.filterMu.Lock()
	defer c.filterMu.Unlock()
	if c.filters != nil && c.now().Before(c.filterUntil) {
		return lookupFilters(c.filters, symbol)
	}
	var info struct {
		Symbols []struct {
			Symbol  string `json:"symbol"`
			Filters []struct {
				Type        string          `json:"filterType"`
				StepSize    decimal.Decimal `json:"stepSize"`
				TickSize    decimal.Decimal `json:"tickSize"`
				MinQty      decimal.Decimal `json:"minQty"`
				MinNotional decimal.Decimal `json:"minNotional"`
				Notional    decimal.Decimal `json:"notional"`
			} `json:"filters"`
		} `json:"symbols"`
	}
	if err := c.request(ctx, http.MethodGet, c.path("exchangeInfo"), nil, false, &info); err != nil {
		return Filters{}, err
	}
	filters := make(map[string]Filters, len(info.Symbols))
	for _, symbol := range info.Symbols {
		var f, market Filters
		for _, rule := range symbol.Filters {
			switch rule.Type {
			case "LOT_SIZE":
				f.StepSize, f.MinQty = rule.StepSize, rule.MinQty
			case "MARKET_LOT_SIZE":
				market.StepSize, market.MinQty = rule.StepSize, rule.MinQty
			case "PRICE_FILTER":
				f.TickSize = rule.TickSize
			case "MIN_NOTIONAL", "NOTIONAL":
				minimum := rule.MinNotional
				if minimum.IsZero() {
					minimum = rule.Notional
				}
				if minimum.GreaterThan(f.MinNotional) {
					f.MinNotional = minimum
				}
			}
		}
		// Some spot symbols disable MARKET_LOT_SIZE with zero values.
		if market.StepSize.IsPositive() {
			f.StepSize = market.StepSize
		}
		if market.MinQty.GreaterThan(f.MinQty) {
			f.MinQty = market.MinQty
		}
		filters[symbol.Symbol] = f
	}
	c.filters, c.filterUntil = filters, c.now().Add(filterTTL)
	return lookupFilters(filters, symbol)
}

func lookupFilters(filters map[string]Filters, symbol string) (Filters, error) {
	if f, ok := filters[symbol]; ok {
		return f, nil
	}
	return Filters{}, &APIError{Code: -1121, Msg: "unknown symbol"}
}
