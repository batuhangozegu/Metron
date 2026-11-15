import '../theme/risk_tokens.dart';

/// Independent of profit/loss. A limit is exceeded only above 100%.
enum RiskLevel {
  safe,
  caution,
  limit;

  static RiskLevel forUsage(double value, double limit) {
    if (!value.isFinite || !limit.isFinite || value < 0 || limit <= 0) {
      throw ArgumentError(
        'Usage must be finite and non-negative; limit must be positive.',
      );
    }
    final fraction = value / limit;
    if (fraction > 1) return RiskLevel.limit;
    if (fraction >= MetronRisk.cautionAtFractionOfLimit) {
      return RiskLevel.caution;
    }
    return RiskLevel.safe;
  }
}
