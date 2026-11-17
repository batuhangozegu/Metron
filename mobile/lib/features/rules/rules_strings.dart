import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';
import '../../data/models.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class RulesStrings {
  static RulesStrings of(BuildContext context) =>
      CatalogScope.maybeOf<RulesStrings>(context) ??
      const EnglishRulesStrings();
  String get introduction;
  String get accountLimits;
  String get openRisk;
  String get dailyCap;
  String get weeklyCap;
  String get positions;
  String get riskReward;
  String get coinClasses;
  String get stop;
  String get riskPerTrade;
  String get largestPosition;
  String get discipline;
  String get account;
  String get tradingAccount;
  String get paper;
  String get twoFactor;
  String get on;
  String get off;
  String get signOut;
  String get noClasses;
  String get noDiscipline;
  String coinClass(CoinClass value);
  String mode(DisciplineMode value);
  String ruleName(DisciplineRule rule);
  String ruleDescription(DisciplineRule rule);
  String ruleHistory(DisciplineRule rule);
}

final class EnglishRulesStrings implements RulesStrings {
  const EnglishRulesStrings();

  @override
  String get introduction =>
      'Shown here, changed on the web. Rules are set at a desk, not in the moment.';
  @override
  String get accountLimits => 'Account limits';
  @override
  String get openRisk => 'Total open risk';
  @override
  String get dailyCap => 'Daily loss cap';
  @override
  String get weeklyCap => 'Weekly loss cap';
  @override
  String get positions => 'Open positions at once';
  @override
  String get riskReward => 'Target risk/reward';
  @override
  String get coinClasses => 'Coin classes';
  @override
  String get stop => 'Stop';
  @override
  String get riskPerTrade => 'Risk per trade';
  @override
  String get largestPosition => 'Largest position';
  @override
  String get discipline => 'Discipline';
  @override
  String get account => 'Account';
  @override
  String get tradingAccount => 'Trading account';
  @override
  String get paper => 'Paper ·';
  @override
  String get twoFactor => 'Two-factor sign-in';
  @override
  String get on => 'On';
  @override
  String get off => 'Off';
  @override
  String get signOut => 'Sign out';
  @override
  String get noClasses => 'No coin class rules set.';
  @override
  String get noDiscipline => 'No discipline rules set.';
  @override
  String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large-cap',
    CoinClass.smallCap => 'Small-cap',
  };
  @override
  String mode(DisciplineMode value) => switch (value) {
    DisciplineMode.warnAndRecord => 'Warn and record',
    DisciplineMode.block => 'Block',
  };
  @override
  String ruleName(DisciplineRule rule) => rule.name;
  @override
  String ruleDescription(DisciplineRule rule) => rule.description;
  @override
  String ruleHistory(DisciplineRule rule) => rule.history;
}

final class TurkishRulesStrings implements RulesStrings {
  const TurkishRulesStrings();
  @override
  String get introduction =>
      'Kurallar burada görüntülenir, web üzerinden değiştirilir. Kurallar işlem anında değil, önceden belirlenir.';
  @override
  String get accountLimits => 'Hesap limitleri';
  @override
  String get openRisk => 'Toplam açık risk';
  @override
  String get dailyCap => 'Günlük zarar limiti';
  @override
  String get weeklyCap => 'Haftalık zarar limiti';
  @override
  String get positions => 'Aynı anda açık pozisyon';
  @override
  String get riskReward => 'Hedef risk/getiri';
  @override
  String get coinClasses => 'Coin grupları';
  @override
  String get stop => 'Stop';
  @override
  String get riskPerTrade => 'İşlem başına risk';
  @override
  String get largestPosition => 'Azami pozisyon';
  @override
  String get discipline => 'Disiplin';
  @override
  String get account => 'Hesap';
  @override
  String get tradingAccount => 'İşlem hesabı';
  @override
  String get paper => 'Sanal ·';
  @override
  String get twoFactor => 'İki aşamalı doğrulama';
  @override
  String get on => 'Açık';
  @override
  String get off => 'Kapalı';
  @override
  String get signOut => 'Çıkış yap';
  @override
  String get noClasses => 'Coin grubu kuralı belirlenmemiş.';
  @override
  String get noDiscipline => 'Disiplin kuralı belirlenmemiş.';
  @override
  String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Büyük piyasa değerli',
    CoinClass.smallCap => 'Küçük piyasa değerli',
  };
  @override
  String mode(DisciplineMode value) => switch (value) {
    DisciplineMode.warnAndRecord => 'Uyar ve kaydet',
    DisciplineMode.block => 'Engelle',
  };
  @override
  String ruleName(DisciplineRule rule) => switch (rule.id) {
    'cooldown' => 'Zarardan sonra bekleme süresi',
    'overnight' => '00:00 sonrası emir yasağı',
    'stop' => 'Stop asla uzaklaştırılmaz',
    _ => rule.name,
  };
  @override
  String ruleDescription(DisciplineRule rule) => switch (rule.id) {
    'cooldown' =>
      'Bir pozisyon zararla kapandıktan sonra 60 dakika yeni emir verilmez.',
    'overnight' => '00:00 ile 06:00 UTC arasında yeni emir verilmez.',
    'stop' => 'Stop fiyatını yalnızca piyasa fiyatına yaklaştırabilirsiniz, uzaklaştıramazsınız.',
    _ => rule.description,
  };
  @override
  String ruleHistory(DisciplineRule rule) => switch (rule.id) {
    'cooldown' || 'overnight' => '30 günde ${rule.violationCount} ihlal',
    'stop' => '${rule.violationCount} kez engellendi',
    _ => rule.history,
  };
}
