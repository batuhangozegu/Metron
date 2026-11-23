import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/router.dart';

import '../live_test_app.dart';

void main() {
  testWidgets(
    'Markets uses live data, search, periodic refresh and error retry',
    (tester) async {
      final fake = FakeBackend();
      final container = await bootLive(tester, backend: fake);
      container.read(routerProvider).go('/markets');
      await tester.pumpAndSettle();
      expect(find.text('BTC/USDT'), findsOneWidget);
      expect(find.text('LIVE'), findsOneWidget);
      await tester.enterText(find.byType(TextField), 'ETH');
      await tester.pumpAndSettle();
      expect(find.text('BTC/USDT'), findsNothing);
      expect(find.text('ETH/USDT'), findsOneWidget);
      final before = fake.requests
          .where((r) => r.url.path == '/v1/markets')
          .length;
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();
      expect(
        fake.requests.where((r) => r.url.path == '/v1/markets').length,
        greaterThan(before),
      );
      fake.marketsFail = true;
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();
      expect(find.text('Retry'), findsOneWidget);
      fake.marketsFail = false;
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();
      expect(find.text('ETH/USDT'), findsOneWidget);
    },
  );
}
