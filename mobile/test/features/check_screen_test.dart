import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/features/check/check_screen.dart';
import 'package:metron/features/check_blocked/check_blocked_screen.dart';
import 'package:metron/features/position/position_screen.dart';

import '../live_test_app.dart';

Future<void> openCheck(WidgetTester tester) async {
  await tester.tap(find.text('Check a new order'));
  await tester.pumpAndSettle();
  await tester.tap(find.text('Run risk check'));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('Check then place sends string entry and server-sized order', (
    tester,
  ) async {
    final fake = FakeBackend();
    await bootLive(tester, backend: fake);
    await openCheck(tester);
    expect(find.byType(CheckScreen), findsOneWidget);
    final preview = fake.requests.lastWhere(
      (r) => r.url.path.endsWith('/risk/check'),
    );
    expect(jsonDecode(preview.body)['entry_price'], '100.12345678');
    await tester.scrollUntilVisible(
      find.text('Place paper order'),
      350,
      scrollable: find.byType(Scrollable).last,
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Place paper order'));
    await tester.pumpAndSettle();
    final order = fake.requests.lastWhere((r) => r.url.path == '/v1/orders');
    expect(jsonDecode(order.body), {
      'account_id': 'account-1',
      'symbol': 'BTCUSDT',
      'side': 'long',
    });
    expect(order.headers['content-type'], 'application/json');
    expect(find.byType(PositionScreen), findsOneWidget);
  });
  for (final rejectOrder in [false, true]) {
    testWidgets(
      'Real failed rules route to blocked; order rejection=$rejectOrder',
      (tester) async {
        final fake = FakeBackend()
          ..blocked = !rejectOrder
          ..rejectOrder = rejectOrder;
        await bootLive(tester, backend: fake);
        await openCheck(tester);
        if (rejectOrder) {
          await tester.scrollUntilVisible(
            find.text('Place paper order'),
            350,
            scrollable: find.byType(Scrollable).last,
          );
          await tester.pumpAndSettle();
          await tester.tap(find.text('Place paper order'));
          await tester.pumpAndSettle();
        }
        expect(find.byType(CheckBlockedScreen), findsOneWidget);
        expect(
          GoRouterState.of(tester.element(find.byType(CheckBlockedScreen)))
              .uri
              .path,
          '/check/blocked',
        );
        await tester.scrollUntilVisible(
          find.text('Daily loss reached; stop trading.'),
          300,
          scrollable: find.byType(Scrollable).last,
        );
        expect(find.text('Failed'), findsOneWidget);
        expect(find.text('Place paper order'), findsNothing);
        if (!rejectOrder) {
          expect(fake.requests.any((r) => r.url.path == '/v1/orders'), isFalse);
        }
      },
    );
  }
  testWidgets('Futures short preview and order retain leverage', (
    tester,
  ) async {
    final fake = FakeBackend();
    fake.accounts[0]['market'] = 'futures';
    await bootLive(tester, backend: fake);
    await tester.tap(find.text('Check a new order'));
    await tester.pumpAndSettle();
    await tester.tap(find.byType(DropdownButtonFormField<String>).last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Short').last);
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const ValueKey('leverage')), '10');
    await tester.tap(find.text('Run risk check'));
    await tester.pumpAndSettle();
    final body = jsonDecode(
      fake.requests.lastWhere((r) => r.url.path.endsWith('/risk/check')).body,
    );
    expect(body['side'], 'short');
    expect(body['leverage'], 10);
    await tester.scrollUntilVisible(
      find.text('Place paper order'),
      350,
      scrollable: find.byType(Scrollable).last,
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Place paper order'));
    await tester.pumpAndSettle();
    expect(
      jsonDecode(
        fake.requests.lastWhere((r) => r.url.path == '/v1/orders').body,
      )['leverage'],
      10,
    );
  });
}
