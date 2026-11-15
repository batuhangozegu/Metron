import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/widgets/metron_button.dart';
import 'package:metron/widgets/metron_card.dart';

import 'test_app.dart';

void main() {
  testWidgets('Shared buttons are 44 px and call their actions', (
    tester,
  ) async {
    var taps = 0;
    await pumpTestApp(
      tester,
      Column(
        children: [
          PrimaryButton(label: 'Primary', onPressed: () => taps++),
          SecondaryButton(label: 'Secondary', onPressed: () => taps++),
          const PrimaryButton(label: 'Disabled', onPressed: null),
        ],
      ),
    );
    expect(tester.getSize(find.byType(FilledButton).first).height, 44);
    expect(tester.getSize(find.byType(OutlinedButton)).height, 44);
    await tester.tap(find.text('Primary'));
    await tester.tap(find.text('Secondary'));
    await tester.tap(find.text('Disabled'));
    expect(taps, 2);
  });

  testWidgets('Shared cards have a one-pixel line and zero elevation', (
    tester,
  ) async {
    await pumpTestApp(tester, const MetronCard(child: Text('Content')));
    final theme = Theme.of(tester.element(find.byType(Card))).cardTheme;
    expect(theme.elevation, 0);
    expect((theme.shape! as RoundedRectangleBorder).side.width, 1);
    expect(theme.shadowColor, Colors.transparent);
  });
}
