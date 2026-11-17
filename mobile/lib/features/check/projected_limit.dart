import 'package:flutter/material.dart';

import '../../l10n/number_formatter.dart';
import '../../data/models.dart';
import '../../data/risk_level.dart';
import '../../theme/metron_colors.dart';
import '../../widgets/widgets.dart';
import 'check_strings.dart';

/// Compact before/after projection; hatching distinguishes the proposed order.
class ProjectedLimit extends StatelessWidget {
  const ProjectedLimit({super.key, required this.check});
  final LimitCheck check;
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final color = check.level.color(c);
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
      decoration: BoxDecoration(
        border: Border(top: BorderSide(color: c.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: 8,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  CheckStrings.of(context).limitName(check),
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              if (check.brokenRule || check.level == RiskLevel.limit)
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(check.level.icon, size: 14, color: color),
                    const SizedBox(width: 5),
                    Text(
                      check.brokenRule
                          ? CheckStrings.of(context).broken
                          : CheckStrings.of(context).exceeded,
                      style: TextStyle(fontSize: 12, color: color),
                    ),
                  ],
                )
              else
                RiskBadge(check.level),
            ],
          ),
          SizedBox(
            height: 6,
            child: CustomPaint(
              painter: _ProjectionPainter(
                before: (check.before / check.limit).clamp(0, 1),
                after: (check.after / check.limit).clamp(0, 1),
                color: color,
                track: c.raised,
              ),
            ),
          ),
          Num(
            CheckStrings.of(context).limitValue(
              NumberFormatter.of(context).fixtureText(check.valueLabel),
            ),
            style: TextStyle(fontSize: 12, color: c.muted),
          ),
        ],
      ),
    );
  }
}

class _ProjectionPainter extends CustomPainter {
  const _ProjectionPainter({
    required this.before,
    required this.after,
    required this.color,
    required this.track,
  });
  final double before, after;
  final Color color, track;
  @override
  void paint(Canvas canvas, Size size) {
    final p = Paint();
    canvas.save();
    canvas.clipRRect(
      RRect.fromRectAndRadius(Offset.zero & size, const Radius.circular(2)),
    );
    canvas.drawRect(Offset.zero & size, p..color = track);
    canvas.drawRect(
      Rect.fromLTWH(0, 0, before * size.width, size.height),
      p..color = color,
    );
    canvas.clipRect(
      Rect.fromLTRB(before * size.width, 0, after * size.width, size.height),
    );
    p
      ..color = color
      ..strokeWidth = 2;
    for (double x = 0; x < size.width + size.height; x += 5) {
      canvas.drawLine(Offset(x, 0), Offset(x - size.height, size.height), p);
    }
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _ProjectionPainter old) =>
      before != old.before ||
      after != old.after ||
      color != old.color ||
      track != old.track;
}
