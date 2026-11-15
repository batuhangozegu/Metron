import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/app.dart';
import 'package:metron/features/alerts/alerts_screen.dart';
import 'package:metron/features/check/check_screen.dart';
import 'package:metron/features/check_blocked/check_blocked_screen.dart';
import 'package:metron/features/home/home_screen.dart';
import 'package:metron/features/markets/markets_screen.dart';
import 'package:metron/features/position/position_screen.dart';
import 'package:metron/features/rules/rules_screen.dart';
import 'package:metron/features/sign_in/sign_in_screen.dart';
import 'package:metron/router.dart';
import 'package:metron/widgets/fixed_main_action.dart';

Future<void> boot(
  WidgetTester tester, {
  String location = '/',
  ThemeMode theme = ThemeMode.dark,
}) async {
  final router = createRouter(initialLocation: location);
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [routerProvider.overrideWithValue(router)],
      child: MetronApp(themeMode: theme),
    ),
  );
  await tester.runAsync(() async {
    await GoogleFonts.pendingFonts();
  });
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('App boots dark and all four bottom tabs switch', (tester) async {
    await boot(tester);
    expect(find.byType(HomeScreen), findsOneWidget);
    expect(
      Theme.of(tester.element(find.byType(HomeScreen))).brightness,
      Brightness.dark,
    );
    expect(find.byType(FixedMainAction), findsOneWidget);
    final nav = find.byType(BottomNavigationBar);
    for (final (label, type, index) in [
      ('Markets', MarketsScreen, 1),
      ('Alerts', AlertsScreen, 2),
      ('Rules', RulesScreen, 3),
      ('Risk', HomeScreen, 0),
    ]) {
      await tester.tap(find.descendant(of: nav, matching: find.text(label)));
      await tester.pumpAndSettle();
      expect(find.byType(type), findsOneWidget);
      expect(tester.widget<BottomNavigationBar>(nav).currentIndex, index);
    }
    expect(tester.takeException(), isNull);
  });

  testWidgets('Fixed main action opens check; back returns to tabs', (
    tester,
  ) async {
    await boot(tester);
    await tester.tap(find.text('Check a new order'));
    await tester.pumpAndSettle();
    expect(find.byType(CheckScreen), findsOneWidget);
    expect(find.byType(BottomNavigationBar), findsNothing);
    await tester.pageBack();
    await tester.pumpAndSettle();
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  for (final (path, screen) in [
    ('/check/blocked', CheckBlockedScreen),
    ('/positions/btc', PositionScreen),
    ('/sign-in', SignInScreen),
    ('/check?pair=SOL%2FUSDT', CheckScreen),
  ]) {
    testWidgets('Deep link $path resolves independently', (tester) async {
      await boot(tester, location: path, theme: ThemeMode.light);
      expect(find.byType(screen), findsOneWidget);
      if (screen == PositionScreen) {
        expect(
          tester.widget<PositionScreen>(find.byType(screen)).positionId,
          'btc',
        );
      }
      if (screen == CheckScreen) {
        expect(
          tester.widget<CheckScreen>(find.byType(screen)).pair,
          'SOL/USDT',
        );
      }
      expect(
        Theme.of(tester.element(find.byType(screen))).brightness,
        Brightness.light,
      );
      expect(tester.takeException(), isNull);
    });
  }
}
