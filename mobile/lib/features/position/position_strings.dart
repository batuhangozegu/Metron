import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';
import '../../data/models.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class PositionStrings {
  static PositionStrings of(BuildContext context) =>
      CatalogScope.maybeOf<PositionStrings>(context) ??
      const EnglishPositionStrings();
  String get unrealized;
  String get price;
  String get stop;
  String get entry;
  String get target;
  String get now;
  String get stopOnExchange;
  String get stopMissing;
  String get away;
  String get quantity;
  String get valueNow;
  String get moneyAtRisk;
  String get share;
  String get opened;
  String get stopRule;
  String get tighten;
  String get close;
  String get dismiss;
  String get previewOnly;
  String get notFound;
  String get detailsMissing;
  String get priceMissing;
  String coinClass(CoinClass value);
  String openedAt(DateTime value);
}

final class EnglishPositionStrings implements PositionStrings {
  const EnglishPositionStrings();

  @override
  String get unrealized => 'Unrealized result';
  @override
  String get price => 'price';
  @override
  String get stop => 'Stop';
  @override
  String get entry => 'Entry';
  @override
  String get target => 'Target';
  @override
  String get now => 'Now';
  @override
  String get stopOnExchange => 'The stop order is on the exchange,';
  @override
  String get stopMissing => 'The stop order is not on the exchange.';
  @override
  String get away => 'away.';
  @override
  String get quantity => 'Quantity';
  @override
  String get valueNow => 'Value now';
  @override
  String get moneyAtRisk => 'Money at risk';
  @override
  String get share => 'Share of account';
  @override
  String get opened => 'Opened';
  @override
  String get stopRule =>
      'A stop can move toward the price, never away from it.';
  @override
  String get tighten => 'Tighten stop';
  @override
  String get close => 'Close position';
  @override
  String get dismiss => 'Dismiss';
  @override
  String get previewOnly =>
      'This is a sample position. Trading actions are not connected yet.';
  @override
  String get notFound => 'Position not found.';
  @override
  String get detailsMissing =>
      'Detailed position data is not available for this sample.';
  @override
  String get priceMissing => 'Current price unavailable';

  @override
  String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large cap',
    CoinClass.smallCap => 'Small cap',
  };

  @override
  String openedAt(DateTime value) {
    final utc = value.toUtc();
    const months = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec',
    ];
    String two(int n) => n.toString().padLeft(2, '0');
    return '${utc.day} ${months[utc.month - 1]} ${two(utc.hour)}:${two(utc.minute)} UTC';
  }
}

final class TurkishPositionStrings implements PositionStrings {
  const TurkishPositionStrings();
  @override
  String get unrealized => 'Gerçekleşmemiş kâr/zarar';
  @override
  String get price => 'fiyat';
  @override
  String get stop => 'Stop';
  @override
  String get entry => 'Giriş';
  @override
  String get target => 'Hedef';
  @override
  String get now => 'Şimdi';
  @override
  String get stopOnExchange => 'Stop emri borsada,';
  @override
  String get stopMissing => 'Stop emri borsada değil.';
  @override
  String get away => 'uzakta.';
  @override
  String get quantity => 'Miktar';
  @override
  String get valueNow => 'Güncel değer';
  @override
  String get moneyAtRisk => 'Riske edilen tutar';
  @override
  String get share => 'Hesaptaki payı';
  @override
  String get opened => 'Açılış';
  @override
  String get stopRule =>
      'Stop fiyatını yalnızca piyasa fiyatına yaklaştırabilirsiniz, uzaklaştıramazsınız.';
  @override
  String get tighten => 'Stopu yaklaştır';
  @override
  String get close => 'Pozisyonu kapat';
  @override
  String get dismiss => 'Kapat';
  @override
  String get previewOnly =>
      'Bu bir örnek pozisyondur. İşlem fonksiyonları henüz bağlı değil.';
  @override
  String get notFound => 'Pozisyon bulunamadı.';
  @override
  String get detailsMissing => 'Bu örnek için ayrıntılı pozisyon verisi yok.';
  @override
  String get priceMissing => 'Güncel fiyat alınamıyor';
  @override
  String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Büyük piyasa değerli',
    CoinClass.smallCap => 'Küçük piyasa değerli',
  };
  @override
  String openedAt(DateTime value) {
    final utc = value.toUtc();
    const months = [
      'Oca',
      'Şub',
      'Mar',
      'Nis',
      'May',
      'Haz',
      'Tem',
      'Ağu',
      'Eyl',
      'Eki',
      'Kas',
      'Ara',
    ];
    String two(int n) => n.toString().padLeft(2, '0');
    return '${utc.day} ${months[utc.month - 1]} ${two(utc.hour)}:${two(utc.minute)} UTC';
  }
}
