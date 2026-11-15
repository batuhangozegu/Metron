import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/widgets/num.dart';
import 'package:metron/widgets/pnl_text.dart';

import 'test_app.dart';

void main() {
  for (final light in [false, true]) {
    testWidgets(
      'PnL supplies arrow, sign, semantic description and color (${light ? 'light' : 'dark'})',
      (tester) async {
        final semantics = tester.ensureSemantics();
        final colors = light ? MetronColors.light : MetronColors.dark;
        await pumpTestApp(
          tester,
          const Column(
            children: [PnlText(125.01), PnlText(-55.59), PnlText(0)],
          ),
          light: light,
        );
        for (final (label, description, color) in [
          ('▲ +125.01', 'Profit 125.01', colors.profit),
          ('▼ −55.59', 'Loss 55.59', colors.loss),
          ('— ±0.00', 'Unchanged 0.00', colors.muted),
        ]) {
          final text = tester.widget<Text>(find.text(label));
          expect(text.style!.color, color);
          expect(text.style!.fontFamily, contains('JetBrainsMono'));
          expect(
            text.style!.fontFeatures,
            contains(const FontFeature.tabularFigures()),
          );
          expect(find.bySemanticsLabel(description), findsOneWidget);
        }
        expect(find.byType(Num), findsNWidgets(3));
        semantics.dispose();
      },
    );
  }
  testWidgets('PnL formatter gets magnitude and retains sign and suffix', (
    tester,
  ) async {
    await pumpTestApp(
      tester,
      PnlText(
        -1234.5,
        suffix: ' USDT',
        format: (magnitude) {
          expect(magnitude, 1234.5);
          return '1,234.50';
        },
      ),
    );
    expect(find.text('▼ −1,234.50 USDT'), findsOneWidget);
  });
}
