import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import '../theme/tokens.dart';

/// Use for every number, including formatted currency strings and percentages.
class Num extends StatelessWidget {
  const Num(
    this.data, {
    super.key,
    this.style,
    this.textAlign,
    this.maxLines,
    this.overflow,
    this.semanticsLabel,
  });

  final String data;
  final TextStyle? style;
  final TextAlign? textAlign;
  final int? maxLines;
  final TextOverflow? overflow;
  final String? semanticsLabel;

  @override
  Widget build(BuildContext context) => Text(
    data,
    style: GoogleFonts.jetBrainsMono(textStyle: MetronType.figure.merge(style))
        .copyWith(fontFeatures: const [FontFeature.tabularFigures()]),
    textAlign: textAlign,
    maxLines: maxLines,
    overflow: overflow,
    semanticsLabel: semanticsLabel,
  );
}
