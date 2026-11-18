import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/data/models.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/features/alerts/alerts_screen.dart';
import 'package:metron/theme/metron_theme.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/widgets/widgets.dart';

Future<void> boot(
  WidgetTester tester, {
  bool light = false,
  bool empty = false,
  bool missingPrice = false,
  double scale = 1,
}) async {
  tester.view.physicalSize = const Size(390, 940);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  GoogleFonts.config.allowRuntimeFetching = false;
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        if (empty) alertsProvider.overrideWithValue([]),
        if (empty) priceAlarmsProvider.overrideWithValue([]),
        if (missingPrice) pricesProvider.overrideWithValue([]),
      ],
      child: MaterialApp(
        theme: light ? MetronTheme.light : MetronTheme.dark,
        builder: (context, child) => MediaQuery(
          data: MediaQuery.of(context)
              .copyWith(textScaler: TextScaler.linear(scale)),
          child: child!,
        ),
        home: const AlertsScreen(),
      ),
    ),
  );
  await tester.runAsync(() async {
    await GoogleFonts.pendingFonts();
  });
  await tester.pumpAndSettle();
}

void main() {
  for (final light in [false, true]) {
    testWidgets(
      'Mockup content, risk and PnL at 390px (${light ? 'light' : 'dark'})',
      (tester) async {
        await boot(tester, light: light);
        expect(find.text('ALERTS'), findsOneWidget);
        expect(find.text('PAPER'), findsOneWidget);
        expect(find.text('TODAY'), findsOneWidget);
        expect(
          find.text('Cooldown is counting down', findRichText: true),
          findsOneWidget,
        );
        expect(find.text('14:20'), findsOneWidget);
        expect(find.text('14:42'), findsOneWidget);
        expect(find.text('31.93'), findsOneWidget);
        expect(find.text('97,500.00'), findsOneWidget);
        expect(find.text('20.58'), findsOneWidget);
        expect(find.byType(PnlText), findsNWidgets(2));
        expect(find.text('▼ −187.50'), findsOneWidget);
        expect(find.text('▼ −125.00'), findsOneWidget);
        expect(find.text('Caution'), findsOneWidget);
        expect(find.text('Limit exceeded'), findsOneWidget);
        expect(find.text('Order blocked', findRichText: true), findsOneWidget);
        await tester.ensureVisible(find.text('ETH/USDT falls below'));
        expect(find.text('BTC/USDT rises above'), findsOneWidget);
        for (final price in [
          '100,000.00',
          '3,300.00',
          '97,846.00',
          '3,461.04',
        ]) {
          expect(find.widgetWithText(Num, price), findsOneWidget);
        }
        final context = tester.element(find.byType(AlertsScreen));
        final badges = tester
            .widgetList<RiskBadge>(find.byType(RiskBadge))
            .toList();
        expect(
          badges.first.level.color(context.colors),
          context.colors.caution,
        );
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets(
    'Empty providers and alarm form validation, creation and cancellation',
    (tester) async {
      await boot(tester, empty: true);
      expect(find.text('No events today.'), findsOneWidget);
      expect(find.text('No price alarms yet.'), findsOneWidget);
      await tester.tap(find.text('Add an alarm'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(PrimaryButton, 'Add an alarm'));
      await tester.pumpAndSettle();
      expect(find.text('Enter a positive price.'), findsOneWidget);
      await tester.enterText(find.byType(TextFormField), '-1');
      await tester.tap(find.widgetWithText(PrimaryButton, 'Add an alarm'));
      await tester.pumpAndSettle();
      expect(find.text('Enter a positive price.'), findsOneWidget);
      await tester.enterText(find.byType(TextFormField), '105000');
      await tester.tap(find.widgetWithText(PrimaryButton, 'Add an alarm'));
      await tester.pumpAndSettle();
      expect(find.text('105,000.00'), findsOneWidget);
      expect(find.text('No price alarms yet.'), findsNothing);
      await tester.tap(find.text('Add an alarm'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(find.text('105,000.00'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('Missing current prices and large text remain usable', (
    tester,
  ) async {
    await boot(tester, missingPrice: true, scale: 1.5);
    await tester.ensureVisible(find.text('ETH/USDT falls below'));
    expect(find.text('—'), findsNWidgets(2));
    await tester.ensureVisible(find.text('Add an alarm'));
    await tester.tap(find.text('Add an alarm'));
    await tester.pumpAndSettle();
    expect(find.text('No pairs available.'), findsOneWidget);
    expect(
      tester
          .widget<PrimaryButton>(
            find.widgetWithText(PrimaryButton, 'Add an alarm'),
          )
          .onPressed,
      isNull,
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets('Today excludes events from other sample dates', (tester) async {
    GoogleFonts.config.allowRuntimeFetching = false;
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          alertsProvider.overrideWithValue([
            AlertEvent(
              id: 'old',
              kind: AlertKind.priceAlarm,
              title: 'Yesterday alarm',
              body: 'Old event',
              at: DateTime.utc(2026, 11, 13, 12),
            ),
          ]),
        ],
        child: MaterialApp(theme: MetronTheme.dark, home: const AlertsScreen()),
      ),
    );
    await tester.runAsync(() async {
      await GoogleFonts.pendingFonts();
    });
    await tester.pumpAndSettle();
    expect(find.text('No events today.'), findsOneWidget);
    expect(find.text('Yesterday alarm', findRichText: true), findsNothing);
  });
}
