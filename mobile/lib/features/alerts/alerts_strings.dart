import '../../data/models.dart';

abstract final class AlertsStrings {
  static const paper = 'PAPER';
  static const today = 'Today';
  static const priceAlarms = 'Price alarms';
  static const addAlarm = 'Add an alarm';
  static const noEvents = 'No events today.';
  static const noAlarms = 'No price alarms yet.';
  static const now = 'now ';
  static const unavailable = '—';
  static const pair = 'Pair';
  static const condition = 'Condition';
  static const targetPrice = 'Target price';
  static const invalidPrice = 'Enter a positive price.';
  static const cancel = 'Cancel';
  static const sampleOnly = 'Sample alarm · kept for this session only.';
  static const noPrices = 'No pairs available.';
  static String direction(AlarmDirection direction) => switch (direction) {
    AlarmDirection.above => 'rises above',
    AlarmDirection.below => 'falls below',
  };
}

/// Group decimal strings without rounding or financial arithmetic.
String formatAlarmPrice(String value) {
  final parts = value.split('.');
  final whole = parts.first.replaceAllMapped(
    RegExp(r'\B(?=(\d{3})+(?!\d))'),
    (_) => ',',
  );
  return '$whole.${parts.length > 1 ? parts[1].padRight(2, '0') : '00'}';
}
