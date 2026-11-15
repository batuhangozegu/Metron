import '../../data/models.dart';

/// Feature-local English copy.
abstract final class MarketsStrings {
  static const live = 'Live';
  static const search = 'Search pairs';
  static const coinClass = 'Coin class';
  static const all = 'All';
  static const stop = 'stop';
  static const atr = 'ATR';
  static const noResults = 'No pairs found. Try another search or coin class.';
  static const barStart = 'The bar shows daily movement from ';
  static const barTo = ' to ';
  static const barEnd = '. Tap a pair to check a trade.';
  static String className(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large-cap',
    CoinClass.smallCap => 'Small-cap',
  };
}
