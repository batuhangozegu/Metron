import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/app.dart';
import 'package:metron/data/models.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/data/risk_level.dart';
import 'package:metron/data/sample_data.dart';
import 'package:metron/features/check/order_measure.dart';
import 'package:metron/features/check/projected_limit.dart';
import 'package:metron/router.dart';
import 'package:metron/widgets/widgets.dart';
import 'package:google_fonts/google_fonts.dart';

Future<void> boot(
  WidgetTester tester, {
  bool blocked = false,
  bool light = false,
  OrderPreview? preview,
  String? location,
}) async {
  tester.view.physicalSize = const Size(390, 844);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  final router = createRouter(
    initialLocation: location ?? (blocked ? '/check/blocked' : '/check'),
  );
  addTearDown(router.dispose);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        routerProvider.overrideWithValue(router),
        if (preview != null) orderPreviewProvider.overrideWithValue(preview),
      ],
      child: MetronApp(themeMode: light ? ThemeMode.light : ThemeMode.dark),
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
      'Caution content, fixed actions and projections; light=$light',
      (tester) async {
        await boot(tester, light: light);
        expect(find.text('One rule would be broken'), findsOneWidget);
        expect(find.text('5.200'), findsNWidgets(2));
        expect(find.text('4.576'), findsOneWidget);
        expect(find.text('6.448'), findsOneWidget);
        expect(find.byType(PnlText), findsNWidgets(4));
        expect(find.text('Place paper order anyway'), findsOneWidget);
        expect(find.text('14:42'), findsOneWidget);
        final actionY = tester
            .getTopLeft(find.text('Place paper order anyway'))
            .dy;
        await tester.drag(
          find.byType(SingleChildScrollView),
          const Offset(0, -950),
        );
        await tester.pumpAndSettle();
        expect(find.byType(ProjectedLimit), findsNWidgets(6));
        expect(find.text('38 of 60 min waited'), findsOneWidget);
        expect(find.text('3 → 4 of 5'), findsOneWidget);
        expect(
          tester.getTopLeft(find.text('Place paper order anyway')).dy,
          actionY,
        );
        expect(tester.takeException(), isNull);
        await tester.tap(find.text('Place paper order anyway'));
        await tester.pump();
        expect(
          find.text('Sample preview only. No order was placed.'),
          findsOneWidget,
        );
      },
    );

    testWidgets('Blocked has no placement path; light=$light', (tester) async {
      await boot(tester, blocked: true, light: light);
      expect(find.text('Blocked: position too large'), findsOneWidget);
      expect(find.text('94,910.62'), findsOneWidget);
      expect(find.text('103,716.76'), findsOneWidget);
      expect(find.text('Place paper order anyway'), findsNothing);
      expect(find.text('Place paper order'), findsNothing);
      expect(find.text('1,666.67'), findsOneWidget);
      expect(find.byType(OrderMeasure), findsOneWidget);
      await tester.drag(
        find.byType(SingleChildScrollView),
        const Offset(0, -950),
      );
      await tester.pumpAndSettle();
      expect(find.byType(ProjectedLimit), findsNWidgets(5));
      expect(find.text('Exceeded'), findsOneWidget);
      expect(find.text('33.33% → 66.67% of 40.00%'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.tap(find.text('Reduce to fit the limit'));
      await tester.pumpAndSettle();
      expect(find.text('Suggested reduction'), findsOneWidget);
      expect(find.text('Place paper order'), findsNothing);
      await tester.tap(find.text('Close'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Discard'));
      await tester.pumpAndSettle();
      expect(find.text('Check a new order'), findsOneWidget);
    });
  }

  testWidgets('All clear uses the same layout with safe banner and actions', (
    tester,
  ) async {
    const source = SampleData.orderPreview;
    await boot(
      tester,
      preview: OrderPreview(
        pair: source.pair,
        verdict: RiskLevel.safe,
        entry: source.entry,
        stop: source.stop,
        target: source.target,
        quantity: source.quantity,
        orderValue: source.orderValue,
        lossAtStop: source.lossAtStop,
        profitAtTarget: source.profitAtTarget,
        checks: const [],
      ),
    );
    expect(find.text('Within your limits'), findsOneWidget);
    expect(find.text('Place paper order'), findsOneWidget);
    expect(find.text('Place paper order anyway'), findsNothing);
    expect(find.text('Discard'), findsOneWidget);
    expect(find.byType(OrderMeasure), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'Pair changes use matching previews; unsupported pairs cannot place',
    (tester) async {
      await boot(tester, location: '/check?pair=SOL%2FUSDT');
      expect(
        find.text('No order preview is available for this pair.'),
        findsOneWidget,
      );
      expect(find.byType(PrimaryButton), findsNothing);
      expect(find.byType(SecondaryButton), findsNothing);
      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text('BTC/USDT').last);
      await tester.pumpAndSettle();
      expect(find.text('Blocked: position too large'), findsOneWidget);
      expect(find.text('Place paper order anyway'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );
}
