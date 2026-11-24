import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/l10n/locale_provider.dart';
import 'package:metron/router.dart';

import 'live_test_app.dart';

void main() {
  test('Device locale selects Turkish only for Turkish devices', () {
    expect(resolveDeviceLocale(const Locale('tr', 'TR')), const Locale('tr'));
    expect(resolveDeviceLocale(const Locale('de')), const Locale('en'));
  });
  testWidgets('Turkish live amounts and alerts use localized copy', (
    tester,
  ) async {
    final container = await bootLive(tester, language: 'tr');
    expect(find.text('10.000,12345678 USDT'), findsOneWidget);
    container.read(routerProvider).go('/alerts');
    await tester.pumpAndSettle();
    expect(find.text('Henüz fiyat alarmı yok.'), findsOneWidget);
    container.read(localeProvider.notifier).select(const Locale('en'));
    await tester.pumpAndSettle();
    expect(find.text('No price alerts yet.'), findsOneWidget);
  });
}
