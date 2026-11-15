import '../../data/models.dart';

/// Feature-local English copy, ready for locale wiring.
abstract final class RulesStrings {
  static const introduction =
      'Shown here, changed on the web. Rules are set at a desk, not in the moment.';
  static const accountLimits = 'Account limits';
  static const openRisk = 'Total open risk';
  static const dailyCap = 'Daily loss cap';
  static const weeklyCap = 'Weekly loss cap';
  static const positions = 'Open positions at once';
  static const riskReward = 'Target risk/reward';
  static const coinClasses = 'Coin classes';
  static const stop = 'Stop';
  static const riskPerTrade = 'Risk per trade';
  static const largestPosition = 'Largest position';
  static const discipline = 'Discipline';
  static const account = 'Account';
  static const tradingAccount = 'Trading account';
  static const paper = 'Paper ·';
  static const twoFactor = 'Two-factor sign-in';
  static const on = 'On';
  static const off = 'Off';
  static const signOut = 'Sign out';
  static const noClasses = 'No coin class rules set.';
  static const noDiscipline = 'No discipline rules set.';
  static String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large-cap',
    CoinClass.smallCap => 'Small-cap',
  };
  static String mode(DisciplineMode value) => switch (value) {
    DisciplineMode.warnAndRecord => 'Warn and record',
    DisciplineMode.block => 'Block',
  };
}
