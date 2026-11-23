import 'package:flutter/widgets.dart';

import '../data/api_client.dart';

/// Copy shared by the live account, transport and order flows (EN/TR).
class LiveStrings {
  LiveStrings(this.tr);
  final bool tr;
  static LiveStrings of(BuildContext context) =>
      LiveStrings(Localizations.localeOf(context).languageCode == 'tr');
  String text(String en, String turkish) => tr ? turkish : en;
  String get loading => text('Loading…', 'Yükleniyor…');
  String get retry => text('Retry', 'Tekrar dene');
  String get save => text('Save changes', 'Değişiklikleri kaydet');
  String get saved => text('Changes saved.', 'Değişiklikler kaydedildi.');
  String get cancel => text('Cancel', 'İptal');
  String get signOut => text('Sign out', 'Çıkış yap');
  String get account => text('Paper account', 'Kağıt hesap');
  String get create => text('Create paper account', 'Kağıt hesap oluştur');
  String get name => text('Account name', 'Hesap adı');
  String get balance =>
      text('Starting balance, USDT', 'Başlangıç bakiyesi, USDT');
  String get noAccounts => text(
    'Create a USDT paper account to start trading.',
    'İşleme başlamak için bir USDT kağıt hesabı oluşturun.',
  );
  String get invalidAccount => text(
    'Use a name of 1–100 characters and a balance greater than 0 and at most 10,000,000, with up to 8 decimal places.',
    '1–100 karakterlik ad ve 0’dan büyük, en fazla 10.000.000 olan, en fazla 8 ondalık basamaklı bakiye kullanın.',
  );
  String get emptyPositions => text('No open positions.', 'Açık pozisyon yok.');
  String get missingPosition => text(
    'Position is closed or unavailable.',
    'Pozisyon kapalı veya kullanılamıyor.',
  );
  String get close => text('Close position', 'Pozisyonu kapat');
  String get confirmClose => text(
    'Close this paper position at the latest market price?',
    'Bu kağıt pozisyon son piyasa fiyatından kapatılsın mı?',
  );
  String get equity => text('Equity', 'Öz sermaye');
  String get cash => text('Cash', 'Nakit');
  String get returns => text('Return', 'Getiri');
  String get drawdown => text('Maximum drawdown', 'Maksimum düşüş');
  String get positions => text('Open positions', 'Açık pozisyonlar');
  String get noMarkets =>
      text('No market data available yet.', 'Henüz piyasa verisi yok.');
  String get feed => text('LIVE', 'CANLI');
  String get unavailable => text('Data unavailable', 'Veri kullanılamıyor');
  String get comingSoon =>
      text('Price alarms — coming soon', 'Fiyat alarmları — yakında');
  String get alarmsNote => text(
    'Price alarms and alert history are not available yet.',
    'Fiyat alarmları ve bildirim geçmişi henüz kullanılamıyor.',
  );
  String get runCheck => text('Run risk check', 'Risk kontrolü yap');
  String get side => text('Side', 'Yön');
  String get long => text('Long', 'Uzun');
  String get short => text('Short', 'Kısa');
  String get leverage => text('Leverage (1–125)', 'Kaldıraç (1–125)');
  String get invalidLeverage => text(
    'Enter a whole leverage from 1 to 125.',
    '1 ile 125 arasında tam sayı kaldıraç girin.',
  );
  String get blocked => text('ORDER BLOCKED', 'EMİR ENGELLENDİ');
  String get caution => text('REVIEW CAUTIONS', 'UYARILARI İNCELEYİN');
  String get allowed => text('WITHIN YOUR LIMITS', 'LİMİTLERİNİZ DAHİLİNDE');
  String get noCheck => text(
    'Run a risk check before placing an order.',
    'Emir vermeden önce risk kontrolü yapın.',
  );
  String get suggestedStop => text('Suggested stop', 'Önerilen stop');
  String get risk => text('Risk, USDT', 'Risk, USDT');
  String get rules => text('Rule results', 'Kural sonuçları');
  String status(String status) => switch (status) {
    'fail' => text('Failed', 'Başarısız'),
    'caution' => text('Caution', 'Uyarı'),
    _ => text('Passed', 'Başarılı'),
  };
  String ruleValue(String key) => switch (key) {
    'current' || 'current_percent' => text('Current', 'Mevcut'),
    'proposed' ||
    'proposed_percent' ||
    'projected' => text('After order', 'Emir sonrası'),
    'limit' || 'limit_percent' => text('Limit', 'Limit'),
    'requested_percent' => text('Requested risk, %', 'İstenen risk, %'),
    'max_percent' => text('Maximum, %', 'Maksimum, %'),
    'effective_percent' => text('Effective risk, %', 'Etkin risk, %'),
    'distance' => text('Stop distance', 'Stop mesafesi'),
    'suggested_distance' => text('Suggested distance', 'Önerilen mesafe'),
    'atr' => 'ATR',
    'multiplier' => text('ATR multiplier', 'ATR çarpanı'),
    'step' => text('Quantity step', 'Miktar adımı'),
    'size' => text('Size', 'Miktar'),
    'budget' => text('Risk budget', 'Risk bütçesi'),
    'actual_risk' => risk,
    'price' => text('Price', 'Fiyat'),
    'risk_reward' => text('Risk/reward', 'Risk/getiri'),
    'stop' => 'Stop',
    'liquidation_price' => text('Estimated liquidation', 'Tahmini tasfiye'),
    'leverage' => text('Leverage', 'Kaldıraç'),
    'maintenance_margin_rate' => text(
      'Maintenance margin rate',
      'Sürdürme marjini oranı',
    ),
    'buffer_rate' => text('Buffer rate', 'Güvenlik payı oranı'),
    'notional' => text('Order value', 'Emir tutarı'),
    'equity' => equity,
    _ => key.replaceAll('_', ' '),
  };
  String get invalidDecimal => text(
    'Enter a value greater than 0 and at most 100, with up to 8 decimal places.',
    "0'dan büyük, en fazla 100 olan ve en fazla 8 ondalık basamak içeren bir değer girin.",
  );
  String get invalidCount => text(
    'Enter a whole number from 1 to 4,294,967,295.',
    '1 ile 4.294.967.295 arasında bir tam sayı girin.',
  );
  String get defaults =>
      text('Defaults, not saved yet', 'Varsayılanlar, henüz kaydedilmedi');
  String profileLabel(String field) => switch (field) {
    'risk_per_trade_percent' => text(
      'Default risk per trade, %',
      'Varsayılan işlem başına risk, %',
    ),
    'target_risk_reward' => text('Target risk/reward', 'Hedef risk/getiri'),
    'total_open_risk_limit_percent' => text(
      'Total open risk limit, %',
      'Toplam açık risk limiti, %',
    ),
    'max_open_positions' => text(
      'Open positions at once',
      'Aynı anda açık pozisyon',
    ),
    'max_losing_trades_today' => text(
      'Losing trades per day',
      'Günlük zararlı işlem sayısı',
    ),
    'daily_loss_limit_percent' => text(
      'Daily loss cap, %',
      'Günlük zarar limiti, %',
    ),
    'max_stopped_trades_this_week' => text(
      'Stopped trades per week',
      'Haftalık stopla kapanan işlem sayısı',
    ),
    _ => text('Weekly loss cap, %', 'Haftalık zarar limiti, %'),
  };

  String error(Object error, {String context = ''}) {
    if (error is! ApiError || error.status == 0) {
      return text(
        'Could not reach the service. Check your connection and try again.',
        'Servise ulaşılamadı. Bağlantınızı kontrol edip tekrar deneyin.',
      );
    }
    final data = error.data;
    final wire = data['error'];
    if (wire is Map) {
      final message = wire[tr ? 'message_tr' : 'message_en'];
      if (message is String && message.isNotEmpty) return message;
    }
    if (data['code'] != null) {
      final message = tr ? data['error_tr'] : wire;
      if (message is String && message.isNotEmpty) return message;
    }
    return switch (error.status) {
      400 => text(
        'Check the fields and try again.',
        'Alanları kontrol edip tekrar deneyin.',
      ),
      401 =>
        context == 'code'
            ? text(
                'The code is invalid, expired or already used. Use a fresh code; if sign-in has expired, start again.',
                'Kod geçersiz, süresi dolmuş veya kullanılmış. Yeni kod kullanın; giriş süresi dolduysa yeniden başlayın.',
              )
            : context == 'login'
            ? text(
                'The email or password is incorrect.',
                'E-posta veya parola yanlış.',
              )
            : text(
                'Your session has expired. Sign in again.',
                'Oturumunuz sona erdi. Tekrar giriş yapın.',
              ),
      409 => text(
        'An account with this email already exists. Sign in instead.',
        'Bu e-posta ile bir hesap zaten var. Giriş yapın.',
      ),
      429 => text(
        'Too many attempts. Try again in 900 seconds.',
        'Çok fazla deneme. 900 saniye sonra tekrar deneyin.',
      ),
      503 => text(
        'The service is temporarily unavailable. Please try again shortly.',
        'Servis geçici olarak kullanılamıyor. Biraz sonra tekrar deneyin.',
      ),
      _ => text(
        'The request could not be completed. Please try again.',
        'İstek tamamlanamadı. Tekrar deneyin.',
      ),
    };
  }
}
