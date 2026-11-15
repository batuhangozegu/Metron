import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/risk_level.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/widgets/risk_badge.dart';

import 'test_app.dart';

void main() {
  for (final light in [false, true]) {
    for (final (level, label, icon) in [
      (RiskLevel.safe, 'Safe', Icons.check_circle_outline),
      (RiskLevel.caution, 'Caution', Icons.warning_amber_rounded),
      (RiskLevel.limit, 'Limit exceeded', Icons.report_outlined),
    ]) {
      testWidgets(
        '$label always includes icon, word and color (${light ? 'light' : 'dark'})',
        (tester) async {
          final semantics = tester.ensureSemantics();
          await pumpTestApp(tester, RiskBadge(level), light: light);
          expect(find.text(label), findsOneWidget);
          expect(find.byIcon(icon), findsOneWidget);
          expect(find.bySemanticsLabel(label), findsOneWidget);
          final color = level.color(
            light ? MetronColors.light : MetronColors.dark,
          );
          expect(tester.widget<Icon>(find.byIcon(icon)).color, color);
          expect(tester.widget<Text>(find.text(label)).style!.color, color);
          semantics.dispose();
        },
      );
    }
  }
}
