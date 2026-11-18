import 'package:flutter/material.dart';

import '../theme/tokens.dart';

/// Locale-aware display casing; the spoken label retains sentence casing.
String commandUppercase(BuildContext context, String text) =>
    Localizations.localeOf(context).languageCode == 'tr'
    ? text.replaceAll('i', 'İ').replaceAll('ı', 'I').toUpperCase()
    : text.toUpperCase();

class CommandTitle extends StatelessWidget {
  const CommandTitle(this.text, {super.key, this.style});
  final String text;
  final TextStyle? style;

  @override
  Widget build(BuildContext context) => Text(
    commandUppercase(context, text),
    semanticsLabel: text,
    style: MetronType.title.merge(style),
  );
}
