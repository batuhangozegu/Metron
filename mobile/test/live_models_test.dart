import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/live_models.dart';

import 'live_test_app.dart';

void main() {
  test('Paper fields stay nullable and live environments retain key hints', () {
    final paper = TradingAccount.fromJson(accountJson);
    expect(paper.live, isFalse);
    expect(paper.environment, isNull);
    expect(paper.keyHint, isNull);
    expect(LivePosition.fromJson(positionJson).exchangeStop, isNull);
    final live = TradingAccount.fromJson({
      ...accountJson,
      'type': 'live',
      'environment': 'mainnet',
      'key_hint': '…WXYZ',
    });
    expect(live.mainnet, isTrue);
    expect(live.keyHint, '…WXYZ');
    final stop = LivePosition.fromJson({
      ...positionJson,
      'exchange_stop': {'status': 'filled', 'order_id': '123', 'error': null},
    }).exchangeStop!;
    expect(stop.status, 'filled');
    expect(stop.orderId, '123');
    expect(stop.error, isNull);
  });
  test('Alert retains decimal strings and nullable trigger metadata', () {
    final alert = Alert.fromJson({
      'id': 'alert',
      'symbol': 'BTC',
      'direction': 'below',
      'price': '65000.12345678',
      'note': null,
      'status': 'active',
      'created_at': '2026-11-24T00:00:00Z',
      'triggered_at': null,
      'triggered_price': null,
    });
    expect(alert.price, '65000.12345678');
    expect(alert.triggeredAt, isNull);
    expect(alert.triggeredPrice, isNull);
    expect(alert.note, isNull);
    expect(alert.direction, 'below');
  });
}
