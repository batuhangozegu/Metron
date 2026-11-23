package portfolio

import (
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestDrawdown(t *testing.T) {
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		values                []string
		percent, peak, trough string
	}{
		{nil, "", "", ""}, {[]string{"0", "-1"}, "", "", ""},
		{[]string{"100", "120", "150"}, "", "", ""},
		{[]string{"100", "120", "90", "140", "112"}, "-25", "120", "90"},
		{[]string{"100", "-20"}, "-120", "100", "-20"},
	} {
		var points []Point
		for i, v := range tc.values {
			points = append(points, Point{At: base.Add(time.Duration(i) * time.Hour), Equity: v})
		}
		d := drawdown(points)
		if tc.percent == "" {
			if d != nil {
				t.Fatalf("expected no drawdown: %+v", d)
			}
			continue
		}
		if d == nil || d.Percent != tc.percent || d.Peak != tc.peak || d.Trough != tc.trough || !d.PeakAt.Before(d.TroughAt) {
			t.Fatalf("wrong drawdown: %+v", d)
		}
	}
}

func TestDownsample(t *testing.T) {
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	var points []Point
	for i := range 1001 {
		points = append(points, Point{At: base.Add(time.Duration(i) * time.Hour), Equity: decimal.NewFromInt(int64(i)).String()})
	}
	got := downsample(points, 200)
	if len(got) > 200 || got[len(got)-1] != points[len(points)-1] {
		t.Fatalf("bad downsample size/live: %d", len(got))
	}
	width := points[len(points)-1].At.Sub(base)/200 + 1
	for i, p := range got {
		if i > 0 && !p.At.After(got[i-1].At) {
			t.Fatal("not ascending")
		}
		bucket := p.At.Sub(base) / width
		for _, source := range points {
			if source.At.After(p.At) && source.At.Sub(base)/width == bucket {
				t.Fatal("did not retain last in bucket")
			}
		}
	}
	if len(downsample(points[:2], 200)) != 2 {
		t.Fatal("small series changed")
	}
	if len(downsample([]Point{{At: base}, {At: base}}, 1)) != 1 {
		t.Fatal("equal timestamps")
	}
}

func TestStatisticsAndRMultiple(t *testing.T) {
	var positions []closedPosition
	for i, value := range []string{"100", "300", "-100", "0"} {
		positions = append(positions, closedPosition{CoinClass: []string{"core", "large", "core", "small"}[i], NetPNL: decimal.RequireFromString(value)})
	}
	s := statistics(positions)
	if s.Trades != 4 || s.Wins != 2 || s.Losses != 2 || *s.WinRate != "50.0" || *s.AverageWin != "200.00" || *s.AverageLoss != "-50.00" || *s.WinLossRatio != "4.00" {
		t.Fatalf("bad stats: %+v", s)
	}
	if len(s.ByClass) != 3 || s.ByClass[0].CoinClass != "core" || s.ByClass[0].NetPNL != "0" || s.ByClass[0].WinRate != "50.0" {
		t.Fatalf("bad class stats: %+v", s.ByClass)
	}
	empty := statistics(nil)
	if empty.WinRate != nil || empty.AverageWin != nil || empty.AverageLoss != nil || empty.WinLossRatio != nil || empty.ByClass == nil {
		t.Fatal("empty stats must use null/[]")
	}
	for _, tc := range []struct{ pnl, initial, want string }{{"-10", "10", "-1.00"}, {"10", "3", "3.33"}, {"0", "1", "0.00"}, {"10", "0", ""}, {"10", "-1", ""}} {
		v := rMultiple(decimal.RequireFromString(tc.pnl), decimal.RequireFromString(tc.initial))
		if tc.want == "" {
			if v != nil {
				t.Fatal("invalid risk should be null")
			}
		} else if v == nil || *v != tc.want {
			t.Fatalf("bad r: %v", v)
		}
	}
}

func TestHoldings(t *testing.T) {
	d := decimal.RequireFromString
	h := holdings([]openPosition{{Symbol: "BTCUSDT", Side: "long", Quantity: d("1"), Entry: d("100"), Fee: d("0.1")}, {Symbol: "BTCUSDT", Side: "long", Quantity: d("3"), Entry: d("200"), Fee: d("0.6")}, {Symbol: "BTCUSDT", Side: "short", Quantity: d("2"), Entry: d("300"), Fee: d("0.3")}})
	if len(h) != 2 || h[0].Quantity != "4" || h[0].AverageCost != "175" || h[0].FeesPaid != "0.7" || h[1].Side != "short" {
		t.Fatalf("bad holdings: %+v", h)
	}
}

func TestBenchmark(t *testing.T) {
	empty, ret := benchmark(nil, decimal.NewFromInt(1000))
	if empty == nil || len(empty) != 0 || ret != nil {
		t.Fatal("empty benchmark")
	}
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	points := []Point{{At: at, Equity: "100"}, {At: at.Add(24 * time.Hour), Equity: "110"}}
	got, ret := benchmark(points, decimal.NewFromInt(1000))
	if len(got) != 2 || got[0].Equity != "1000" || got[1].Equity != "1100" || ret == nil || *ret != "10" {
		t.Fatalf("benchmark %+v %v", got, ret)
	}
	for i := 2; i < 1000; i++ {
		points = append(points, Point{At: at.Add(time.Duration(i) * 24 * time.Hour), Equity: decimal.NewFromInt(int64(i + 100)).String()})
	}
	got, _ = benchmark(points, decimal.NewFromInt(1000))
	if len(got) > 200 || got[0].Equity != "1000" {
		t.Fatal("downsample lost benchmark base")
	}
}
