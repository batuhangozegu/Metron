import 'package:flutter/material.dart';

import '../../data/models.dart';
import '../../theme/metron_colors.dart';
import '../../widgets/num.dart';
import 'position_format.dart';
import 'position_strings.dart';

/// A price range rather than the shared Measure's risk usage against a limit.
class PriceRuler extends StatelessWidget {
  const PriceRuler({super.key, required this.detail, required this.price});
  final PositionDetail detail;
  final String? price;

  @override
  Widget build(BuildContext context) {
    final stop = double.parse(detail.stop);
    final target = double.parse(detail.target);
    final entry = double.parse(detail.entry);
    final current = double.tryParse(price ?? '');
    final range = target - stop;
    double fraction(double value) =>
        range == 0 ? 0 : ((value - stop) / range).clamp(0, 1);
    final now = current == null ? null : fraction(current);
    final risk = (entry - stop).abs();
    String rLabel(double value) {
      if (risk == 0) return '—';
      final r = (value - entry) / risk;
      final magnitude = r
          .abs()
          .toStringAsFixed(1)
          .replaceFirst(RegExp(r'\.0$'), '');
      return '${r < 0 ? '−' : '+'}${magnitude}R';
    }

    final c = context.colors;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            for (final (label, value, color, alignment) in [
              (
                PositionStrings.stop,
                detail.stop,
                c.loss,
                CrossAxisAlignment.start,
              ),
              (
                PositionStrings.entry,
                detail.entry,
                c.subtle,
                CrossAxisAlignment.center,
              ),
              (
                PositionStrings.target,
                detail.target,
                c.profit,
                CrossAxisAlignment.end,
              ),
            ])
              Expanded(
                child: Column(
                  crossAxisAlignment: alignment,
                  spacing: 3,
                  children: [
                    Text(
                      label.toUpperCase(),
                      style: TextStyle(
                        fontSize: 11,
                        letterSpacing: .88,
                        color: color,
                      ),
                    ),
                    FittedBox(
                      fit: BoxFit.scaleDown,
                      child: Num(
                        formatDecimal(value),
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
          ],
        ),
        const SizedBox(height: 14),
        Semantics(
          label:
              '${PositionStrings.stop} ${formatDecimal(detail.stop)}, '
              '${PositionStrings.entry} ${formatDecimal(detail.entry)}, '
              '${PositionStrings.target} ${formatDecimal(detail.target)}, '
              '${current == null ? PositionStrings.priceMissing : '${PositionStrings.now} ${formatDecimal(price!)}'}',
          child: ExcludeSemantics(
            child: Column(
              children: [
                SizedBox(
                  width: double.infinity,
                  height: 46,
                  child: CustomPaint(
                    key: const ValueKey('position-price-ruler'),
                    painter: PriceRulerPainter(
                      entry: fraction(entry),
                      now: now,
                      colors: c,
                    ),
                  ),
                ),
                SizedBox(
                  height: 20,
                  child: Stack(
                    children: [
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Num(
                          rLabel(stop),
                          style: TextStyle(fontSize: 11, color: c.subtle),
                        ),
                      ),
                      if (now != null)
                        Align(
                          alignment: Alignment(now * 2 - 1, 0),
                          child: const Text(
                            PositionStrings.now,
                            style: TextStyle(fontSize: 11),
                          ),
                        ),
                      Align(
                        alignment: Alignment.centerRight,
                        child: Num(
                          rLabel(target),
                          style: TextStyle(fontSize: 11, color: c.subtle),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        if (current == null)
          Text(
            PositionStrings.priceMissing,
            style: TextStyle(fontSize: 12, color: c.muted),
          ),
      ],
    );
  }
}

class PriceRulerPainter extends CustomPainter {
  const PriceRulerPainter({
    required this.entry,
    required this.now,
    required this.colors,
  });
  final double entry;
  final double? now;
  final MetronColors colors;

  @override
  void paint(Canvas canvas, Size size) {
    final width = size.width - 2;
    final paint = Paint()..color = colors.subtle;
    for (var i = 0; i <= 10; i++) {
      final x = 1 + width * i / 10;
      canvas.drawLine(Offset(x, 0), Offset(x, 8), paint);
    }
    final track = RRect.fromRectAndRadius(
      Rect.fromLTWH(1, 13, width, 24),
      const Radius.circular(4),
    );
    canvas.save();
    canvas.clipRRect(track);
    canvas.drawRRect(track, paint..color = colors.profitTint);
    canvas.drawRect(
      Rect.fromLTWH(1, 13, width * entry, 24),
      paint..color = colors.lossTint,
    );
    paint
      ..color = colors.muted
      ..strokeWidth = 2;
    canvas.drawLine(
      Offset(1 + width * entry, 13),
      Offset(1 + width * entry, 37),
      paint,
    );
    canvas.restore();
    if (now case final current?) {
      final x = 1 + width * current;
      paint
        ..color = colors.text
        ..strokeWidth = 3
        ..strokeCap = StrokeCap.round;
      canvas.drawLine(Offset(x, 8), Offset(x, 42), paint);
    }
  }

  @override
  bool shouldRepaint(covariant PriceRulerPainter oldDelegate) =>
      entry != oldDelegate.entry ||
      now != oldDelegate.now ||
      colors != oldDelegate.colors;
}
