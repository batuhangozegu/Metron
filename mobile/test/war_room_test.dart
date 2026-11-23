import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/router.dart';
import 'package:metron/widgets/widgets.dart';

import 'live_test_app.dart';

void main() {
  for (final route in [
    '/',
    '/markets',
    '/alerts',
    '/rules',
    '/check',
    '/check/blocked',
    '/positions/position-1',
  ]) {
    testWidgets('Live ticker sits above screen content on $route', (
      tester,
    ) async {
      final container = await bootLive(tester);
      container.read(routerProvider).go(route);
      await tester.pumpAndSettle();
      expect(find.byType(MarketTicker), findsOneWidget);
      expect(find.text('LIVE'), findsOneWidget);
      expect(
        tester.getBottomLeft(find.byType(MarketTicker)).dy,
        lessThanOrEqualTo(
          tester.getTopLeft(find.byType(CommandTitle).first).dy,
        ),
      );
      expect(tester.takeException(), isNull);
    });
  }
  testWidgets('Ticker still scrolls, pauses and resumes with live data', (
    tester,
  ) async {
    await bootLive(tester);
    final controller = tester
        .widget<ListView>(find.byKey(const ValueKey('ticker-quotes')))
        .controller!;
    await tester.pump(const Duration(seconds: 4));
    await tester.pump(const Duration(seconds: 1));
    expect(controller.offset, greaterThan(0));
    await tester.tap(find.byTooltip('Pause ticker'));
    await tester.pumpAndSettle();
    final offset = controller.offset;
    await tester.pump(const Duration(seconds: 8));
    await tester.pumpAndSettle();
    expect(controller.offset, offset);
    expect(find.byTooltip('Resume ticker'), findsOneWidget);
  });
  testWidgets('Font assets remain available offline', (tester) async {
    await bootLive(tester);
    expect(
      (await rootBundle.load('assets/fonts/Oswald-Variable.ttf')).lengthInBytes,
      greaterThan(100000),
    );
    expect(
      await rootBundle.loadString('assets/fonts/Oswald-OFL.txt'),
      contains('SIL OPEN FONT LICENSE'),
    );
  });
}
