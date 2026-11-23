import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/features/home/home_screen.dart';
import 'package:metron/router.dart';

import 'live_test_app.dart';

void main() {
  testWidgets(
    'Startup shows loading, an error with retry, then the verified account',
    (tester) async {
      final fake = FakeBackend()
        ..sessionWait = Completer<void>()
        ..sessionFail = true;
      await bootLive(tester, backend: fake, settle: false);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.byType(HomeScreen), findsNothing);
      fake.sessionWait!.complete();
      await tester.pumpAndSettle();
      expect(find.text('Retry'), findsOneWidget);
      fake.sessionFail = false;
      await tester.tap(find.text('Retry'));
      await tester.pumpAndSettle();
      expect(find.byType(HomeScreen), findsOneWidget);
    },
  );
  testWidgets('Account errors retry; empty positions and markets stay empty', (
    tester,
  ) async {
    final fake = FakeBackend()
      ..accountsFail = true
      ..closed = true
      ..emptyMarkets = true;
    final container = await bootLive(tester, backend: fake);
    expect(find.text('Retry'), findsOneWidget);
    fake.accountsFail = false;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();
    expect(find.text('No open positions.'), findsOneWidget);
    container.read(routerProvider).go('/markets');
    await tester.pumpAndSettle();
    expect(find.text('No market data available yet.'), findsOneWidget);
    expect(find.text('LIVE'), findsNothing);
  });
  testWidgets(
    'Home polling pauses on hidden branches and background lifecycle',
    (tester) async {
      final fake = FakeBackend();
      final container = await bootLive(tester, backend: fake);
      int reads() =>
          fake.requests.where((r) => r.url.path.endsWith('/portfolio')).length;
      container.read(routerProvider).go('/markets');
      await tester.pumpAndSettle();
      final hiddenReads = reads();
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();
      expect(reads(), hiddenReads);
      container.read(routerProvider).go('/');
      await tester.pumpAndSettle();
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.hidden);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
      final pausedReads = reads();
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();
      expect(reads(), pausedReads);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.hidden);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.inactive);
      tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
      await tester.pump(const Duration(seconds: 5));
      await tester.pumpAndSettle();
      expect(reads(), greaterThan(pausedReads));
    },
  );
}
