import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/app.dart';
import 'package:metron/data/models.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/data/risk_level.dart';
import 'package:metron/data/sample_data.dart';
import 'package:metron/features/check/check_screen.dart';
import 'package:metron/features/position/position_screen.dart';
import 'package:metron/router.dart';
import 'package:metron/widgets/measure.dart';
import 'package:metron/widgets/risk_badge.dart';

Future<void> boot(
  WidgetTester tester, {
  bool light = false,
  PaperAccount? account,
  List<Position>? positions,
}) async {
  tester.view.physicalSize = const Size(390, 1060);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final router = createRouter();
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        routerProvider.overrideWithValue(router),
        if (account != null) accountProvider.overrideWithValue(account),
        if (positions != null) positionsProvider.overrideWithValue(positions),
      ],
      child: MetronApp(themeMode: light ? ThemeMode.light : ThemeMode.dark),
    ),
  );
  await tester.runAsync(() async {
    await GoogleFonts.pendingFonts();
  });
  await tester.pumpAndSettle();
}

PaperAccount stateAccount({
  double openRisk = 2.75,
  double dailyLoss = 1.25,
  bool cooldown = true,
}) {
  final a = SampleData.account;
  return PaperAccount(
    equity: a.equity,
    openRiskPercent: openRisk,
    dailyLossPercent: dailyLoss,
    weeklyLossPercent: a.weeklyLossPercent,
    moneyAtRisk: a.moneyAtRisk,
    dailyRemaining: a.dailyRemaining,
    weeklyRemaining: a.weeklyRemaining,
    unrealizedPnl: a.unrealizedPnl,
    asOf: a.asOf,
    cooldownEndsAt: cooldown ? a.cooldownEndsAt : a.asOf,
    cooldownWaitedMinutes: a.cooldownWaitedMinutes,
    twoFactorEnabled: a.twoFactorEnabled,
  );
}

void main() {
  for (final light in [false, true]) {
    testWidgets('Home matches fixture content at 390 px, light=$light', (
      tester,
    ) async {
      await boot(tester, light: light);
      for (final text in [
        'Metron',
        'PAPER',
        '25,000.00',
        '14:20 UTC',
        'Within your limits',
        'Total open risk',
        '2.75%',
        '6.00%',
        '687.50 USDT',
        'Daily loss cap',
        '1.25%',
        '937.50',
        'Weekly loss cap',
        '1.80%',
        '2,050.00',
        'Cooldown after a loss',
        '14:42',
        '22',
        'Open positions',
        'BTC/USDT',
        'SOL/USDT',
        'ETH/USDT',
      ]) {
        expect(find.text(text), findsOneWidget, reason: text);
      }
      expect(find.text('▲ +140.85'), findsOneWidget);
      expect(find.text('▲ +125.01'), findsOneWidget);
      expect(find.text('▼ −55.59'), findsOneWidget);
      expect(find.text('▲ +71.43'), findsOneWidget);
      final painters = tester
          .widgetList<CustomPaint>(find.byType(CustomPaint))
          .map((w) => w.painter)
          .whereType<MeasurePainter>()
          .toList();
      expect(painters.length, 3);
      expect(painters.first.fraction, closeTo(2.75 / 6, .00001));
      expect(
        find.byWidgetPredicate(
          (w) => w is RiskBadge && w.level == RiskLevel.safe,
        ),
        findsNWidgets(3),
      );
      expect(tester.takeException(), isNull);
      final actionY = tester.getTopLeft(find.text('Check a new order')).dy;
      await tester.drag(
        find.byType(SingleChildScrollView),
        const Offset(0, -400),
      );
      await tester.pumpAndSettle();
      expect(tester.getTopLeft(find.text('Check a new order')).dy, actionY);
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('Risk headline responds to caution and exceeded daily cap', (
    tester,
  ) async {
    await boot(tester, account: stateAccount(openRisk: 4.5));
    expect(find.text('Approaching your limits'), findsOneWidget);
    expect(
      find.byWidgetPredicate(
        (w) => w is RiskBadge && w.level == RiskLevel.caution,
      ),
      findsNWidgets(2),
    );
    // Replace the scope so the second state uses a fresh provider container.
    await tester.pumpWidget(const SizedBox());
    await boot(
      tester,
      account: stateAccount(dailyLoss: 5.01, cooldown: false),
      positions: [],
    );
    expect(find.text('Limit exceeded'), findsNWidgets(2));
    expect(find.text('Cooldown after a loss'), findsNothing);
    expect(find.text('No open positions'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Position tap passes its ID and fixed action opens check', (
    tester,
  ) async {
    await boot(tester);
    await tester.ensureVisible(find.text('SOL/USDT'));
    await tester.tap(find.text('SOL/USDT'));
    await tester.pumpAndSettle();
    expect(
      tester.widget<PositionScreen>(find.byType(PositionScreen)).positionId,
      'sol',
    );
    await tester.pageBack();
    await tester.pumpAndSettle();
    await tester.tap(find.text('Check a new order'));
    await tester.pumpAndSettle();
    expect(find.byType(CheckScreen), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Cooldown opens the Alerts tab', (tester) async {
    await boot(tester);
    await tester.ensureVisible(find.text('Cooldown after a loss'));
    await tester.tap(find.text('Cooldown after a loss'));
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<BottomNavigationBar>(find.byType(BottomNavigationBar))
          .currentIndex,
      2,
    );
    expect(tester.takeException(), isNull);
  });
}
