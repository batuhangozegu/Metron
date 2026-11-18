import 'risk_level.dart';

/// Fixtures preserve money and quantities as decimal strings, not binary money
/// arithmetic. Percentages are percentage points (2.75 means 2.75%, not .0275).
enum CoinClass { btcEth, largeCap, smallCap }

enum AccountMarket { spot, futures }

class CoinPrice {
  const CoinPrice({
    required this.pair,
    required this.coinClass,
    required this.price,
    required this.changePercent,
    required this.atrPercent,
    required this.suggestedStopPercent,
  });
  final String pair;
  final CoinClass coinClass;
  final String price;
  final double changePercent, atrPercent, suggestedStopPercent;
}

class Position {
  const Position({
    required this.id,
    required this.pair,
    required this.coinClass,
    required this.distanceToStopPercent,
    required this.moneyAtRisk,
    required this.unrealizedPnl,
    this.detail,
  });
  final String id, pair;
  final CoinClass coinClass;
  final double distanceToStopPercent;
  final String moneyAtRisk, unrealizedPnl;
  // Only BTC has a detail mockup. Missing details are deliberately not invented.
  final PositionDetail? detail;
}

class PositionDetail {
  const PositionDetail({
    required this.quantity,
    required this.entry,
    required this.stop,
    required this.target,
    required this.valueNow,
    required this.shareOfAccountPercent,
    required this.openedAt,
    required this.stopOnExchange,
    required this.returnPercent,
    required this.returnR,
  });
  final String quantity, entry, stop, target, valueNow;
  final double shareOfAccountPercent, returnPercent, returnR;
  final DateTime openedAt;
  final bool stopOnExchange;
}

class PaperAccount {
  const PaperAccount({
    required this.equity,
    required this.openRiskPercent,
    required this.dailyLossPercent,
    required this.weeklyLossPercent,
    required this.moneyAtRisk,
    required this.dailyRemaining,
    required this.weeklyRemaining,
    required this.unrealizedPnl,
    required this.asOf,
    required this.cooldownEndsAt,
    required this.cooldownWaitedMinutes,
    required this.twoFactorEnabled,
  });
  final String equity,
      moneyAtRisk,
      dailyRemaining,
      weeklyRemaining,
      unrealizedPnl;
  final double openRiskPercent, dailyLossPercent, weeklyLossPercent;
  final DateTime asOf, cooldownEndsAt;
  final int cooldownWaitedMinutes;
  final bool twoFactorEnabled;
}

class AccountLimits {
  const AccountLimits({
    required this.openRiskPercent,
    required this.dailyLossPercent,
    required this.weeklyLossPercent,
    required this.maxOpenPositions,
    required this.targetRiskReward,
    required this.smallCapTotalPercent,
    this.riskPerTradePercent = 2,
    this.maxLosingTradesPerDay = 3,
    this.maxStoppedTradesPerWeek = 5,
  });
  final double openRiskPercent,
      dailyLossPercent,
      weeklyLossPercent,
      targetRiskReward,
      smallCapTotalPercent;
  final int maxOpenPositions, maxLosingTradesPerDay, maxStoppedTradesPerWeek;
  final double riskPerTradePercent;
}

class CoinClassRule {
  const CoinClassRule({
    required this.coinClass,
    required this.atrMultiplier,
    required this.riskPerTradePercent,
    required this.maxPositionPercent,
  });
  final CoinClass coinClass;
  final double atrMultiplier, riskPerTradePercent, maxPositionPercent;
}

enum DisciplineMode { warnAndRecord, block }

class DisciplineRule {
  const DisciplineRule({
    required this.id,
    required this.name,
    required this.description,
    required this.mode,
    required this.history,
    required this.violationCount,
  });
  final String id, name, description, history;
  final DisciplineMode mode;
  final int violationCount;
}

enum AlertKind { cooldown, stopTriggered, priceAlarm, orderBlocked }

class AlertEvent {
  const AlertEvent({
    required this.id,
    required this.kind,
    required this.title,
    required this.body,
    required this.at,
    this.pnl,
  });
  final String id, title, body;
  final AlertKind kind;
  final DateTime at;
  final String? pnl;
}

enum AlarmDirection { above, below }

class PriceAlarm {
  const PriceAlarm({
    required this.id,
    required this.pair,
    required this.direction,
    required this.targetPrice,
  });
  final String id, pair, targetPrice;
  final AlarmDirection direction;
}

/// A mock verdict, not a calculation performed by the mobile client.
class LimitCheck {
  const LimitCheck({
    required this.id,
    required this.name,
    required this.before,
    required this.after,
    required this.limit,
    required this.unit,
    required this.level,
    required this.valueLabel,
    this.brokenRule = false,
  });
  final String id, name, unit, valueLabel;
  final double before, after, limit;
  final RiskLevel level;
  final bool brokenRule;
}

class OrderPreview {
  const OrderPreview({
    required this.pair,
    required this.verdict,
    required this.entry,
    required this.stop,
    required this.target,
    required this.quantity,
    required this.orderValue,
    required this.lossAtStop,
    required this.profitAtTarget,
    required this.checks,
    this.largestFittingOrderValue,
    this.market = AccountMarket.spot,
    this.largestFittingRisk,
  });
  final String pair,
      entry,
      stop,
      target,
      quantity,
      orderValue,
      lossAtStop,
      profitAtTarget;
  final RiskLevel verdict;
  final AccountMarket market;
  final List<LimitCheck> checks;
  final String? largestFittingOrderValue, largestFittingRisk;
}
