import 'package:flutter/widgets.dart';

import 'catalog_scope.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class AppStrings {
  static AppStrings of(BuildContext context) =>
      CatalogScope.maybeOf<AppStrings>(context) ?? const EnglishAppStrings();
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
