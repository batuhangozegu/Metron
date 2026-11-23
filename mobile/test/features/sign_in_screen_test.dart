import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/features/sign_in/sign_in_screen.dart';
import 'package:metron/features/home/home_screen.dart';

import '../live_test_app.dart';

void main() {
  testWidgets(
    'Sign-in sends credentials, pending cookie and current 2FA code',
    (tester) async {
      final fake = FakeBackend()
        ..signedIn = false
        ..twoFactor = true;
      await bootLive(tester, backend: fake);
      expect(find.byType(SignInScreen), findsOneWidget);
      await tester.enterText(
        find.byKey(const ValueKey('email')),
        'me@example.com',
      );
      await tester.enterText(
        find.byKey(const ValueKey('password')),
        'a long password',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Sign in'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('code')), findsOneWidget);
      expect(
        fake.store.values.values.any((v) => v.contains('pending')),
        isFalse,
      );
      fake.codeFails = true;
      await tester.enterText(find.byKey(const ValueKey('code')), '123456');
      await tester.tap(find.widgetWithText(FilledButton, 'Verify'));
      await tester.pumpAndSettle();
      expect(find.textContaining('invalid, expired'), findsOneWidget);
      fake.codeFails = false;
      await tester.enterText(find.byKey(const ValueKey('code')), '654321');
      await tester.tap(find.widgetWithText(FilledButton, 'Verify'));
      await tester.pumpAndSettle();
      final request = fake.requests.lastWhere(
        (r) => r.url.path.endsWith('/login/2fa'),
      );
      expect(request.headers['cookie'], contains('metron_pending=pending'));
      expect(request.body, contains('654321'));
      expect(find.byType(HomeScreen), findsOneWidget);
      expect(fake.store.values[fake.api.sessionKey], contains('full'));
    },
  );
  testWidgets(
    'Registration and shell logout use the API and clear the session',
    (tester) async {
      final fake = FakeBackend()..signedIn = false;
      final container = await bootLive(tester, backend: fake);
      await tester.tap(find.widgetWithText(TextButton, 'Create an account'));
      await tester.pumpAndSettle();
      await tester.enterText(
        find.byKey(const ValueKey('email')),
        'new@example.com',
      );
      await tester.enterText(
        find.byKey(const ValueKey('password')),
        'a long password',
      );
      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();
      expect(
        fake.requests.any((r) => r.url.path.endsWith('/register')),
        isTrue,
      );
      expect(find.byType(HomeScreen), findsOneWidget);
      await tester.tap(find.text('Sign out'));
      await tester.pumpAndSettle();
      expect(find.byType(SignInScreen), findsOneWidget);
      expect(container.read(sessionProvider).value, isNull);
      expect(fake.store.values[fake.api.sessionKey], isNull);
    },
  );
}
