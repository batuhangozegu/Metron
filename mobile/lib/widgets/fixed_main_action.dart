import 'package:flutter/material.dart';

import '../theme/metron_colors.dart';
import '../theme/tokens.dart';
import 'metron_button.dart';

/// Place after Expanded(body) in a Column, or in Scaffold.bottomNavigationBar.
/// The app shell uses it immediately above the four-item navigation bar.
class FixedMainAction extends StatelessWidget {
  const FixedMainAction({
    super.key,
    required this.label,
    required this.onPressed,
  });
  final String label;
  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) => ColoredBox(
    color: context.colors.bg,
    child: SafeArea(
      top: false,
      bottom: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          MetronSpace.lg,
          MetronSpace.md,
          MetronSpace.lg,
          MetronSpace.md,
        ),
        child: SizedBox(
          width: double.infinity,
          child: PrimaryButton(label: label, onPressed: onPressed),
        ),
      ),
    ),
  );
}
