import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/router.dart';

import '../live_test_app.dart';

void main() {
  testWidgets('Position close confirms, posts JSON and refreshes positions', (
    tester,
  ) async {
    final fake = FakeBackend();
    final container = await bootLive(tester, backend: fake);
    container.read(routerProvider).go('/positions/position-1');
    await tester.pumpAndSettle();
    expect(find.text('100.12345678'), findsOneWidget);
    await tester.tap(find.text('Close position'));
    await tester.pumpAndSettle();
    expect(fake.closed, isFalse);
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(fake.closed, isFalse);
    await tester.tap(find.text('Close position'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('confirm-close')));
    await tester.pumpAndSettle();
    final close = fake.requests.lastWhere((r) => r.url.path.endsWith('/close'));
    expect(close.method, 'POST');
    expect(close.body, '{}');
    expect(close.headers['content-type'], 'application/json');
    expect(fake.closed, isTrue);
    expect(find.text('No open positions.'), findsOneWidget);
  });
  testWidgets('Unknown or closed position has an empty state', (tester) async {
    final container = await bootLive(tester);
    container.read(routerProvider).go('/positions/missing');
    await tester.pumpAndSettle();
    expect(find.text('Position is closed or unavailable.'), findsOneWidget);
    expect(find.text('Close position'), findsNothing);
  });
}
