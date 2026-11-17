import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';
import '../../data/models.dart';
import '../../data/risk_level.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class HomeStrings {
  static HomeStrings of(BuildContext context) =>
      CatalogScope.maybeOf<HomeStrings>(context) ?? const EnglishHomeStrings();
  String get paper;
  String get riskStatus;
  String get openRisk;
  String get dailyCap;
  String get weeklyCap;
  String get stopsTrigger;
  String get today;
  String get thisWeek;
  String get cooldown;
  String get waitUntil;
  String get minutesFromNow;
  String get positions;
  String get noPositions;
  String get toStop;
  String get atRisk;
  String headline(RiskLevel level);
  String coinClass(CoinClass value);
}

final class EnglishHomeStrings implements HomeStrings {
  const EnglishHomeStrings();

  @override
  String get paper => 'PAPER';
  @override
  String get riskStatus => 'Risk status';
  @override
  String get openRisk => 'Total open risk';
  @override
  String get dailyCap => 'Daily loss cap';
  @override
  String get weeklyCap => 'Weekly loss cap';
  @override
  String get stopsTrigger => 'lost if every stop triggers at once.';
  @override
  String get today => 'left today';
  @override
  String get thisWeek => 'left this week';
  @override
  String get cooldown => 'Cooldown after a loss';
  @override
  String get waitUntil => 'Your rule asks you to wait until ';
  @override
  String get minutesFromNow => ' minutes from now.';
  @override
  String get positions => 'Open positions';
  @override
  String get noPositions => 'No open positions';
  @override
  String get toStop => 'to stop';
  @override
  String get atRisk => 'at risk';
  @override
  String headline(RiskLevel level) => switch (level) {
    RiskLevel.safe => 'Within your limits',
    RiskLevel.caution => 'Approaching your limits',
    RiskLevel.limit => 'Limit exceeded',
  };
  @override
  String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large-cap',
    CoinClass.smallCap => 'Small-cap',
  };
}

final class TurkishHomeStrings implements HomeStrings {
  const TurkishHomeStrings();
  @override
  String get paper => 'SANAL';
  @override
  String get riskStatus => 'Risk durumu';
  @override
  String get openRisk => 'Toplam açık risk';
  @override
  String get dailyCap => 'Günlük zarar limiti';
  @override
  String get weeklyCap => 'Haftalık zarar limiti';
  @override
  String get stopsTrigger => 'tüm stoplar aynı anda tetiklenirse kaybedilir.';
  @override
  String get today => 'bugün kalan';
  @override
  String get thisWeek => 'bu hafta kalan';
  @override
  String get cooldown => 'Zarardan sonra bekleme süresi';
  @override
  String get waitUntil => 'Kuralınıza göre beklemeniz gereken saat: ';
  @override
  String get minutesFromNow => ' dakika sonra.';
  @override
  String get positions => 'Açık pozisyonlar';
  @override
  String get noPositions => 'Açık pozisyon yok';
  @override
  String get toStop => 'stopa mesafe';
  @override
  String get atRisk => 'risk altında';
  @override
  String headline(RiskLevel level) => switch (level) {
    RiskLevel.safe => 'Limitleriniz dahilinde',
    RiskLevel.caution => 'Limitlerinize yaklaşıyorsunuz',
    RiskLevel.limit => 'Limit aşıldı',
  };
  @override
  String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Büyük piyasa değerli',
    CoinClass.smallCap => 'Küçük piyasa değerli',
  };
}
