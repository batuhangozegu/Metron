import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/risk_level.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/widgets/measure.dart';
import 'package:metron/widgets/risk_badge.dart';

import 'test_app.dart';

void main() {
  for (final light in [false, true]) {
    testWidgets(
      'Measure labels quantity, limit and risk (${light ? 'light' : 'dark'})',
      (tester) async {
        final semantics = tester.ensureSemantics();
        await pumpTestApp(
          tester,
          const Measure(value: 2.75, limit: 6, unit: '%', label: 'Open risk'),
          light: light,
        );
        expect(find.text('2.75%'), findsOneWidget);
        expect(find.text('6.00%'), findsOneWidget);
        expect(
          find.bySemanticsLabel('Open risk, 2.75% of 6.00%, Safe'),
          findsOneWidget,
        );
        final paint =
            tester
                    .widget<CustomPaint>(
                      find.byKey(const ValueKey('measure-ruler')),
                    )
                    .painter!
                as MeasurePainter;
        expect(paint.fraction, closeTo(2.75 / 6, .00001));
        expect(
          paint.fill,
          light ? MetronColors.light.safe : MetronColors.dark.safe,
        );
        expect(
          paint.caution,
          light ? MetronColors.light.caution : MetronColors.dark.caution,
        );
        expect(
          paint.limit,
          light ? MetronColors.light.limit : MetronColors.dark.limit,
        );
        expect(MeasurePainter.divisions, 10);
        semantics.dispose();
      },
    );
  }

  testWidgets('Measure handles thresholds and clamps only the painted fill', (
    tester,
  ) async {
    for (final (value, level, fraction) in [
      (0.0, RiskLevel.safe, 0.0),
      (74.99, RiskLevel.safe, .7499),
      (75.0, RiskLevel.caution, .75),
      (100.0, RiskLevel.caution, 1.0),
      (125.0, RiskLevel.limit, 1.0),
    ]) {
      await pumpTestApp(tester, Measure(value: value, limit: 100));
      expect(tester.widget<RiskBadge>(find.byType(RiskBadge)).level, level);
      final painter =
          tester
                  .widget<CustomPaint>(
                    find.byKey(const ValueKey('measure-ruler')),
                  )
                  .painter!
              as MeasurePainter;
      expect(painter.fraction, closeTo(fraction, .00001));
      expect(painter.fill, level.color(MetronColors.dark));
      expect(find.text(value.toStringAsFixed(2)), findsWidgets);
    }
  });

  testWidgets('Measure retains labels at narrow widths and large text', (
    tester,
  ) async {
    await pumpTestApp(
      tester,
      MediaQuery(
        data: const MediaQueryData(textScaler: TextScaler.linear(2)),
        child: const SizedBox(
          width: 180,
          child: Measure(value: 8, limit: 6, label: 'Open risk', unit: '%'),
        ),
      ),
    );
    expect(tester.takeException(), isNull);
  });

  test('invalid risk quantities fail explicitly', () {
    for (final (value, limit) in [
      (1.0, 0.0),
      (-1.0, 2.0),
      (double.nan, 1.0),
      (1.0, double.infinity),
    ]) {
      expect(() => RiskLevel.forUsage(value, limit), throwsArgumentError);
    }
  });
}
