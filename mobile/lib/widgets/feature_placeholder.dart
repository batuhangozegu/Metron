import 'package:flutter/material.dart';

import '../l10n/app_strings.dart';
import '../theme/tokens.dart';

/// Temporary scaffold only. Each screen agent replaces its feature file.
class FeaturePlaceholder extends StatelessWidget {
  const FeaturePlaceholder({super.key, required this.title, this.detail});
  final String title;
  final Widget? detail;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(title)),
    body: SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(MetronSpace.lg),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(AppStrings.todo),
            if (detail != null) ...[
              const SizedBox(height: MetronSpace.lg),
              detail!,
            ],
          ],
        ),
      ),
    ),
  );
}
