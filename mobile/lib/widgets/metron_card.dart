import 'package:flutter/material.dart';

import '../theme/tokens.dart';
import '../theme/metron_colors.dart';
import '../data/risk_level.dart';
import 'risk_badge.dart';

/// Sharp outlined panel; risk adds a signal rail and stronger limit treatment.
class MetronCard extends StatelessWidget {
  const MetronCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(MetronSize.cardPadding),
    this.risk,
  });
  final Widget child;
  final EdgeInsetsGeometry padding;
  final RiskLevel? risk;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final signal = risk?.color(c);
    return Card(
      child: Container(
        decoration: BoxDecoration(
          border: Border(
            top: BorderSide(
              color: signal ?? Color.lerp(c.line, c.limitEdge, .4)!,
            ),
            left: BorderSide(
              color: signal ?? c.line,
              width: risk == null ? 1 : 3,
            ),
            right: BorderSide(color: signal ?? c.line),
            bottom: BorderSide(color: signal ?? c.line),
          ),
          gradient: risk == RiskLevel.limit
              ? LinearGradient(colors: [c.limitTint, c.surface])
              : null,
        ),
        padding: padding,
        child: child,
      ),
    );
  }
}
