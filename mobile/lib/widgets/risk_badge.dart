import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../data/risk_level.dart';
import '../l10n/app_strings.dart';
import '../theme/metron_colors.dart';
import '../theme/tokens.dart';

extension RiskLevelPresentation on RiskLevel {
  String label(BuildContext context) => switch (this) {
    RiskLevel.safe => AppStrings.of(context).safe,
    RiskLevel.caution => AppStrings.of(context).caution,
    RiskLevel.limit => AppStrings.of(context).limit,
  };
  IconData get icon => switch (this) {
    RiskLevel.safe => Icons.check_circle_outline,
    RiskLevel.caution => Icons.warning_amber_rounded,
    RiskLevel.limit => Icons.report_outlined,
  };
  Color color(MetronColors c) => switch (this) {
    RiskLevel.safe => c.safe,
    RiskLevel.caution => c.caution,
    RiskLevel.limit => c.limit,
  };
  Color tint(MetronColors c) => switch (this) {
    RiskLevel.safe => c.safeTint,
    RiskLevel.caution => c.cautionTint,
    RiskLevel.limit => c.limitTint,
  };
}

/// Always displays the state as an icon, a word and its semantic color.
class RiskBadge extends StatelessWidget {
  const RiskBadge(this.level, {super.key});
  final RiskLevel level;

  @override
  Widget build(BuildContext context) {
    final color = level.color(context.colors);
    return Semantics(
      label: level.label(context),
      excludeSemantics: true,
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: MetronSpace.sm,
          vertical: MetronSpace.xs,
        ),
        decoration: BoxDecoration(
          color: level.tint(context.colors),
          borderRadius: BorderRadius.circular(MetronRadius.tag),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(level.icon, size: 16, color: color),
            const SizedBox(width: MetronSpace.xs),
            Flexible(
              child: Text(
                level.label(context),
                style: GoogleFonts.inter(
                  textStyle: Theme.of(context).textTheme.bodySmall,
                  color: color,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
