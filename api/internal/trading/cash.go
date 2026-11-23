package trading

import "github.com/shopspring/decimal"

// Capital describes capital reserved by an open position, including its fees.
type Capital struct {
	Notional, Fee decimal.Decimal
	Leverage      uint32
}

func Cash(balance decimal.Decimal, market string, positions []Capital) decimal.Decimal {
	numerator, denominator := cashFraction(balance, market, positions)
	value := numerator.DivRound(denominator, 28)
	if value.Mul(denominator).GreaterThan(numerator) {
		value = value.Sub(decimal.New(1, -28))
	}
	return value
}

// cashFraction retains exact repeating margins for authorization. Grouping by
// leverage bounds the denominator's size regardless of the position count.
func cashFraction(balance decimal.Decimal, market string, positions []Capital) (decimal.Decimal, decimal.Decimal) {
	margin := map[uint32]decimal.Decimal{}
	for _, p := range positions {
		balance = balance.Sub(p.Fee)
		leverage := uint32(1)
		if market == "futures" {
			leverage = max(1, p.Leverage)
		}
		margin[leverage] = margin[leverage].Add(p.Notional)
	}
	numerator, denominator := balance, decimal.NewFromInt(1)
	for leverage, notional := range margin {
		l := decimal.NewFromInt(int64(leverage))
		numerator = numerator.Mul(l).Sub(notional.Mul(denominator))
		denominator = denominator.Mul(l)
	}
	return numerator, denominator
}

func affordable(balance decimal.Decimal, market string, positions []Capital, notional, fee decimal.Decimal, leverage uint32) bool {
	numerator, denominator := cashFraction(balance, market, positions)
	l := decimal.NewFromInt(1)
	if market == "futures" {
		l = decimal.NewFromInt(int64(leverage))
	}
	return !notional.Add(fee.Mul(l)).Mul(denominator).GreaterThan(numerator.Mul(l))
}
