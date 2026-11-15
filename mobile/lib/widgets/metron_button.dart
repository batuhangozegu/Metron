import 'package:flutter/material.dart';

import '../theme/tokens.dart';

/// Neutral primary action. A null callback disables the control.
class PrimaryButton extends StatelessWidget {
  const PrimaryButton({
    super.key,
    required this.label,
    required this.onPressed,
  });
  final String label;
  final VoidCallback? onPressed;
  @override
  Widget build(BuildContext context) => SizedBox(
    height: MetronSize.control,
    child: FilledButton(onPressed: onPressed, child: Text(label)),
  );
}

/// Outlined secondary action, using the same 44 px control size.
class SecondaryButton extends StatelessWidget {
  const SecondaryButton({
    super.key,
    required this.label,
    required this.onPressed,
  });
  final String label;
  final VoidCallback? onPressed;
  @override
  Widget build(BuildContext context) => SizedBox(
    height: MetronSize.control,
    child: OutlinedButton(onPressed: onPressed, child: Text(label)),
  );
}
