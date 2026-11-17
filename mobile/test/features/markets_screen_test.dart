import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/app.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/features/check/check_screen.dart';
import 'package:metron/features/markets/markets_strings.dart';
import 'package:metron/router.dart';
import 'package:metron/widgets/num.dart';
import 'package:metron/widgets/pnl_text.dart';

Future<void> boot(
  WidgetTester tester, {
  ThemeMode theme = ThemeMode.dark,
  bool empty = false,
}) async {
  tester.view.physicalSize = const Size(390, 1000);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final router = createRouter(initialLocation: '/markets');
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        routerProvider.overrideWithValue(router),
        if (empty) pricesProvider.overrideWithValue([]),
      ],
      child: MetronApp(themeMode: theme),
    ),
  );
  await tester.runAsync(() => GoogleFonts.pendingFonts());
  await tester.pumpAndSettle();
}

Finder filter(String text) => find.widgetWithText(OutlinedButton, text);

void main() {
  for (final theme in [ThemeMode.dark, ThemeMode.light]) {
    testWidgets('Markets renders mockup data at 390 px in $theme', (
      tester,
    ) async {
      await boot(tester, theme: theme);
      expect(find.text('Live'), findsOneWidget);
      expect(find.text('BTC/USDT'), findsOneWidget);
      expect(find.text('PENDLE/USDT'), findsOneWidget);
      expect(find.text('97,846.00'), findsOneWidget);
      expect(find.text('3,461.04'), findsOneWidget);
      expect(find.text('1.124'), findsOneWidget);
      expect(find.text('ATR 2.0%'), findsOneWidget);
      expect(find.text('6.75%'), findsOneWidget);
      expect(find.byType(PnlText), findsNWidgets(9));
      expect(find.byType(LinearProgressIndicator), findsNWidgets(9));
      final bars = tester.widgetList<LinearProgressIndicator>(
        find.byType(LinearProgressIndicator),
      );
      expect(bars.first.value, .2);
      expect(bars.elementAt(7).value, .8);
      expect(
        tester
            .widget<Num>(
              find.ancestor(
                of: find.text('97,846.00'),
                matching: find.byType(Num),
              ),
            )
            .data,
        '97,846.00',
      );
      for (final button in tester.widgetList<OutlinedButton>(
        find.byType(OutlinedButton),
      )) {
        expect(tester.getSize(find.byWidget(button)).height, 44);
      }
      await tester.drag(
        find.byType(SingleChildScrollView).first,
        const Offset(0, -350),
      );
      await tester.pumpAndSettle();
      expect(find.text('INJ/USDT').hitTestable(), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('Search intersects class filters and recovers from no matches', (
    tester,
  ) async {
    await boot(tester);
    await tester.tap(filter('BTC/ETH'));
    await tester.pumpAndSettle();
    expect(find.text('BTC/USDT'), findsOneWidget);
    expect(find.text('ETH/USDT'), findsOneWidget);
    expect(find.byType(PnlText), findsNWidgets(2));
    await tester.enterText(find.byType(TextField), '  sol  ');
    await tester.pumpAndSettle();
    expect(find.text(const EnglishMarketsStrings().noResults), findsOneWidget);
    await tester.tap(filter('Large-cap'));
    await tester.pumpAndSettle();
    expect(find.text('SOL/USDT'), findsOneWidget);
    expect(find.byType(PnlText), findsOneWidget);
    await tester.enterText(find.byType(TextField), '');
    await tester.ensureVisible(filter('Small-cap'));
    await tester.pumpAndSettle();
    await tester.tap(filter('Small-cap'));
    await tester.pumpAndSettle();
    expect(find.text('PENDLE/USDT'), findsOneWidget);
    expect(find.text('INJ/USDT'), findsOneWidget);
    expect(find.byType(PnlText), findsNWidgets(2));
    await tester.ensureVisible(filter('All'));
    await tester.pumpAndSettle();
    await tester.tap(filter('All'));
    await tester.pumpAndSettle();
    expect(find.byType(PnlText), findsNWidgets(9));
    expect(tester.takeException(), isNull);
  });

  testWidgets('Row opens check with pair and back preserves search', (
    tester,
  ) async {
    await boot(tester);
    await tester.enterText(find.byType(TextField), 'pendle');
    await tester.pumpAndSettle();
    await tester.tap(find.text('PENDLE/USDT'));
    await tester.pumpAndSettle();
    expect(
      tester.widget<CheckScreen>(find.byType(CheckScreen)).pair,
      'PENDLE/USDT',
    );
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.text('PENDLE/USDT'), findsOneWidget);
    expect(find.byType(PnlText), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Empty provider shows useful empty state', (tester) async {
    await boot(tester, empty: true);
    expect(find.text(const EnglishMarketsStrings().noResults), findsOneWidget);
    expect(find.byType(PnlText), findsNothing);
    expect(tester.takeException(), isNull);
  });
}
