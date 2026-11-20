package risk

import "time"

// SavedProfile adds persistence metadata without accepting it in request JSON.
type SavedProfile struct {
	Profile
	UpdatedAt *time.Time `json:"updated_at"`
}

func count(value uint32) *uint32 { return &value }

func DefaultProfile() Profile {
	return Profile{
		RiskPerTradePercent: "2", TargetRiskReward: "2", TotalOpenRiskLimitPercent: "6",
		MaxOpenPositions: count(5), MaxLosingTradesToday: count(3), DailyLossLimitPercent: "5",
		MaxStoppedTradesThisWeek: count(5), WeeklyLossLimitPercent: "10",
	}
}

// Valid allows partial preview overrides; a saved profile requires all fields.
func (p Profile) Valid(required bool) bool {
	for _, value := range []string{p.RiskPerTradePercent, p.TargetRiskReward, p.TotalOpenRiskLimitPercent, p.DailyLossLimitPercent, p.WeeklyLossLimitPercent} {
		if value == "" && !required {
			continue
		}
		if !validDecimal(value, 100) {
			return false
		}
	}
	for _, value := range []*uint32{p.MaxOpenPositions, p.MaxLosingTradesToday, p.MaxStoppedTradesThisWeek} {
		if value == nil {
			if required {
				return false
			}
		} else if *value == 0 {
			return false
		}
	}
	return true
}

// ResolveProfile overlays only supplied preview fields on the account limits.
// Nil stored profiles represent accounts that have never saved their defaults.
func ResolveProfile(stored, preview *Profile) Profile {
	p := DefaultProfile()
	for _, override := range []*Profile{stored, preview} {
		if override == nil {
			continue
		}
		for _, pair := range []struct{ dst, src *string }{
			{&p.RiskPerTradePercent, &override.RiskPerTradePercent}, {&p.TargetRiskReward, &override.TargetRiskReward},
			{&p.TotalOpenRiskLimitPercent, &override.TotalOpenRiskLimitPercent}, {&p.DailyLossLimitPercent, &override.DailyLossLimitPercent},
			{&p.WeeklyLossLimitPercent, &override.WeeklyLossLimitPercent},
		} {
			if *pair.src != "" {
				*pair.dst = *pair.src
			}
		}
		for _, pair := range []struct {
			dst **uint32
			src *uint32
		}{
			{&p.MaxOpenPositions, override.MaxOpenPositions}, {&p.MaxLosingTradesToday, override.MaxLosingTradesToday},
			{&p.MaxStoppedTradesThisWeek, override.MaxStoppedTradesThisWeek},
		} {
			if pair.src != nil {
				*pair.dst = count(*pair.src)
			}
		}
	}
	return p
}
