import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/app.dart';
import 'package:metron/features/rules/rules_screen.dart';
import 'package:metron/l10n/locale_provider.dart';
import 'package:metron/router.dart';

void main() {
  test('Device locale defaults to Turkish only for Turkish devices', () {
    expect(resolveDeviceLocale(const Locale('tr', 'TR')), const Locale('tr'));
    expect(resolveDeviceLocale(const Locale('en', 'GB')), const Locale('en'));
    expect(resolveDeviceLocale(const Locale('de')), const Locale('en'));
  });

  testWidgets(
    'Turkish device renders Rules and language choice survives navigation',
    (tester) async {
      tester.binding.platformDispatcher.localeTestValue = const Locale(
        'tr',
        'TR',
      );
      addTearDown(tester.binding.platformDispatcher.clearLocaleTestValue);
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final router = createRouter(initialLocation: '/rules');
      addTearDown(router.dispose);
      await tester.pumpWidget(
        ProviderScope(
          overrides: [routerProvider.overrideWithValue(router)],
          child: const MetronApp(),
        ),
      );
      await tester.runAsync(() async => GoogleFonts.pendingFonts());
      await tester.pumpAndSettle();
      final context = tester.element(find.byType(RulesScreen));
      expect(Localizations.localeOf(context), const Locale('tr'));
      expect(MaterialLocalizations.of(context), isA<MaterialLocalizationTr>());
      expect(find.text('HESAP LİMİTLERİ'), findsOneWidget);
      expect(find.text('%6,00'), findsOneWidget);
      expect(find.text('Piyasalar'), findsOneWidget);

      await tester.tap(find.byKey(const ValueKey('language-selector')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('English').last);
      await tester.pumpAndSettle();
      expect(Localizations.localeOf(context), const Locale('en'));
      expect(find.text('ACCOUNT LIMITS'), findsOneWidget);
      expect(find.text('6.00%'), findsOneWidget);

      await tester.tap(find.text('Markets'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Rules'));
      await tester.pumpAndSettle();
      expect(find.text('ACCOUNT LIMITS'), findsOneWidget);

      await tester.tap(find.byKey(const ValueKey('language-selector')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Türkçe').last);
      await tester.pumpAndSettle();
      expect(find.text('HESAP LİMİTLERİ'), findsOneWidget);
      expect(Localizations.localeOf(context), const Locale('tr'));
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('Turkish home renders a localized account amount', (
    tester,
  ) async {
    tester.binding.platformDispatcher.localeTestValue = const Locale('tr');
    addTearDown(tester.binding.platformDispatcher.clearLocaleTestValue);
    final router = createRouter();
    addTearDown(router.dispose);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [routerProvider.overrideWithValue(router)],
        child: const MetronApp(),
      ),
    );
    await tester.runAsync(() async => GoogleFonts.pendingFonts());
    await tester.pumpAndSettle();
    expect(find.text('25.000,00'), findsOneWidget);
    expect(find.text('%2,75'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
