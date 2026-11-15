import '../../data/models.dart';

abstract final class PositionStrings {
  static const unrealized = 'Unrealized result';
  static const price = 'price';
  static const stop = 'Stop';
  static const entry = 'Entry';
  static const target = 'Target';
  static const now = 'Now';
  static const stopOnExchange = 'The stop order is on the exchange,';
  static const stopMissing = 'The stop order is not on the exchange.';
  static const away = 'away.';
  static const quantity = 'Quantity';
  static const valueNow = 'Value now';
  static const moneyAtRisk = 'Money at risk';
  static const share = 'Share of account';
  static const opened = 'Opened';
  static const stopRule =
      'A stop can move toward the price, never away from it.';
  static const tighten = 'Tighten stop';
  static const close = 'Close position';
  static const dismiss = 'Dismiss';
  static const previewOnly =
      'This is a sample position. Trading actions are not connected yet.';
  static const notFound = 'Position not found.';
  static const detailsMissing =
      'Detailed position data is not available for this sample.';
  static const priceMissing = 'Current price unavailable';

  static String coinClass(CoinClass value) => switch (value) {
    CoinClass.btcEth => 'BTC/ETH',
    CoinClass.largeCap => 'Large cap',
    CoinClass.smallCap => 'Small cap',
  };

  static String openedAt(DateTime value) {
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
