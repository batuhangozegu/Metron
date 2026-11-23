import 'package:flutter/material.dart';

import '../../data/api_client.dart';
import '../../data/risk_level.dart';
import '../../l10n/live_strings.dart';
import '../../l10n/number_formatter.dart';
import '../../theme/metron_colors.dart';
import '../../widgets/widgets.dart';
import '../../widgets/live_state.dart';
import '../../widgets/segmented_bar.dart';
import 'visual_ratio.dart';

/// Server rule status drives color; meters visualize its returned values only.
class ProjectedLimit extends StatelessWidget {
  const ProjectedLimit({super.key, required this.rule});
  final Json rule;
  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context), f = NumberFormatter.of(context);
    final level = rule['status'] == 'fail'
        ? RiskLevel.limit
        : rule['status'] == 'caution'
        ? RiskLevel.caution
        : RiskLevel.safe;
    final values =
        (rule['values'] as Map?)?.cast<String, String>() ?? <String, String>{};
    final limit = values['limit'] ?? values['limit_percent'];
    final current = values['current'] ?? values['current_percent'];
    final proposed =
        values['proposed'] ?? values['proposed_percent'] ?? current;
    return Container(
      key: ValueKey('rule-${rule['rule_id']}'),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: level == RiskLevel.limit ? context.colors.limitTint : null,
        border: Border(top: BorderSide(color: context.colors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: 8,
        children: [
          Row(
            children: [
              Icon(level.icon, color: level.color(context.colors), size: 18),
              const SizedBox(width: 8),
              Text(s.status(rule['status'] as String)),
            ],
          ),
          Text(rule[s.tr ? 'message_tr' : 'message_en'] as String),
          if (limit != null && proposed != null)
            SizedBox(
              height: 12,
              child: CustomPaint(
                painter: _ProjectionPainter(
                  before: visualFraction(
                    decimalTicks(current ?? '0'),
                    decimalTicks(limit),
                  ),
                  after: visualFraction(
                    decimalTicks(proposed),
                    decimalTicks(limit),
                  ),
                  color: level.color(context.colors),
                  startColor: level == RiskLevel.limit
                      ? context.colors.caution
                      : level.color(context.colors),
                  track: context.colors.raised,
                ),
              ),
            ),
          for (final value in values.entries)
            ValueRow(s.ruleValue(value.key), f.decimal(value.value)),
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
    required this.startColor,
    required this.track,
  });
  final double before, after;
  final Color color, startColor, track;
  @override
  void paint(Canvas canvas, Size size) => paintSegmentedBar(
    canvas,
    Offset.zero & size,
    fraction: after,
    solidFraction: before,
    color: color,
    startColor: startColor,
    track: track,
  );
  @override
  bool shouldRepaint(covariant _ProjectionPainter old) =>
      before != old.before ||
      after != old.after ||
      color != old.color ||
      startColor != old.startColor ||
      track != old.track;
}
