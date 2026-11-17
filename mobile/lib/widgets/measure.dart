import 'package:flutter/material.dart';

import '../data/risk_level.dart';
import '../l10n/app_strings.dart';
import '../theme/metron_colors.dart';
import '../theme/tokens.dart';
import 'num.dart';
import 'risk_badge.dart';

/// Ten-division risk ruler. Values use the same unit (e.g. 2.75 and 6 percent).
/// Fill clamps to the track; the labels retain an exceeded value in full.
class Measure extends StatelessWidget {
  const Measure({
    super.key,
    required this.value,
    required this.limit,
    this.label,
    this.unit = '',
    this.decimalPlaces = 2,
  });
  final double value;
  final double limit;
  final String? label;
  final String unit;
  final int decimalPlaces;

  @override
  Widget build(BuildContext context) {
    final level = RiskLevel.forUsage(value, limit);
    final colors = context.colors;
    final valueText = '${value.toStringAsFixed(decimalPlaces)}$unit';
    final limitText = '${limit.toStringAsFixed(decimalPlaces)}$unit';
    return Semantics(
      label: [
        ?label,
        '$valueText ${AppStrings.of(context).outOf} $limitText',
        level.label(context),
      ].join(', '),
      excludeSemantics: true,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        mainAxisSize: MainAxisSize.min,
        children: [
          Wrap(
            alignment: WrapAlignment.spaceBetween,
            crossAxisAlignment: WrapCrossAlignment.center,
            spacing: MetronSpace.sm,
            runSpacing: MetronSpace.xs,
            children: [
              if (label != null)
                Text(label!, style: Theme.of(context).textTheme.bodySmall),
              RiskBadge(level),
            ],
          ),
          const SizedBox(height: MetronSpace.sm),
          Wrap(
            crossAxisAlignment: WrapCrossAlignment.center,
            spacing: MetronSpace.sm,
            children: [
              Num(valueText),
              Text(AppStrings.of(context).outOf),
              Num(limitText, style: TextStyle(color: colors.subtle)),
            ],
          ),
          const SizedBox(height: MetronSpace.sm),
          SizedBox(
            height: 30,
            child: CustomPaint(
              key: const ValueKey('measure-ruler'),
              painter: MeasurePainter(
                fraction: (value / limit).clamp(0, 1),
                fill: level.color(colors),
                track: colors.raised,
                tick: colors.subtle,
                caution: colors.caution,
                limit: colors.limit,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Paints eleven ticks (ten equal divisions), then the bar and both markers.
class MeasurePainter extends CustomPainter {
  const MeasurePainter({
    required this.fraction,
    required this.fill,
    required this.track,
    required this.tick,
    required this.caution,
    required this.limit,
  });
  final double fraction;
  final Color fill, track, tick, caution, limit;
  static const divisions = 10;

  @override
  void paint(Canvas canvas, Size size) {
    final width = size.width - 2;
    final paint = Paint()..strokeWidth = 1;
    for (var i = 0; i <= divisions; i++) {
      final x = 1 + width * i / divisions;
      canvas.drawLine(Offset(x, 0), Offset(x, 8), paint..color = tick);
    }
    final bar = RRect.fromRectAndRadius(
      Rect.fromLTWH(1, 16, width, 8),
      const Radius.circular(2),
    );
    canvas.drawRRect(bar, paint..color = track);
    canvas.save();
    canvas.clipRRect(bar);
    canvas.drawRect(
      Rect.fromLTWH(1, 16, width * fraction, 8),
      paint..color = fill,
    );
    canvas.restore();
    paint.strokeWidth = 2;
    final cautionX = 1 + width * MetronRisk.cautionAtFractionOfLimit;
    canvas.drawLine(
      Offset(cautionX, 12),
      Offset(cautionX, 28),
      paint..color = caution,
    );
    canvas.drawLine(
      Offset(1 + width, 12),
      Offset(1 + width, 28),
      paint..color = limit,
    );
  }

  @override
  bool shouldRepaint(covariant MeasurePainter oldDelegate) =>
      fraction != oldDelegate.fraction ||
      fill != oldDelegate.fill ||
      track != oldDelegate.track ||
      tick != oldDelegate.tick ||
      caution != oldDelegate.caution ||
      limit != oldDelegate.limit;
}
