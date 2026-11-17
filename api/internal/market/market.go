// Package market serves the engine's market statistics and latest cached tickers.
package market

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// DefaultATRMultiplier is the system stop default until roadmap step 3 adds
// per-class risk rules. Stablecoins have no suggested stop.
const DefaultATRMultiplier = "1.5"

type Symbol struct {
	Symbol, BaseAsset, QuoteAsset, CoinClass string
	ATR, ATRPercent                          *string
}

type SymbolSource interface {
	Symbols(context.Context) ([]Symbol, error)
}

type TickerSource interface {
	Tickers(context.Context, []string) (map[string]map[string]string, error)
}

type Market struct {
	Symbol               string  `json:"symbol"`
	BaseAsset            string  `json:"base_asset"`
	QuoteAsset           string  `json:"quote_asset"`
	CoinClass            string  `json:"coin_class"`
	Price                *string `json:"price"`
	Change24hPercent     *string `json:"change_24h_percent"`
	QuoteVolume24h       *string `json:"quote_volume_24h"`
	PriceTime            *string `json:"price_time"`
	ATRPercent           *string `json:"atr_percent"`
	ATR                  *string `json:"atr"`
	SuggestedStopPercent *string `json:"suggested_stop_percent"`
}

type Response struct {
	AsOf    string   `json:"as_of"`
	Markets []Market `json:"markets"`
}

func Handler(symbols SymbolSource, tickers TickerSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()
		if symbols == nil {
			unavailable(w)
			return
		}
		rows, err := symbols.Symbols(ctx)
		if err != nil {
			unavailable(w)
			return
		}
		keys := make([]string, len(rows))
		for i, row := range rows {
			keys[i] = row.Symbol
		}
		var cached map[string]map[string]string
		if tickers != nil && len(keys) > 0 {
			// Redis is optional for a response: absent or failed hashes yield nulls.
			redisCtx, stop := context.WithTimeout(ctx, time.Second)
			cached, _ = tickers.Tickers(redisCtx, keys)
			stop()
		}
		result := Response{AsOf: time.Now().UTC().Format(time.RFC3339), Markets: make([]Market, 0, len(rows))}
		for _, row := range rows {
			result.Markets = append(result.Markets, present(row, cached[row.Symbol]))
		}
		_ = json.NewEncoder(w).Encode(result)
	}
}

func unavailable(w http.ResponseWriter) {
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "market data unavailable"})
}

func present(row Symbol, ticker map[string]string) Market {
	m := Market{
		Symbol: row.Symbol, BaseAsset: row.BaseAsset, QuoteAsset: row.QuoteAsset, CoinClass: row.CoinClass,
		ATR: normalisePointer(row.ATR), ATRPercent: normalisePointer(row.ATRPercent),
		Price: normalise(ticker["price"]), QuoteVolume24h: normalise(ticker["quote_volume"]),
	}
	price, priceErr := decimal.NewFromString(ticker["price"])
	open, openErr := decimal.NewFromString(ticker["open"])
	if priceErr == nil && openErr == nil && open.IsPositive() {
		change := price.Sub(open).Mul(decimal.NewFromInt(100)).DivRound(open, 2).StringFixed(2)
		m.Change24hPercent = &change
	}
	if ms, err := strconv.ParseInt(ticker["event_time_ms"], 10, 64); err == nil && ms >= 0 {
		stamp := time.UnixMilli(ms).UTC()
		if stamp.Year() <= 9999 {
			value := stamp.Format(time.RFC3339Nano)
			m.PriceTime = &value
		}
	}
	if m.ATRPercent != nil && row.CoinClass != "stable" {
		atr, _ := decimal.NewFromString(*m.ATRPercent)
		stop := atr.Mul(decimal.RequireFromString(DefaultATRMultiplier)).StringFixed(2)
		m.SuggestedStopPercent = &stop
	}
	return m
}

func normalisePointer(value *string) *string {
	if value == nil {
		return nil
	}
	return normalise(*value)
}

func normalise(value string) *string {
	n, err := decimal.NewFromString(value)
	if err != nil {
		return nil
	}
	s := n.String()
	return &s
}
