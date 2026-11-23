import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/features/sign_in/sign_in_screen.dart';
import 'package:metron/router.dart';

import 'live_test_app.dart';

void main() {
  testWidgets('Session bootstrap protects deep links and expired sessions', (
    tester,
  ) async {
    final fake = FakeBackend()..signedIn = false;
    final container = await bootLive(tester, backend: fake);
    expect(fake.requests.first.url.path, '/v1/me');
    container.read(routerProvider).go('/check');
    await tester.pumpAndSettle();
    expect(find.byType(SignInScreen), findsOneWidget);
  });
  testWidgets('All four shell tabs remain navigable', (tester) async {
    final container = await bootLive(tester);
    for (final path in ['/markets', '/alerts', '/rules', '/']) {
      container.read(routerProvider).go(path);
      await tester.pumpAndSettle();
      expect(find.byType(BottomNavigationBar), findsOneWidget);
      expect(
        container.read(routerProvider).routeInformationProvider.value.uri.path,
        path,
      );
      expect(tester.takeException(), isNull);
    }
  });
  testWidgets('401 clears authenticated state and redirects to sign-in', (
    tester,
  ) async {
    final container = await bootLive(tester);
    final api = container.read(apiClientProvider);
    api.onUnauthorized?.call();
    await tester.pumpAndSettle();
    expect(container.read(sessionProvider).value, isNull);
    expect(find.byType(SignInScreen), findsOneWidget);
  });
}
