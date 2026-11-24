import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/router.dart';

import '../live_test_app.dart';
import 'check_screen_test.dart' show openCheck;

void main() {
  for (final environment in ['mainnet', 'testnet']) {
    testWidgets('$environment live accounts display badges and place safely', (
      tester,
    ) async {
      final fake = FakeBackend();
      fake.accounts[0].addAll({
        'type': 'live',
        'environment': environment,
        'key_hint': '…WXYZ',
      });
      await bootLive(tester, backend: fake);
      expect(find.text('…WXYZ'), findsOneWidget);
      expect(find.widgetWithText(Chip, 'LIVE'), findsOneWidget);
      expect(
        find.widgetWithText(Chip, environment.toUpperCase()),
        findsOneWidget,
      );
      expect(find.text('Connect Binance keys on the web.'), findsOneWidget);
      await openCheck(tester);
      if (environment == 'testnet') {
        expect(find.text('TESTNET: orders use test funds.'), findsOneWidget);
      }
      await tester.tap(find.text('Place live order'));
      await tester.pumpAndSettle();
      if (environment == 'mainnet') {
        expect(find.text('Real money order'), findsOneWidget);
        expect(fake.requests.any((r) => r.url.path == '/v1/orders'), isFalse);
        await tester.tap(find.text('Cancel'));
        await tester.pumpAndSettle();
        expect(fake.requests.any((r) => r.url.path == '/v1/orders'), isFalse);
        await tester.tap(find.text('Place live order'));
        await tester.pumpAndSettle();
        await tester.tap(find.byKey(const ValueKey('confirm-live')));
        await tester.pumpAndSettle();
      }
      final order = jsonDecode(
        fake.requests.lastWhere((r) => r.url.path == '/v1/orders').body,
      );
      expect(order['confirm_live'], environment == 'mainnet' ? true : isNull);
    });
  }
  testWidgets(
    'Server confirmation requirement is shown without silent resubmission',
    (tester) async {
      final fake = FakeBackend()..confirmationRequired = true;
      fake.accounts[0].addAll({'type': 'live', 'environment': 'mainnet'});
      await bootLive(tester, backend: fake);
      await openCheck(tester);
      await tester.tap(find.text('Place live order'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('confirm-live')));
      await tester.pumpAndSettle();
      expect(
        find.text(
          'This MAINNET order uses real money. Confirm to send it to Binance.',
        ),
        findsOneWidget,
      );
      expect(fake.requests.where((r) => r.url.path == '/v1/orders').length, 1);
    },
  );
  testWidgets(
    'Failed exchange stop retries and shows returned status and close reason',
    (tester) async {
      final fake = FakeBackend();
      fake.position['exchange_stop'] = {
        'status': 'failed',
        'order_id': null,
        'error': 'Exchange unavailable',
      };
      fake.position['close_reason'] = 'exchange_stop';
      final container = await bootLive(tester, backend: fake);
      container.read(routerProvider).go('/positions/position-1');
      await tester.pumpAndSettle();
      expect(find.text('Failed'), findsOneWidget);
      expect(find.text('Exchange unavailable'), findsOneWidget);
      expect(find.text('Exchange stop'), findsNWidgets(2));
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();
      final retry = fake.requests.lastWhere(
        (r) => r.url.path.endsWith('/stop/retry'),
      );
      expect(retry.method, 'POST');
      expect(retry.body, '{}');
      expect(find.text('Placed'), findsOneWidget);
      expect(find.text('123'), findsOneWidget);
      expect(find.text('Retry'), findsNothing);
    },
  );
}
