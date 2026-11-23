import 'api_client.dart';

/// API decimal fields are strings, including nullable market values.
class TradingAccount {
  TradingAccount.fromJson(Json json)
    : id = json['id'] as String,
      name = json['name'] as String,
      market = json['market'] as String? ?? 'spot';
  final String id, name, market;
  bool get futures => market == 'futures';
}

class MarketQuote {
  MarketQuote.fromJson(Json json)
    : symbol = json['symbol'] as String,
      pair = '${json['base_asset']}/${json['quote_asset']}',
      coinClass = json['coin_class'] as String,
      price = json['price'] as String?,
      change = json['change_24h_percent'] as String?,
      atr = json['atr_percent'] as String?,
      stop = json['suggested_stop_percent'] as String?;
  final String symbol, pair, coinClass;
  final String? price, change, atr, stop;
}

class LivePosition {
  LivePosition.fromJson(this.json);
  final Json json;
  String get id => json['id'] as String;
  String get symbol => json['symbol'] as String;
  String get side => json['side'] as String;
  String value(String key) => json[key] as String? ?? '—';
}

class RiskCheck {
  RiskCheck(this.json);
  final Json json;
  List<Json> get rules => (json['checks'] as List).cast<Json>();
  bool get blocked =>
      json['allowed'] != true || rules.any((r) => r['status'] == 'fail');
  String value(String key) => json[key] as String? ?? '—';
}

const profileFields = [
  'risk_per_trade_percent',
  'target_risk_reward',
  'total_open_risk_limit_percent',
  'max_open_positions',
  'max_losing_trades_today',
  'daily_loss_limit_percent',
  'max_stopped_trades_this_week',
  'weekly_loss_limit_percent',
];
const profileCounts = {
  'max_open_positions',
  'max_losing_trades_today',
  'max_stopped_trades_this_week',
};

bool validPositiveDecimal(String value, String maximum) {
  if (!RegExp(r'^\d+(\.\d{1,8})?$').hasMatch(value)) return false;
  BigInt units(String v) {
    final parts = v.split('.');
    return BigInt.parse(
      parts[0] + (parts.length == 1 ? '' : parts[1]).padRight(8, '0'),
    );
  }

  final number = units(value);
  return number > BigInt.zero && number <= units(maximum);
}

bool validRiskCount(String value) =>
    RegExp(r'^\d+$').hasMatch(value) &&
    BigInt.parse(value) > BigInt.zero &&
    BigInt.parse(value) <= BigInt.from(4294967295);
