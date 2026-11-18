import '../../data/models.dart';

import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';
import '../../l10n/app_strings.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class CheckStrings {
  static CheckStrings of(BuildContext context) =>
      CatalogScope.maybeOf<CheckStrings>(context) ??
      const EnglishCheckStrings();
  String get paper;
  String get limit;
  String get wait;
  String get safeTitle;
  String get cautionTitle;
  String get safeBody;
  String get broken;
  String get exceeded;
  String get pair;
  String get price;
  String get limits;
  String get stop;
  String get entry;
  String get target;
  String get stopResult;
  String get targetResult;
  String get size;
  String get value;
  String get discard;
  String get reduce;
  String get place;
  String get anyway;
  String get violation;
  String get unavailable;
  String get sampleOrder;
  String get sampleReminder;
  String get reductionTitle;
  String get reductionBody;
  String get close;
  String cooldown(String minutes);
  String reduction(String value, String risk);
  String limitName(LimitCheck check);
  String limitValue(String value);
}

final class EnglishCheckStrings implements CheckStrings {
  const EnglishCheckStrings();

  @override
  String get paper => 'PAPER';
  @override
  String get limit => 'Limit';
  @override
  String get wait => 'Wait, remind me at ';
  @override
  String get safeTitle => 'Within your limits';
  @override
  String get cautionTitle => 'One rule would be broken';
  @override
  String get safeBody => 'Every check is within your limits.';
  @override
  String get broken => 'Broken';
  @override
  String get exceeded => 'Exceeded';
  @override
  String cooldown(String minutes) =>
      'No limit is exceeded. Your cooldown rule asks you '
      'to wait $minutes more minutes.';
  @override
  String reduction(String value, String risk) =>
      'The largest order that fits is $value USDT, risking '
      '$risk USDT. A blocked order cannot be placed anyway.';
  @override
  String get pair => 'Pair';
  @override
  String get price => 'Limit price, USDT';
  @override
  String get limits => 'Your limits after this order';
  @override
  String get stop => 'Stop';
  @override
  String get entry => 'Entry';
  @override
  String get target => 'Target';
  @override
  String get stopResult => 'If the stop triggers';
  @override
  String get targetResult => 'If the target fills';
  @override
  String get size => 'Position size';
  @override
  String get value => 'Order value';
  @override
  String get discard => 'Discard';
  @override
  String get reduce => 'Reduce to fit the limit';
  @override
  String get place => 'Place paper order';
  @override
  String get anyway => 'Place paper order anyway';
  @override
  String get violation => 'Placing it now is recorded as a broken rule.';
  @override
  String get unavailable => 'No order preview is available for this pair.';
  @override
  String get sampleOrder => 'Sample preview only. No order was placed.';
  @override
  String get sampleReminder =>
      'Sample preview only. No reminder was scheduled.';
  @override
  String get reductionTitle => 'Suggested reduction';
  @override
  String get reductionBody =>
      'Recheck the reduced order before placing it. '
      'This sample cannot recalculate limits or clear the cooldown.';
  @override
  String get close => 'Close';
  @override
  String limitName(LimitCheck check) => check.id == 'daily-losing-trades'
      ? const EnglishAppStrings().losingTradesToday
      : check.name;
  @override
  String limitValue(String value) => value;
}

final class TurkishCheckStrings implements CheckStrings {
  const TurkishCheckStrings();
  @override
  String get paper => 'SANAL';
  @override
  String get limit => 'Limit';
  @override
  String get wait => 'Bekle, şu saatte hatırlat: ';
  @override
  String get safeTitle => 'Limitleriniz dahilinde';
  @override
  String get cautionTitle => 'Bir kural ihlal edilecek';
  @override
  String get safeBody => 'Tüm kontroller limitleriniz dahilinde.';
  @override
  String get broken => 'İhlal';
  @override
  String get exceeded => 'Aşıldı';
  @override
  String get pair => 'Parite';
  @override
  String get price => 'Limit fiyatı, USDT';
  @override
  String get limits => 'Bu emirden sonra limitleriniz';
  @override
  String get stop => 'Stop';
  @override
  String get entry => 'Giriş';
  @override
  String get target => 'Hedef';
  @override
  String get stopResult => 'Stop tetiklenirse';
  @override
  String get targetResult => 'Hedef gerçekleşirse';
  @override
  String get size => 'Pozisyon büyüklüğü';
  @override
  String get value => 'Emir tutarı';
  @override
  String get discard => 'Vazgeç';
  @override
  String get reduce => 'Limite uyacak şekilde küçült';
  @override
  String get place => 'Sanal emir ver';
  @override
  String get anyway => 'Yine de sanal emir ver';
  @override
  String get violation => 'Emri şimdi vermek kural ihlali olarak kaydedilir.';
  @override
  String get unavailable => 'Bu parite için emir önizlemesi yok.';
  @override
  String get sampleOrder => 'Yalnızca örnek önizleme. Emir verilmedi.';
  @override
  String get sampleReminder =>
      'Yalnızca örnek önizleme. Hatırlatıcı kurulmadı.';
  @override
  String get reductionTitle => 'Önerilen küçültme';
  @override
  String get reductionBody =>
      'Küçültülen emri vermeden önce yeniden kontrol edin. Bu örnek, limitleri yeniden hesaplayamaz veya bekleme süresini kaldıramaz.';
  @override
  String get close => 'Kapat';
  @override
  String cooldown(String minutes) =>
      'Hiçbir limit aşılmıyor. Bekleme kuralınıza göre $minutes dakika daha beklemelisiniz.';
  @override
  String reduction(String value, String risk) =>
      'Limite uyan en büyük emir $value USDT, riske edilen tutar $risk USDT. Engellenen emir hiçbir koşulda verilemez.';
  @override
  String limitName(LimitCheck check) => switch (check.id) {
    'daily-losing-trades' => const TurkishAppStrings().losingTradesToday,
    'cooldown' => 'Zarardan sonra bekleme süresi',
    'risk' => 'Toplam açık risk',
    'daily' => 'Günlük zarar limiti',
    'size' => 'Pozisyon büyüklüğü',
    'small-cap' => 'Küçük piyasa değerli coinlerin toplamı',
    'positions' => 'Açık pozisyonlar',
    _ => check.name,
  };
  @override
  String limitValue(String value) =>
      value.replaceAll(' of ', ' / ').replaceAll(' min waited', ' dk beklendi');
}
