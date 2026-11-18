import 'package:flutter/widgets.dart';

import 'catalog_scope.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class AppStrings {
  static AppStrings of(BuildContext context) =>
      CatalogScope.maybeOf<AppStrings>(context) ?? const EnglishAppStrings();
  String get spot;
  String get futures;
  String get accountMarket;
  String get dailyLossBlockedTitle;
  String get dailyLossBlockedBody;
  String get waitUntilReset;
  String get losingTradesToday;
  String get riskPerTradeLimit;
  String get dailyLosingTradesLimit;
  String get weeklyStoppedTradesLimit;
  String get spotAllocationCap;
  String get ticker;
  String get sampleFeed;
  String get pauseTicker;
  String get resumeTicker;
  String get dailyChange;
  String get blockedStamp;
  String get warRoom;
  String get coreClass;
  String get largeCapClass;
  String get smallCapClass;
  String get dailyAtr;
  String get tradeRisk;
  String get stopDistance;
  String get appName;
  String get risk;
  String get markets;
  String get alerts;
  String get rules;
  String get check;
  String get checkBlocked;
  String get position;
  String get signIn;
  String get todo;
  String get safe;
  String get caution;
  String get limit;
  String get profit;
  String get loss;
  String get unchanged;
  String get outOf;
  String get checkNewOrder;
  String get back;
  String get notFound;
  String get language;
}

final class EnglishAppStrings implements AppStrings {
  const EnglishAppStrings();

  @override
  String get spot => 'Spot';
  @override
  String get futures => 'Futures';
  @override
  String get accountMarket => 'Account market';
  @override
  String get dailyLossBlockedTitle => 'Blocked: daily loss limit reached';
  @override
  String get dailyLossBlockedBody =>
      '3 of today’s trades closed at a loss; your rule allows 3 per day. No new trades until 00:00 UTC.';
  @override
  String get waitUntilReset => 'Wait until 00:00 UTC';
  @override
  String get losingTradesToday => 'Losing trades today';
  @override
  String get riskPerTradeLimit => 'Risk per trade';
  @override
  String get dailyLosingTradesLimit => 'Losing trades per day';
  @override
  String get weeklyStoppedTradesLimit => 'Stopped trades per week';
  @override
  String get spotAllocationCap => 'Spot portfolio allocation cap';
  @override
  String get ticker => 'Market ticker';
  @override
  String get sampleFeed => 'SAMPLE';
  @override
  String get pauseTicker => 'Pause ticker';
  @override
  String get resumeTicker => 'Resume ticker';
  @override
  String get dailyChange => '24-hour change';
  @override
  String get blockedStamp => 'BLOCKED';
  @override
  String get warRoom => 'RISK CONTROL / WAR ROOM';
  @override
  String get coreClass => 'BTC/ETH';
  @override
  String get largeCapClass => 'Large-cap';
  @override
  String get smallCapClass => 'Small-cap';
  @override
  String get dailyAtr => 'ATR / day';
  @override
  String get tradeRisk => 'risk';
  @override
  String get stopDistance => 'stop';

  @override
  String get appName => 'Metron';
  @override
  String get risk => 'Risk';
  @override
  String get markets => 'Markets';
  @override
  String get alerts => 'Alerts';
  @override
  String get rules => 'Rules';
  @override
  String get check => 'Pre-trade check';
  @override
  String get checkBlocked => 'Pre-trade check — blocked';
  @override
  String get position => 'Position';
  @override
  String get signIn => 'Sign in';
  @override
  String get todo => 'TODO: Replace this placeholder with the mobile mockup.';
  @override
  String get safe => 'Safe';
  @override
  String get caution => 'Caution';
  @override
  String get limit => 'Limit exceeded';
  @override
  String get profit => 'Profit';
  @override
  String get loss => 'Loss';
  @override
  String get unchanged => 'Unchanged';
  @override
  String get outOf => 'of';
  @override
  String get checkNewOrder => 'Check a new order';
  @override
  String get back => 'Back';
  @override
  String get notFound => 'Page not found';

  @override
  String get language => 'Language';
}

final class TurkishAppStrings implements AppStrings {
  const TurkishAppStrings();

  @override
  String get spot => 'Spot';
  @override
  String get futures => 'Vadeli';
  @override
  String get accountMarket => 'Hesap piyasası';
  @override
  String get dailyLossBlockedTitle => 'Engellendi: günlük kayıp sınırı doldu';
  @override
  String get dailyLossBlockedBody =>
      'Bugün 3 işlem zararla kapandı; kuralınız günde en fazla 3 kayıplı işleme izin veriyor. 00:00 UTC’ye kadar yeni işlem açılamaz.';
  @override
  String get waitUntilReset => '00:00 UTC’ye kadar bekle';
  @override
  String get losingTradesToday => 'Bugünkü kayıplı işlemler';
  @override
  String get riskPerTradeLimit => 'İşlem başına risk';
  @override
  String get dailyLosingTradesLimit => 'Günlük kayıplı işlem sınırı';
  @override
  String get weeklyStoppedTradesLimit => 'Haftalık stop olan işlem sınırı';
  @override
  String get spotAllocationCap => 'Spot portföy payı üst sınırı';
  @override
  String get ticker => 'Piyasa bandı';
  @override
  String get sampleFeed => 'ÖRNEK';
  @override
  String get pauseTicker => 'Piyasa bandını duraklat';
  @override
  String get resumeTicker => 'Piyasa bandını sürdür';
  @override
  String get dailyChange => '24 saatlik değişim';
  @override
  String get blockedStamp => 'ENGELLENDİ';
  @override
  String get warRoom => 'RİSK KONTROL / HAREKÂT ODASI';
  @override
  String get coreClass => 'BTC/ETH';
  @override
  String get largeCapClass => 'Büyük hacimli';
  @override
  String get smallCapClass => 'Küçük hacimli';
  @override
  String get dailyAtr => 'ATR / gün';
  @override
  String get tradeRisk => 'risk';
  @override
  String get stopDistance => 'stop';
  @override
  String get appName => 'Metron';
  @override
  String get risk => 'Risk';
  @override
  String get markets => 'Piyasalar';
  @override
  String get alerts => 'Bildirimler';
  @override
  String get rules => 'Kurallar';
  @override
  String get check => 'İşlem öncesi kontrol';
  @override
  String get checkBlocked => 'İşlem öncesi kontrol — engellendi';
  @override
  String get position => 'Pozisyon';
  @override
  String get signIn => 'Giriş yap';
  @override
  String get todo => 'YAPILACAK: Bu alanı mobil tasarımla değiştirin.';
  @override
  String get safe => 'Güvenli';
  @override
  String get caution => 'Dikkat';
  @override
  String get limit => 'Limit aşıldı';
  @override
  String get profit => 'Kâr';
  @override
  String get loss => 'Zarar';
  @override
  String get unchanged => 'Değişim yok';
  @override
  String get outOf => '/';
  @override
  String get checkNewOrder => 'Yeni emri kontrol et';
  @override
  String get back => 'Geri';
  @override
  String get notFound => 'Sayfa bulunamadı';
  @override
  String get language => 'Dil';
}
