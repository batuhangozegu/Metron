import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/providers.dart';

import '../live_test_app.dart';

void main() {
  testWidgets('Home renders exact live decimals, refreshes and retries', (
    tester,
  ) async {
    final fake = FakeBackend();
    await bootLive(tester, backend: fake);
    expect(find.text('10,000.12345678 USDT'), findsOneWidget);
    expect(find.text('9,000.12 USDT'), findsOneWidget);
    final before = fake.requests
        .where((r) => r.url.path.endsWith('/portfolio'))
        .length;
    await tester.pump(const Duration(seconds: 5));
    await tester.pumpAndSettle();
    expect(
      fake.requests.where((r) => r.url.path.endsWith('/portfolio')).length,
      greaterThan(before),
    );
    fake.portfolioFail = true;
    await tester.pump(const Duration(seconds: 5));
    await tester.pumpAndSettle();
    expect(find.text('Retry'), findsOneWidget);
    fake.portfolioFail = false;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();
    expect(find.text('10,000.12345678 USDT'), findsOneWidget);
  });
  testWidgets('No accounts offers creation with string balance and market', (
    tester,
  ) async {
    final fake = FakeBackend();
    fake.accounts.clear();
    await bootLive(tester, backend: fake);
    expect(
      find.text('Create a USDT paper account to start trading.'),
      findsOneWidget,
    );
    await tester.enterText(
      find.byKey(const ValueKey('account-name')),
      'Phone practice',
    );
    await tester.enterText(
      find.byKey(const ValueKey('account-balance')),
      '1234.12345678',
    );
    await tester.tap(find.text('Create paper account'));
    await tester.pumpAndSettle();
    final body = jsonDecode(
      fake.requests
          .lastWhere((r) => r.url.path == '/v1/accounts' && r.method == 'POST')
          .body,
    );
    expect(body, {
      'name': 'Phone practice',
      'paper_balance': '1234.12345678',
      'market': 'spot',
    });
    expect(find.text('10,000.12345678 USDT'), findsOneWidget);
  });
  testWidgets('Picking an account persists the choice per user and API', (
    tester,
  ) async {
    final fake = FakeBackend();
    fake.accounts.add({...accountJson, 'id': 'account-2', 'name': 'Second'});
    final container = await bootLive(tester, backend: fake);
    await tester.tap(find.byType(DropdownButtonFormField<String>));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Second · spot').last);
    await tester.pumpAndSettle();
    expect(container.read(selectedAccountIdProvider), 'account-2');
    container.invalidate(accountsProvider);
    await tester.pumpAndSettle();
    expect(container.read(selectedAccountIdProvider), 'account-2');
    expect(fake.store.values.values, contains('account-2'));
  });
}
