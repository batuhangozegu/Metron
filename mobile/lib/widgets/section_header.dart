import 'package:flutter/material.dart';

import '../theme/metron_colors.dart';
import '../theme/tokens.dart';

/// Uppercase section label with optional trailing value or action.
class SectionHeader extends StatelessWidget {
  const SectionHeader(this.title, {super.key, this.trailing});
  final String title;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    // Dart's Unicode casing is locale-independent; Turkish distinguishes i/ı.
    final uppercase = Localizations.localeOf(context).languageCode == 'tr'
        ? title.replaceAll('i', 'İ').replaceAll('ı', 'I').toUpperCase()
        : title.toUpperCase();
    return Row(
      children: [
        Expanded(
          child: Text(
            uppercase,
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
}
