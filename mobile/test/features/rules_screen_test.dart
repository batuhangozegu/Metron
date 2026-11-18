import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/app.dart';
import 'package:metron/data/models.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/data/sample_data.dart';
import 'package:metron/features/rules/rules_screen.dart';
import 'package:metron/features/sign_in/sign_in_screen.dart';
import 'package:metron/router.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/widgets/num.dart';

Future<void> boot(
  WidgetTester tester, {
  bool light = false,
  bool empty = false,
}) async {
  tester.view.physicalSize = const Size(390, 844);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final router = createRouter(initialLocation: '/rules');
  addTearDown(router.dispose);
  final sample = SampleData.account;
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        routerProvider.overrideWithValue(router),
        if (empty) ...[
          coinClassRulesProvider.overrideWithValue([]),
          disciplineRulesProvider.overrideWithValue([]),
          limitsProvider.overrideWithValue(
            const AccountLimits(
              openRiskPercent: 7.25,
              dailyLossPercent: 3,
              weeklyLossPercent: 8,
              maxOpenPositions: 2,
              targetRiskReward: 3,
              smallCapTotalPercent: 10,
            ),
          ),
          accountProvider.overrideWithValue(
            PaperAccount(
              equity: '123456.78',
              openRiskPercent: sample.openRiskPercent,
              dailyLossPercent: sample.dailyLossPercent,
              weeklyLossPercent: sample.weeklyLossPercent,
              moneyAtRisk: sample.moneyAtRisk,
              dailyRemaining: sample.dailyRemaining,
              weeklyRemaining: sample.weeklyRemaining,
              unrealizedPnl: sample.unrealizedPnl,
              asOf: sample.asOf,
              cooldownEndsAt: sample.cooldownEndsAt,
              cooldownWaitedMinutes: sample.cooldownWaitedMinutes,
              twoFactorEnabled: false,
            ),
          ),
        ],
      ],
      child: MetronApp(themeMode: light ? ThemeMode.light : ThemeMode.dark),
    ),
  );
  await tester.runAsync(() async {
    await GoogleFonts.pendingFonts();
  });
  await tester.pumpAndSettle();
}

Future<void> reveal(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  expect(tester.takeException(), isNull);
}

void main() {
  for (final light in [false, true]) {
    testWidgets(
      'Rules match mockup and sign out at 390px (${light ? 'light' : 'dark'})',
      (tester) async {
        final semantics = tester.ensureSemantics();
        await boot(tester, light: light);
        expect(
          find.text(
            'Shown here, changed on the web. Rules are set at a desk, not in the moment.',
          ),
          findsOneWidget,
        );
        for (final value in ['2.00%', '6.00%', '5.00%', '10.00%', '2.0']) {
          expect(find.text(value), findsOneWidget);
        }
        expect(find.text('5'), findsNWidgets(2));
        expect(find.text('3'), findsNWidgets(2));
        expect(find.text('Spot portfolio allocation cap'), findsNWidgets(3));
        for (final label in ['BTC/ETH', 'Large-cap', 'Small-cap']) {
          await reveal(tester, find.text(label));
        }
        expect(find.text('1.5 ATR'), findsNWidgets(3));
        expect(find.text('0.75%'), findsOneWidget);
        expect(find.text('8%'), findsOneWidget);
        await reveal(tester, find.text('Warn and record').first);
        expect(find.byIcon(Icons.warning_amber_rounded), findsNWidgets(2));
        final colors = light ? MetronColors.light : MetronColors.dark;
        expect(
          tester
              .widget<Icon>(find.byIcon(Icons.warning_amber_rounded).first)
              .color,
          colors.caution,
        );
        await reveal(tester, find.text('Block'));
        expect(
          tester.widget<Icon>(find.byIcon(Icons.report_outlined)).color,
          colors.limit,
        );
        for (final copy in [
          'No orders after 00:00',
          'No new order for 60 minutes after a position closes at a loss.',
          'Blocked 3 times',
        ]) {
          final content = find.byWidgetPredicate(
            (widget) => widget is Semantics && widget.properties.label == copy,
          );
          await reveal(tester, content);
          expect(tester.getSemantics(content).label, contains(copy));
        }
        expect(
          find.byWidgetPredicate(
            (widget) => widget is Num && widget.data == '60',
          ),
          findsOneWidget,
        );
        await reveal(tester, find.text('Sign out'));
        expect(find.text('25,000.00 USDT'), findsOneWidget);
        expect(find.text('On'), findsOneWidget);
        expect(find.byType(TextField), findsNothing);
        final button = find.widgetWithText(TextButton, 'Sign out');
        expect(tester.getSize(button).height, 44);
        await tester.tap(button);
        await tester.pumpAndSettle();
        expect(find.byType(SignInScreen), findsOneWidget);
        expect(find.byType(RulesScreen), findsNothing);
        expect(find.byType(BottomNavigationBar), findsNothing);
        expect(tester.takeException(), isNull);
        semantics.dispose();
      },
    );
  }

  testWidgets('Rules use overridden limits, empty lists and two-factor off', (
    tester,
  ) async {
    await boot(tester, empty: true, light: true);
    expect(find.text('7.25%'), findsOneWidget);
    expect(find.text('3.0'), findsOneWidget);
    expect(find.text('No coin class rules set.'), findsOneWidget);
    expect(find.text('No discipline rules set.'), findsOneWidget);
    await reveal(tester, find.text('Sign out'));
    expect(find.text('123,456.78 USDT'), findsOneWidget);
    expect(find.text('Off'), findsOneWidget);
    expect(find.byIcon(Icons.remove_circle_outline), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
