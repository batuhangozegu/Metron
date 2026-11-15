import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/models.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/data/sample_data.dart';
import 'package:metron/features/position/position_screen.dart';
import 'package:metron/features/position/position_strings.dart';
import 'package:metron/features/position/price_ruler.dart';
import 'package:metron/widgets/widgets.dart';

import '../test_app.dart';

Future<void> pumpPosition(
  WidgetTester tester, {
  String id = 'btc',
  bool light = false,
  List<Position>? positions,
  List<CoinPrice>? prices,
  double height = 844,
  double textScale = 1,
}) async {
  tester.view.physicalSize = Size(390, height);
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);
  await pumpTestApp(
    tester,
    ProviderScope(
      overrides: [
        if (positions != null) positionsProvider.overrideWithValue(positions),
        if (prices != null) pricesProvider.overrideWithValue(prices),
      ],
      child: Builder(
        builder: (context) => MediaQuery(
          data: MediaQuery.of(context)
              .copyWith(textScaler: TextScaler.linear(textScale)),
          child: PositionScreen(positionId: id),
        ),
      ),
    ),
    light: light,
  );
}

Position changedPosition({bool stopOnExchange = true, String pnl = '-125.01'}) {
  final original = SampleData.positions.first;
  final d = original.detail!;
  return Position(
    id: original.id,
    pair: original.pair,
    coinClass: original.coinClass,
    distanceToStopPercent: original.distanceToStopPercent,
    moneyAtRisk: original.moneyAtRisk,
    unrealizedPnl: pnl,
    detail: PositionDetail(
      quantity: d.quantity,
      entry: d.entry,
      stop: d.stop,
      target: d.target,
      valueNow: d.valueNow,
      shareOfAccountPercent: d.shareOfAccountPercent,
      openedAt: d.openedAt,
      stopOnExchange: stopOnExchange,
      returnPercent: -1.5,
      returnR: -.5,
    ),
  );
}

void main() {
  for (final light in [false, true]) {
    testWidgets('BTC detail matches mockup at 390px, light=$light', (
      tester,
    ) async {
      await pumpPosition(tester, light: light);
      for (final text in [
        'BTC/USDT',
        'BTC/ETH',
        'Unrealized result',
        '93,508.00',
        '96,400.00',
        '102,184.00',
        '97,846.00',
        '0.08645 BTC',
        '8,458.79 USDT',
        '250.00 USDT',
        '33.33% of 40%',
        '12 Nov 08:14 UTC',
        'Safe',
        '4.43% away.',
        'Tighten stop',
        'Close position',
        '−1R',
        '+2R',
      ]) {
        expect(find.text(text), findsOneWidget);
      }
      expect(find.byType(BottomNavigationBar), findsNothing);
      expect(tester.widget<PnlText>(find.byType(PnlText).first).value, 125.01);
      final paint =
          tester
                  .widget<CustomPaint>(
                    find.byKey(const ValueKey('position-price-ruler')),
                  )
                  .painter!
              as PriceRulerPainter;
      expect(paint.entry, closeTo(1 / 3, .001));
      expect(paint.now, closeTo(.5, .001));
      expect(tester.getSize(find.byType(PrimaryButton)).height, 44);
      expect(tester.getSize(find.byType(SecondaryButton)).height, 44);
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('Loss and absent stop use explicit result and risk states', (
    tester,
  ) async {
    await pumpPosition(
      tester,
      positions: [changedPosition(stopOnExchange: false)],
    );
    expect(tester.widget<PnlText>(find.byType(PnlText).first).value, -125.01);
    expect(find.text('▼ −125.01'), findsOneWidget);
    expect(find.text('Limit exceeded'), findsOneWidget);
    expect(find.byIcon(Icons.report_outlined), findsOneWidget);
    expect(find.text(PositionStrings.stopMissing), findsOneWidget);
    expect(find.text('Safe'), findsNothing);
    expect(find.text('4.43% away.'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Unknown ID and summary-only position do not invent detail', (
    tester,
  ) async {
    await pumpPosition(tester, id: 'unknown');
    expect(find.text(PositionStrings.notFound), findsOneWidget);
    expect(find.byType(PrimaryButton), findsNothing);
    await pumpPosition(tester, id: 'sol');
    expect(find.text('SOL/USDT'), findsOneWidget);
    expect(find.text(PositionStrings.detailsMissing), findsOneWidget);
    expect(find.byType(PriceRuler), findsNothing);
    expect(find.byType(PrimaryButton), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Missing price hides current marker; changed price moves it', (
    tester,
  ) async {
    await pumpPosition(tester, prices: []);
    expect(find.text(PositionStrings.priceMissing), findsOneWidget);
    expect(find.text('Now'), findsNothing);
    const price = CoinPrice(
      pair: 'BTC/USDT',
      coinClass: CoinClass.btcEth,
      price: '110000.00',
      changePercent: 1,
      atrPercent: 2,
      suggestedStopPercent: 3,
    );
    await pumpPosition(tester, prices: [price]);
    final paint =
        tester
                .widget<CustomPaint>(
                  find.byKey(const ValueKey('position-price-ruler')),
                )
                .painter!
            as PriceRulerPainter;
    expect(paint.now, 1);
    expect(find.text('110,000.00'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('Fixed actions explain sample limitations and dismiss cleanly', (
    tester,
  ) async {
    await pumpPosition(tester);
    for (final label in [PositionStrings.tighten, PositionStrings.close]) {
      await tester.tap(find.text(label));
      await tester.pumpAndSettle();
      expect(find.byType(AlertDialog), findsOneWidget);
      expect(find.textContaining(PositionStrings.previewOnly), findsOneWidget);
      await tester.tap(find.text(PositionStrings.dismiss));
      await tester.pumpAndSettle();
      expect(find.byType(AlertDialog), findsNothing);
    }
    expect(tester.takeException(), isNull);
  });

  testWidgets('Short viewport and enlarged text scroll above fixed actions', (
    tester,
  ) async {
    await pumpPosition(tester, height: 640, textScale: 1.3);
    await tester.drag(
      find.byType(SingleChildScrollView),
      const Offset(0, -600),
    );
    await tester.pumpAndSettle();
    expect(find.text(PositionStrings.stopRule).hitTestable(), findsOneWidget);
    expect(find.text(PositionStrings.close).hitTestable(), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
