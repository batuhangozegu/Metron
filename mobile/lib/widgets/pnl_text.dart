import 'package:flutter/material.dart';

import '../l10n/number_formatter.dart';
import '../l10n/app_strings.dart';
import '../theme/metron_colors.dart';
import 'num.dart';

/// Signed result with a direction arrow. Zero is neutral (— ±0.00).
/// Pass an absolute-value formatter for grouped currency or custom precision.
class PnlText extends StatelessWidget {
  const PnlText(
    this.value, {
    super.key,
    this.decimalPlaces = 2,
    this.suffix = '',
    this.style,
    this.format,
  });
  final double value;
  final int decimalPlaces;
  final String suffix;
  final TextStyle? style;
  final String Function(double magnitude)? format;

  @override
  Widget build(BuildContext context) {
    if (!value.isFinite) {
      throw ArgumentError.value(value, 'value', 'Must be finite');
    }
    final c = context.colors;
    final (prefix, word, color) = value > 0
        ? ('▲ +', AppStrings.of(context).profit, c.profit)
        : value < 0
        ? ('▼ −', AppStrings.of(context).loss, c.loss)
        : ('— ±', AppStrings.of(context).unchanged, c.muted);
    final magnitude =
        format?.call(value.abs()) ??
        NumberFormatter.of(context)
            .number(value.abs(), decimalPlaces: decimalPlaces);
    return Num(
      '$prefix${NumberFormatter.of(context).unit(magnitude, suffix)}',
      style: (style ?? const TextStyle()).copyWith(color: color),
      semanticsLabel:
          '$word ${NumberFormatter.of(context).unit(magnitude, suffix)}',
    );
  }
}
