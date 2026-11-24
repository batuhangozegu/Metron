import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/router.dart';

import '../live_test_app.dart';

void main() {
  testWidgets(
    'Create validates price, uses market symbol and cancels via 204',
    (tester) async {
      final fake = FakeBackend();
      final container = await bootLive(tester, backend: fake);
      container.read(routerProvider).go('/alerts');
      await tester.pumpAndSettle();
      expect(find.text('No price alerts yet.'), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('add-alert')));
      await tester.pumpAndSettle();
      await tester.enterText(find.byKey(const ValueKey('alert-price')), '0');
      await tester.tap(find.byKey(const ValueKey('save-alert')));
      await tester.pumpAndSettle();
      expect(find.text('Enter a positive decimal price.'), findsOneWidget);
      await tester.tap(find.byType(DropdownButtonFormField<String>).last);
      await tester.pumpAndSettle();
      await tester.tap(find.text('Below').last);
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const ValueKey('alert-price')),
        '65000.12345678',
      );
      await tester.enterText(
        find.byKey(const ValueKey('alert-note')),
        'My target',
      );
      await tester.tap(find.byKey(const ValueKey('save-alert')));
      await tester.pumpAndSettle();
      final post = fake.requests.lastWhere(
        (r) => r.url.path == '/v1/alerts' && r.method == 'POST',
      );
      expect(jsonDecode(post.body), {
        'symbol': 'BTCUSDT',
        'direction': 'below',
        'price': '65000.12345678',
        'note': 'My target',
      });
      expect(find.text('My target'), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('cancel-alert-0')));
      await tester.pumpAndSettle();
      expect(find.text('Cancelled'), findsOneWidget);
      expect(
        fake.requests.lastWhere((r) => r.method == 'DELETE').url.path,
        '/v1/alerts/alert-0',
      );
    },
  );

  testWidgets('Alert list retries and create displays the active limit', (
    tester,
  ) async {
    final fake = FakeBackend()..alertsFail = true;
    final container = await bootLive(tester, backend: fake);
    container.read(routerProvider).go('/alerts');
    await tester.pumpAndSettle();
    fake.alertsFail = false;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('add-alert')));
    await tester.pumpAndSettle();
    fake.alertLimit = true;
    await tester.enterText(find.byKey(const ValueKey('alert-price')), '1');
    await tester.tap(find.byKey(const ValueKey('save-alert')));
    await tester.pumpAndSettle();
    expect(find.text('You can have at most 50 active alerts.'), findsOneWidget);
  });

  testWidgets(
    'Foreground polls every 15 seconds across tabs, deduplicates and pauses',
    (tester) async {
      final fake = FakeBackend();
      final container = await bootLive(tester, backend: fake);
      container.read(routerProvider).go('/markets');
      await tester.pumpAndSettle();
      fake.alerts.add({
        'id': 'trigger',
        'symbol': 'BTC',
        'direction': 'above',
        'price': '100',
        'status': 'triggered',
        'created_at': '2026-11-24T00:00:00Z',
        'triggered_at': '2026-11-24T01:00:00Z',
        'triggered_price': '101',
      });
      int polls() => fake.requests
          .where((r) => r.url.queryParameters.containsKey('triggered_since'))
          .length;
      await tester.pump(const Duration(seconds: 14));
      expect(polls(), 0);
      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
      expect(polls(), 1);
      expect(find.text('BTC alert triggered at 101'), findsOneWidget);
      final cursor = fake.requests
          .lastWhere(
            (r) => r.url.queryParameters.containsKey('triggered_since'),
          )
          .url
          .queryParameters['triggered_since'];
      expect(DateTime.tryParse(cursor!), isNotNull);
      await tester.pump(const Duration(seconds: 15));
      await tester.pumpAndSettle();
      expect(polls(), 2);
      expect(find.text('BTC alert triggered at 101'), findsNothing);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.hidden);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
      await tester.pump(const Duration(seconds: 30));
      expect(polls(), 2);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.hidden);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
      await tester.pump(const Duration(seconds: 15));
      await tester.pumpAndSettle();
      expect(polls(), 3);
    },
  );
}
