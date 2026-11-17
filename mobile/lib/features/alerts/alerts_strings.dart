import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';
import '../../data/models.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class AlertsStrings {
  static AlertsStrings of(BuildContext context) =>
      CatalogScope.maybeOf<AlertsStrings>(context) ??
      const EnglishAlertsStrings();
  String get paper;
  String get today;
  String get priceAlarms;
  String get addAlarm;
  String get noEvents;
  String get noAlarms;
  String get now;
  String get unavailable;
  String get pair;
  String get condition;
  String get targetPrice;
  String get invalidPrice;
  String get cancel;
  String get sampleOnly;
  String get noPrices;
  String direction(AlarmDirection direction);
  String eventTitle(AlertEvent event);
  String eventBody(AlertEvent event);
}

final class EnglishAlertsStrings implements AlertsStrings {
  const EnglishAlertsStrings();

  @override
  String get paper => 'PAPER';
  @override
  String get today => 'Today';
  @override
  String get priceAlarms => 'Price alarms';
  @override
  String get addAlarm => 'Add an alarm';
  @override
  String get noEvents => 'No events today.';
  @override
  String get noAlarms => 'No price alarms yet.';
  @override
  String get now => 'now ';
  @override
  String get unavailable => '—';
  @override
  String get pair => 'Pair';
  @override
  String get condition => 'Condition';
  @override
  String get targetPrice => 'Target price';
  @override
  String get invalidPrice => 'Enter a positive price.';
  @override
  String get cancel => 'Cancel';
  @override
  String get sampleOnly => 'Sample alarm · kept for this session only.';
  @override
  String get noPrices => 'No pairs available.';
  @override
  String direction(AlarmDirection direction) => switch (direction) {
    AlarmDirection.above => 'rises above',
    AlarmDirection.below => 'falls below',
  };
  @override
  String eventTitle(AlertEvent event) => event.title;
  @override
  String eventBody(AlertEvent event) => event.body;
}

final class TurkishAlertsStrings implements AlertsStrings {
  const TurkishAlertsStrings();
  @override
  String get paper => 'SANAL';
  @override
  String get today => 'Bugün';
  @override
  String get priceAlarms => 'Fiyat alarmları';
  @override
  String get addAlarm => 'Alarm ekle';
  @override
  String get noEvents => 'Bugün bildirim yok.';
  @override
  String get noAlarms => 'Henüz fiyat alarmı yok.';
  @override
  String get now => 'şimdi ';
  @override
  String get unavailable => '—';
  @override
  String get pair => 'Parite';
  @override
  String get condition => 'Koşul';
  @override
  String get targetPrice => 'Hedef fiyat';
  @override
  String get invalidPrice => 'Sıfırdan büyük bir fiyat girin.';
  @override
  String get cancel => 'İptal';
  @override
  String get sampleOnly => 'Örnek alarm · yalnızca bu oturumda saklanır.';
  @override
  String get noPrices => 'Kullanılabilir parite yok.';
  @override
  String direction(AlarmDirection direction) => switch (direction) {
    AlarmDirection.above => 'üzerine çıkarsa',
    AlarmDirection.below => 'altına düşerse',
  };
  @override
  String eventTitle(AlertEvent event) => switch (event.kind) {
    AlertKind.cooldown => 'Bekleme süresi devam ediyor',
    AlertKind.stopTriggered => event.title.replaceFirst(
      'Stop triggered',
      'Stop tetiklendi',
    ),
    AlertKind.priceAlarm => event.title.replaceFirst(
      'rose above',
      'şu fiyatın üzerine çıktı:',
    ),
    AlertKind.orderBlocked => 'Emir engellendi',
  };
  @override
  String eventBody(AlertEvent event) => switch (event.kind) {
    AlertKind.cooldown => event.body.replaceFirst(
      'New orders are a broken rule until ',
      'Yeni emirler şu saate kadar kural ihlalidir: ',
    ),
    AlertKind.stopTriggered => event.body.replaceFirst(
      'Closed at ',
      'Kapanış fiyatı: ',
    ),
    AlertKind.priceAlarm => 'Fiyat alarmınız.',
    AlertKind.orderBlocked =>
      event.body
          .replaceFirst('Moving the ', '')
          .replaceFirst(
            ' stop away from the price breaks ',
            ' stopunu fiyattan uzaklaştırmak şu kuralı ihlal eder: ',
          )
          .replaceFirst('A stop is never widened', 'Stop asla uzaklaştırılmaz'),
  };
}
