import 'live_strings.dart';

class AlertStrings extends LiveStrings {
  AlertStrings(super.tr);
  String get add => text('Add alert', 'Alarm ekle');
  String get empty => text('No price alerts yet.', 'Henüz fiyat alarmı yok.');
  String get symbol => text('Symbol', 'Sembol');
  String get direction => text('Condition', 'Koşul');
  String get price => text('Target price', 'Hedef fiyat');
  String get note => text('Note (optional)', 'Not (isteğe bağlı)');
  String get invalid => text(
    'Enter a positive decimal price.',
    'Pozitif bir ondalık fiyat girin.',
  );
  String get limit => text(
    'You can have at most 50 active alerts.',
    'En fazla 50 aktif alarmınız olabilir.',
  );
  String directionLabel(String direction) => direction == 'above'
      ? text('Above', 'Üzerinde')
      : text('Below', 'Altında');
  String alertStatus(String status) => switch (status) {
    'active' => text('Active', 'Aktif'),
    'triggered' => text('Triggered', 'Tetiklendi'),
    'cancelled' => text('Cancelled', 'İptal edildi'),
    _ => status,
  };
  String triggered(String symbol, String price) => text(
    '$symbol alert triggered at $price',
    '$symbol alarmı $price fiyatında tetiklendi',
  );
}
