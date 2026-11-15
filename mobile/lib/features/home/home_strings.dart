import '../../data/models.dart';
import '../../data/risk_level.dart';

abstract final class HomeStrings {
  static const paper = 'PAPER';
  static const riskStatus = 'Risk status';
  static const openRisk = 'Total open risk';
  static const dailyCap = 'Daily loss cap';
  static const weeklyCap = 'Weekly loss cap';
  static const stopsTrigger = 'lost if every stop triggers at once.';
  static const today = 'left today';
  static const thisWeek = 'left this week';
  static const cooldown = 'Cooldown after a loss';
  static const waitUntil = 'Your rule asks you to wait until ';
  static const minutesFromNow = ' minutes from now.';
  static const positions = 'Open positions';
  static const noPositions = 'No open positions';
  static const toStop = 'to stop';
  static const atRisk = 'at risk';
  static String headline(RiskLevel level) => switch (level) {
    RiskLevel.safe => 'Within your limits',
    RiskLevel.caution => 'Approaching your limits',
    RiskLevel.limit => 'Limit exceeded',
  };
  static String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large-cap',
    CoinClass.smallCap => 'Small-cap',
  };
}
