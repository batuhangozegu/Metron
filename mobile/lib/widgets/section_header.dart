import 'package:flutter/material.dart';

import '../theme/metron_colors.dart';
import '../theme/tokens.dart';

/// Uppercase section label with optional trailing value or action.
class SectionHeader extends StatelessWidget {
  const SectionHeader(this.title, {super.key, this.trailing});
  final String title;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) => Row(
    children: [
      Expanded(
        child: Text(
          title.toUpperCase(),
          style: Theme.of(context).textTheme.labelSmall!
              .copyWith(color: context.colors.subtle),
        ),
      ),
      if (trailing != null) ...[
        const SizedBox(width: MetronSpace.sm),
        trailing!,
      ],
    ],
  );
}
