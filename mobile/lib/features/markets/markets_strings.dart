import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';
import '../../data/models.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class MarketsStrings {
  static MarketsStrings of(BuildContext context) =>
      CatalogScope.maybeOf<MarketsStrings>(context) ??
      const EnglishMarketsStrings();
  String get live;
  String get search;
  String get coinClass;
  String get all;
  String get stop;
  String get atr;
  String get noResults;
  String get barStart;
  String get barTo;
  String get barEnd;
  String className(CoinClass value);
}

final class EnglishMarketsStrings implements MarketsStrings {
  const EnglishMarketsStrings();

  @override
  String get live => 'Live';
  @override
  String get search => 'Search pairs';
  @override
  String get coinClass => 'Coin class';
  @override
  String get all => 'All';
  @override
  String get stop => 'stop';
  @override
  String get atr => 'ATR';
  @override
  String get noResults => 'No pairs found. Try another search or coin class.';
  @override
  String get barStart => 'The bar shows daily movement from ';
  @override
  String get barTo => ' to ';
  @override
  String get barEnd => '. Tap a pair to check a trade.';
  @override
  String className(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large-cap',
    CoinClass.smallCap => 'Small-cap',
  };
}

final class TurkishMarketsStrings implements MarketsStrings {
  const TurkishMarketsStrings();
  @override
  String get live => 'Canlı';
  @override
  String get search => 'Parite ara';
  @override
  String get coinClass => 'Coin grubu';
  @override
  String get all => 'Tümü';
  @override
  String get stop => 'stop';
  @override
  String get atr => 'ATR';
  @override
  String get noResults =>
      'Parite bulunamadı. Başka bir arama veya coin grubu deneyin.';
  @override
  String get barStart => 'Çubuk günlük hareketi gösterir: ';
  @override
  String get barTo => ' → ';
  @override
  String get barEnd => '. İşlemi kontrol etmek için bir pariteye dokunun.';
  @override
  String className(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Büyük piyasa değerli',
    CoinClass.smallCap => 'Küçük piyasa değerli',
  };
}
