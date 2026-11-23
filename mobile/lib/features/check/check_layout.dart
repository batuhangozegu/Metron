import 'package:flutter/material.dart';

import '../../data/live_models.dart';
import '../../data/risk_level.dart';
import '../../l10n/live_strings.dart';
import '../../widgets/widgets.dart';
import 'order_measure.dart';
import 'projected_limit.dart';

/// Shared check layout retains the sizing ruler and per-rule projections.
class CheckLayout extends StatelessWidget {
  const CheckLayout({super.key, required this.check});
  final RiskCheck check;
  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context);
    final level = check.blocked
        ? RiskLevel.limit
        : check.json['verdict'] == 'caution'
        ? RiskLevel.caution
        : RiskLevel.safe;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: 14,
      children: [
        MetronCard(
          risk: level,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            spacing: 8,
            children: [
              RiskBadge(level),
              CommandTitle(
                check.blocked
                    ? s.blocked
                    : level == RiskLevel.caution
                    ? s.caution
                    : s.allowed,
                style: const TextStyle(fontSize: 23),
              ),
              Text(
                '${check.value('symbol')} · ${check.value('side') == 'short' ? s.short : s.long} · ${check.value('market')}',
              ),
            ],
          ),
        ),
        OrderMeasure(check: check),
        SectionHeader(s.rules),
        MetronCard(
          padding: EdgeInsets.zero,
          child: Column(
            children: [
              for (final rule in check.rules) ProjectedLimit(rule: rule),
            ],
          ),
        ),
      ],
    );
  }
}
