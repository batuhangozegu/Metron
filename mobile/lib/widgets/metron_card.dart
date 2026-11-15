import 'package:flutter/material.dart';

import '../theme/tokens.dart';

/// Flat, one-pixel outlined surface; appearance comes from MetronTheme.
class MetronCard extends StatelessWidget {
  const MetronCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(MetronSize.cardPadding),
  });
  final Widget child;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) => Card(
    child: Padding(padding: padding, child: child),
  );
}
