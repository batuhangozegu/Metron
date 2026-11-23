// Package portfolio exposes paper-account performance and the settled trade journal.
package portfolio

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type Point struct {
	At     time.Time `json:"at"`
	Equity string    `json:"equity"`
}

type Summary struct {
	StartingBalance string `json:"starting_balance"`
	Balance         string `json:"balance"`
	Cash            string `json:"cash"`
	UnrealizedPNL   string `json:"unrealized_pnl"`
	Equity          string `json:"equity"`
	NetResult       string `json:"net_result"`
	ReturnPercent   string `json:"return_percent"`
	FeesPaid        string `json:"fees_paid"`
}

type Drawdown struct {
	Percent  string    `json:"percent"`
	Peak     string    `json:"peak"`
	Trough   string    `json:"trough"`
	PeakAt   time.Time `json:"peak_at"`
	TroughAt time.Time `json:"trough_at"`
}

type ClassStats struct {
	CoinClass string `json:"coin_class"`
	Trades    int    `json:"trades"`
	Wins      int    `json:"wins"`
	WinRate   string `json:"win_rate"`
	NetPNL    string `json:"net_pnl"`
}

type Stats struct {
	Trades       int          `json:"trades"`
	Wins         int          `json:"wins"`
	Losses       int          `json:"losses"`
	WinRate      *string      `json:"win_rate"`
	AverageWin   *string      `json:"average_win"`
	AverageLoss  *string      `json:"average_loss"`
	WinLossRatio *string      `json:"win_loss_ratio"`
	ByClass      []ClassStats `json:"by_class"`
}

type Holding struct {
	Symbol      string `json:"symbol"`
	Side        string `json:"side"`
	Quantity    string `json:"quantity"`
	AverageCost string `json:"average_cost"`
	FeesPaid    string `json:"fees_paid"`
}

type Portfolio struct {
	AccountID              string    `json:"account_id"`
	Market                 string    `json:"market"`
	Currency               string    `json:"currency"`
	Period                 string    `json:"period"`
	Summary                Summary   `json:"summary"`
	EquitySeries           []Point   `json:"equity_series"`
	BenchmarkSeries        []Point   `json:"benchmark_series"`
	BenchmarkReturnPercent *string   `json:"benchmark_return_percent"`
	MaxDrawdown            *Drawdown `json:"max_drawdown"`
	Stats                  Stats     `json:"stats"`
	Holdings               []Holding `json:"holdings"`
}

type Rule struct {
	Rule   string `json:"rule"`
	Status string `json:"status"`
}

type Trade struct {
	PositionID  string    `json:"position_id"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"`
	CoinClass   string    `json:"coin_class"`
	OpenedAt    time.Time `json:"opened_at"`
	ClosedAt    time.Time `json:"closed_at"`
	CloseReason string    `json:"close_reason"`
	Quantity    string    `json:"quantity"`
	EntryPrice  string    `json:"entry_price"`
	ExitPrice   string    `json:"exit_price"`
	Fees        string    `json:"fees"`
	NetPNL      string    `json:"net_pnl"`
	RMultiple   *string   `json:"r_multiple"`
	Rules       []Rule    `json:"rules"`
}

type Journal struct {
	Trades     []Trade    `json:"trades"`
	NextBefore *time.Time `json:"next_before"`
}

func ratio(n, d decimal.Decimal, places int32) string {
	if d.IsZero() {
		return "0"
	}
	return n.DivRound(d, places).String()
}

func rMultiple(pnl, initialRisk decimal.Decimal) *string {
	if !initialRisk.IsPositive() {
		return nil
	}
	v := pnl.DivRound(initialRisk, 2).StringFixed(2)
	return &v
}

func benchmark(candles []Point, initial decimal.Decimal) ([]Point, *string) {
	points := downsample(candles, 200)
	result := make([]Point, 0, len(points))
	if len(points) == 0 {
		return result, nil
	}
	base := decimal.RequireFromString(points[0].Equity)
	for _, p := range points {
		price := decimal.RequireFromString(p.Equity)
		result = append(result, Point{At: p.At, Equity: ratio(price.Mul(initial), base, 28)})
	}
	last := decimal.RequireFromString(points[len(points)-1].Equity)
	v := ratio(last.Sub(base).Mul(decimal.NewFromInt(100)), base, 2)
	return result, &v
}

// downsample keeps the last observation in each equal-duration time bucket.
// The final bucket always retains the live point.
func downsample(points []Point, limit int) []Point {
	if len(points) <= limit {
		return points
	}
	if limit < 1 {
		return []Point{}
	}
	span := points[len(points)-1].At.Sub(points[0].At)
	width := span/time.Duration(limit) + 1
	result := make([]Point, 0, limit)
	var previous time.Duration = -1
	for _, p := range points {
		bucket := p.At.Sub(points[0].At) / width
		if bucket == previous {
			result[len(result)-1] = p
		} else {
			result = append(result, p)
		}
		previous = bucket
	}
	return result
}

func drawdown(points []Point) *Drawdown {
	var result *Drawdown
	peak, worst := decimal.Zero, decimal.Zero
	var peakAt time.Time
	for _, p := range points {
		value := decimal.RequireFromString(p.Equity)
		if value.GreaterThan(peak) {
			peak, peakAt = value, p.At
		}
		if !peak.IsPositive() {
			continue
		}
		drop := value.Sub(peak).Mul(decimal.NewFromInt(100)).DivRound(peak, 28)
		if drop.LessThan(worst) {
			worst = drop
			result = &Drawdown{Percent: drop.Round(2).String(), Peak: peak.String(), Trough: value.String(), PeakAt: peakAt, TroughAt: p.At}
		}
	}
	return result
}

type closedPosition struct {
	CoinClass string
	NetPNL    decimal.Decimal
}

func statistics(positions []closedPosition) Stats {
	s := Stats{ByClass: []ClassStats{}}
	classes := map[string]*ClassStats{}
	wins, losses := decimal.Zero, decimal.Zero
	for _, p := range positions {
		pnl := p.NetPNL
		s.Trades++
		c := classes[p.CoinClass]
		if c == nil {
			c = &ClassStats{CoinClass: p.CoinClass, NetPNL: "0"}
			classes[p.CoinClass] = c
		}
		c.Trades++
		c.NetPNL = decimal.RequireFromString(c.NetPNL).Add(pnl).String()
		if pnl.IsPositive() {
			s.Wins++
			c.Wins++
			wins = wins.Add(pnl)
		} else {
			s.Losses++
			losses = losses.Add(pnl)
		}
	}
	if s.Trades > 0 {
		v := decimal.NewFromInt(int64(s.Wins*100)).DivRound(decimal.NewFromInt(int64(s.Trades)), 1).StringFixed(1)
		s.WinRate = &v
	}
	if s.Wins > 0 {
		v := wins.DivRound(decimal.NewFromInt(int64(s.Wins)), 2).StringFixed(2)
		s.AverageWin = &v
	}
	if s.Losses > 0 {
		v := losses.DivRound(decimal.NewFromInt(int64(s.Losses)), 2).StringFixed(2)
		s.AverageLoss = &v
	}
	if s.Wins > 0 && losses.IsNegative() {
		v := wins.Mul(decimal.NewFromInt(int64(s.Losses))).DivRound(losses.Abs().Mul(decimal.NewFromInt(int64(s.Wins))), 2).StringFixed(2)
		s.WinLossRatio = &v
	}
	for _, name := range []string{"core", "large", "small", "stable"} {
		if c := classes[name]; c != nil {
			c.WinRate = decimal.NewFromInt(int64(c.Wins*100)).DivRound(decimal.NewFromInt(int64(c.Trades)), 1).StringFixed(1)
			s.ByClass = append(s.ByClass, *c)
		}
	}
	return s
}

type openPosition struct {
	Symbol, Side         string
	Quantity, Entry, Fee decimal.Decimal
	Leverage             uint32
}

func holdings(positions []openPosition) []Holding {
	groups := map[string]Holding{}
	for _, p := range positions {
		key := p.Symbol + ":" + p.Side
		h, ok := groups[key]
		if !ok {
			h = Holding{Symbol: p.Symbol, Side: p.Side, Quantity: "0", AverageCost: "0", FeesPaid: "0"}
		}
		// AverageCost holds total cost until all quantities have been summed.
		h.Quantity = decimal.RequireFromString(h.Quantity).Add(p.Quantity).String()
		h.AverageCost = decimal.RequireFromString(h.AverageCost).Add(p.Quantity.Mul(p.Entry)).String()
		h.FeesPaid = decimal.RequireFromString(h.FeesPaid).Add(p.Fee).String()
		groups[key] = h
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := []Holding{}
	for _, key := range keys {
		h := groups[key]
		h.AverageCost = ratio(decimal.RequireFromString(h.AverageCost), decimal.RequireFromString(h.Quantity), 28)
		result = append(result, h)
	}
	return result
}
