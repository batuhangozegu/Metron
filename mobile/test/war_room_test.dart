import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/app.dart';
import 'package:metron/l10n/locale_provider.dart';
import 'package:metron/l10n/app_strings.dart';
import 'package:metron/router.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/theme/tokens.dart';
import 'package:metron/widgets/segmented_bar.dart';
import 'package:metron/widgets/widgets.dart';

Future<void> bootRoom(
  WidgetTester tester, {
  String route = '/',
  String language = 'en',
  bool light = false,
}) async {
  tester.view.physicalSize = const Size(320, 900);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final router = createRouter(initialLocation: route);
  final container = ProviderContainer(
    overrides: [routerProvider.overrideWithValue(router)],
  );
  container.read(localeProvider.notifier).select(Locale(language));
  addTearDown(router.dispose);
  addTearDown(container.dispose);
  await tester.pumpWidget(
    UncontrolledProviderScope(
      container: container,
      child: MetronApp(themeMode: light ? ThemeMode.light : ThemeMode.dark),
    ),
  );
  await tester.runAsync(() => GoogleFonts.pendingFonts());
  await tester.pumpAndSettle();
}

void main() {
  for (final route in [
    '/',
    '/markets',
    '/alerts',
    '/rules',
    '/check',
    '/check/blocked',
    '/positions/btc',
  ]) {
    testWidgets('Ticker sits above the title on $route', (tester) async {
      await bootRoom(tester, route: route);
      expect(find.byType(MarketTicker), findsOneWidget);
      expect(
        tester.getBottomLeft(find.byType(MarketTicker)).dy,
        lessThanOrEqualTo(
          tester.getTopLeft(find.byType(CommandTitle).first).dy,
        ),
      );
      expect(tester.takeException(), isNull);
    });
  }

  for (final language in ['en', 'tr']) {
    testWidgets('Check metadata wraps without dangling separators: $language', (
      tester,
    ) async {
      await bootRoom(tester, route: '/check', language: language);
      final strings = language == 'tr'
          ? const TurkishAppStrings()
          : const EnglishAppStrings();
      expect(
        find.text('${strings.accountMarket}: ${strings.spot}'),
        findsOneWidget,
      );
      final meta = find.byKey(const ValueKey('check-meta'));
      final texts = tester.widgetList<Text>(
        find.descendant(of: meta, matching: find.byType(Text)),
      );
      expect(texts, hasLength(4));
      expect(texts.every((text) => !text.data!.contains('·')), isTrue);
      expect(
        find.text(language == 'tr' ? 'ATR / gün %8,0' : 'ATR / day 8.0%'),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
    });
  }
  testWidgets('Oswald and its license are bundled for offline use', (
    tester,
  ) async {
    await bootRoom(tester);
    final font = await rootBundle.load('assets/fonts/Oswald-Variable.ttf');
    expect(font.lengthInBytes, greaterThan(100000));
    final license = await rootBundle.loadString('assets/fonts/Oswald-OFL.txt');
    expect(license, contains('SIL OPEN FONT LICENSE'));
    expect(GoogleFonts.config.allowRuntimeFetching, isFalse);
    final title = tester.widget<Text>(find.text('WITHIN YOUR LIMITS'));
    expect(title.style!.fontFamily, MetronType.condensed);
    final gauge = tester.widget<Text>(find.text('2.75%'));
    expect(
      gauge.style!.fontFeatures,
      contains(const FontFeature.tabularFigures()),
    );
  });

  testWidgets('Ticker scrolls, pauses immediately and resumes', (tester) async {
    await bootRoom(tester);
    expect(find.text('SAMPLE'), findsOneWidget);
    final list = tester.widget<ListView>(
      find.byKey(const ValueKey('ticker-quotes')),
    );
    final scroll = list.controller!;
    await tester.pump(const Duration(seconds: 4));
    await tester.pump(const Duration(seconds: 1));
    expect(scroll.offset, greaterThan(0));
    await tester.tap(find.byTooltip('Pause ticker'));
    await tester.pumpAndSettle();
    final pausedOffset = scroll.offset;
    await tester.pump(const Duration(seconds: 8));
    expect(scroll.offset, pausedOffset);
    await tester.tap(find.byTooltip('Resume ticker'));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(seconds: 4));
    await tester.pump(const Duration(seconds: 1));
    expect(scroll.offset, greaterThan(pausedOffset));
    expect(tester.takeException(), isNull);
  });

  testWidgets('Reduced motion keeps quotes readable and manually scrollable', (
    tester,
  ) async {
    tester.platformDispatcher.accessibilityFeaturesTestValue =
        const FakeAccessibilityFeatures(disableAnimations: true);
    addTearDown(tester.platformDispatcher.clearAccessibilityFeaturesTestValue);
    await bootRoom(tester, language: 'tr');
    expect(find.text('ÖRNEK'), findsOneWidget);
    final finder = find.byKey(const ValueKey('ticker-quotes'));
    final controller = tester.widget<ListView>(finder).controller!;
    await tester.pump(const Duration(seconds: 12));
    expect(controller.offset, 0);
    await tester.drag(finder, const Offset(-180, 0));
    await tester.pumpAndSettle();
    expect(controller.offset, greaterThan(0));
    expect(find.byTooltip('Piyasa bandını duraklat'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  for (final light in [false, true]) {
    for (final language in ['en', 'tr']) {
      testWidgets(
        'Blocked stays actionable at 320px: $language, light=$light',
        (tester) async {
          await bootRoom(
            tester,
            route: '/check/blocked',
            language: language,
            light: light,
          );
          final strings = language == 'tr'
              ? const TurkishAppStrings()
              : const EnglishAppStrings();
          expect(
            find.text('${strings.accountMarket}: ${strings.futures}'),
            findsOneWidget,
          );
          expect(
            find.text(
              language == 'tr'
                  ? 'ENGELLENDİ: GÜNLÜK KAYIP SINIRI DOLDU'
                  : 'BLOCKED: DAILY LOSS LIMIT REACHED',
            ),
            findsOneWidget,
          );
          expect(
            find.byWidgetPredicate(
              (widget) =>
                  widget is Semantics &&
                  widget.properties.label == strings.dailyLossBlockedBody,
            ),
            findsOneWidget,
          );
          expect(
            tester.widget<PrimaryButton>(find.byType(PrimaryButton)).onPressed,
            isNull,
          );
          expect(find.text(strings.waitUntilReset), findsOneWidget);
          final stamp = find.byKey(const ValueKey('blocked-stamp'));
          expect(stamp, findsOneWidget);
          expect(
            find.text(language == 'tr' ? 'ENGELLENDİ' : 'BLOCKED'),
            findsOneWidget,
          );
          expect(find.byType(PrimaryButton), findsOneWidget);
          expect(find.byType(SecondaryButton), findsNothing);
          final c = light ? MetronColors.light : MetronColors.dark;
          expect(
            tester
                .widget<Text>(
                  find.text(language == 'tr' ? 'ENGELLENDİ' : 'BLOCKED'),
                )
                .style!
                .color,
            c.limit,
          );
          expect(tester.takeException(), isNull);
        },
      );
    }
  }

  testWidgets('Segments keep gaps and heat amber to red at the limit', (
    tester,
  ) async {
    final pixels = await tester.runAsync(() async {
      final recorder = ui.PictureRecorder();
      final canvas = Canvas(recorder);
      paintSegmentedBar(
        canvas,
        const Rect.fromLTWH(0, 0, 140, 14),
        fraction: 1,
        color: MetronColors.dark.limit,
        startColor: MetronColors.dark.caution,
        track: MetronColors.dark.raised,
      );
      final picture = recorder.endRecording();
      final image = await picture.toImage(140, 14);
      final bytes = await image.toByteData(format: ui.ImageByteFormat.rawRgba);
      image.dispose();
      picture.dispose();
      return bytes!;
    });
    int channel(int x, int channel) =>
        pixels!.getUint8((7 * 140 + x) * 4 + channel);
    expect(channel(12, 3), 0, reason: 'Each segment has a transparent gap');
    expect(channel(5, 3), 255);
    expect(
      channel(5, 1),
      greaterThan(channel(130, 1)),
      reason: 'The amber start heats toward the red end',
    );
  });
}
