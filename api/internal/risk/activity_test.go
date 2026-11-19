package risk

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestActivityUTCBoundariesAndNetLoss(t *testing.T) {
	// Thursday 01:00 in Istanbul is still Wednesday in the UTC window.
	now := time.Date(2026, 11, 19, 1, 0, 0, 0, time.FixedZone("Istanbul", 3*60*60))
	day, week := UTCWindows(now)
	if day.Format(time.RFC3339) != "2026-11-18T00:00:00Z" || week.Format(time.RFC3339) != "2026-11-16T00:00:00Z" {
		t.Fatalf("wrong windows %s %s", day, week)
	}
	closed := func(at time.Time, pnl, reason string) ActivityPosition {
		return ActivityPosition{Status: "closed", ClosedAt: at, RealizedPNL: decimal.RequireFromString(pnl), CloseReason: reason}
	}
	positions := []ActivityPosition{
		{Status: "open", CoinClass: "core", Quantity: decimal.RequireFromString("1.25"), EntryPrice: decimal.RequireFromString("100.00000001"), InitialRisk: decimal.RequireFromString("5.123456789")},
		{Status: "open", CoinClass: "core", Quantity: decimal.NewFromInt(2), EntryPrice: decimal.NewFromInt(10), InitialRisk: decimal.RequireFromString("4.876543211")},
		{Status: "open", CoinClass: "large", Quantity: decimal.NewFromInt(9), EntryPrice: decimal.NewFromInt(10), InitialRisk: decimal.NewFromInt(5)},
		closed(week.Add(-time.Nanosecond), "-999", "stop"), // previous week
		closed(week, "-20", "stop"),                        // inclusive Monday boundary
		closed(day.Add(-time.Nanosecond), "-5", "stop"),    // previous day, this week
		closed(day, "-10", "stop"),                         // inclusive daily boundary
		closed(day.Add(time.Hour), "3", "manual"),
		closed(now, "-1", "manual"),
		closed(now.Add(time.Nanosecond), "-999", "stop"), // future events excluded
	}
	a, holding := AggregateActivity(positions, decimal.NewFromInt(1000), "core", "spot", now)
	if a.OpenPositions != 3 || a.OpenRiskPercent != "1.5" || a.LosingTradesToday != 2 || a.DailyLossPercent != "0.8" || a.StoppedTradesThisWeek != 3 || a.WeeklyLossPercent != "3.3" || holding != "145.00000002" { // 145.0000000125 rounded up to the engine's 8 decimals
		t.Fatalf("wrong aggregate %+v, holding %s", a, holding)
	}
	_, holding = AggregateActivity(positions, decimal.NewFromInt(1000), "core", "futures", now)
	if holding != "0" {
		t.Fatal("futures must not have spot class holdings")
	}
}

func TestMondayAndSundayWindows(t *testing.T) {
	for _, tc := range []struct{ now, day, week string }{
		{"2026-11-16T00:00:00Z", "2026-11-16", "2026-11-16"},
		{"2026-11-15T23:59:59.999999999Z", "2026-11-15", "2026-11-09"},
		{"2027-01-01T12:00:00Z", "2027-01-01", "2026-12-28"},
	} {
		now, err := time.Parse(time.RFC3339Nano, tc.now)
		if err != nil {
			t.Fatal(err)
		}
		day, week := UTCWindows(now)
		if day.Format(time.DateOnly) != tc.day || week.Format(time.DateOnly) != tc.week {
			t.Fatalf("%s: %s %s", tc.now, day, week)
		}
	}
}

func TestActivityProfitsOffsetLossesAndEmptyHistory(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	positions := []ActivityPosition{
		{Status: "closed", ClosedAt: now, RealizedPNL: decimal.NewFromInt(-10), CloseReason: "stop"},
		{Status: "closed", ClosedAt: now, RealizedPNL: decimal.NewFromInt(11), CloseReason: "manual"},
	}
	a, holding := AggregateActivity(positions, decimal.NewFromInt(100), "core", "spot", now)
	if a.DailyLossPercent != "0" || a.WeeklyLossPercent != "0" || a.LosingTradesToday != 1 || a.StoppedTradesThisWeek != 1 || holding != "0" {
		t.Fatalf("wrong net loss %+v", a)
	}
	a, holding = AggregateActivity(nil, decimal.NewFromInt(100), "core", "spot", now)
	if a.OpenRiskPercent != "0" || a.DailyLossPercent != "0" || a.WeeklyLossPercent != "0" || a.OpenPositions != 0 || holding != "0" {
		t.Fatalf("wrong empty history %+v", a)
	}
}
