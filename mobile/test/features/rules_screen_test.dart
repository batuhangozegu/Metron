import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/live_models.dart';
import 'package:metron/router.dart';

import '../live_test_app.dart';

void main() {
  testWidgets('Rules loads saved profile, validates and PUTs eight fields', (
    tester,
  ) async {
    final fake = FakeBackend();
    final container = await bootLive(tester, backend: fake);
    container.read(routerProvider).go('/rules');
    await tester.pumpAndSettle();
    expect(find.text('Defaults, not saved yet'), findsOneWidget);
    await tester.enterText(
      find.byKey(const ValueKey('risk_per_trade_percent')),
      '0',
    );
    await tester.scrollUntilVisible(
      find.text('Save changes'),
      300,
      scrollable: find.byType(Scrollable).last,
    );
    await tester.tap(find.text('Save changes'));
    await tester.pumpAndSettle();
    expect(
      find.text(
        'Enter a value greater than 0 and at most 100, with up to 8 decimal places.',
      ),
      findsOneWidget,
    );
    await tester.scrollUntilVisible(
      find.byKey(const ValueKey('risk_per_trade_percent')),
      -300,
      scrollable: find.byType(Scrollable).last,
    );
    await tester.enterText(
      find.byKey(const ValueKey('risk_per_trade_percent')),
      '0.12345678',
    );
    await tester.scrollUntilVisible(
      find.text('Save changes'),
      300,
      scrollable: find.byType(Scrollable).last,
    );
    await tester.tap(find.text('Save changes'));
    await tester.pumpAndSettle();
    final body = jsonDecode(
      fake.requests.lastWhere((r) => r.method == 'PUT').body,
    ) as Map;
    expect(body.keys.toSet(), profileFields.toSet());
    expect(body['risk_per_trade_percent'], '0.12345678');
    expect(body['max_open_positions'], 5);
  });
  test('Risk validation uses exact decimal and uint32 boundaries', () {
    expect(validPositiveDecimal('100.00000000', '100'), isTrue);
    for (final value in [
      '100.00000001',
      '0',
      '-1',
      '1e2',
      ' 2',
      '0.000000001',
    ]) {
      expect(validPositiveDecimal(value, '100'), isFalse);
    }
    expect(validRiskCount('4294967295'), isTrue);
    expect(validRiskCount('4294967296'), isFalse);
  });
}
