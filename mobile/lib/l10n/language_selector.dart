import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app_strings.dart';
import 'locale_provider.dart';

class LanguageSelector extends ConsumerWidget {
  const LanguageSelector({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) => Semantics(
    label: AppStrings.of(context).language,
    child: DropdownButton<Locale>(
      key: const ValueKey('language-selector'),
      value: ref.watch(localeProvider),
      hint: Text(AppStrings.of(context).language),
      icon: const Icon(Icons.language),
      underline: const SizedBox.shrink(),
      items: const [
        DropdownMenuItem(value: Locale('en'), child: Text('English')),
        DropdownMenuItem(value: Locale('tr'), child: Text('Türkçe')),
      ],
      onChanged: (locale) {
        if (locale != null) ref.read(localeProvider.notifier).select(locale);
      },
    ),
  );
}
