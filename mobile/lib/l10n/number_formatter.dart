import 'package:flutter/widgets.dart';

import 'catalog_scope.dart';

/// Formats display values without converting decimal money strings to doubles.
class NumberFormatter {
  const NumberFormatter(this.languageCode);
  final String languageCode;
  bool get _turkish => languageCode == 'tr';

  static NumberFormatter of(BuildContext context) =>
      CatalogScope.maybeOf<NumberFormatter>(context) ??
      const NumberFormatter('en');

  String decimal(
    String value, {
    int minimumDecimals = 0,
    bool trimZero = false,
  }) {
    if (!RegExp(r'^[+−-]?\d+(?:\.\d+)?$').hasMatch(value)) return value;
    final negative = value.startsWith('-') || value.startsWith('−');
    final positive = value.startsWith('+');
    final parts = value.replaceFirst(RegExp(r'^[+−-]'), '').split('.');
    final whole = parts.first.replaceAllMapped(
      RegExp(r'(\d)(?=(\d{3})+(?!\d))'),
      (match) => '${match[1]}${_turkish ? '.' : ','}',
    );
    var fraction = parts.length > 1 ? parts[1] : '';
    if (trimZero && fraction == '0') fraction = '';
    fraction = fraction.padRight(minimumDecimals, '0');
    final separator = _turkish ? ',' : '.';
    return '${negative
        ? '−'
        : positive
        ? '+'
        : ''}$whole${fraction.isEmpty ? '' : '$separator$fraction'}';
  }

  String number(num value, {int decimalPlaces = 2, int? minimumDecimals}) {
    if (!value.isFinite) {
      throw ArgumentError.value(value, 'value', 'Must be finite');
    }
    var fixed = value.toStringAsFixed(decimalPlaces);
    if (minimumDecimals != null) {
      while (fixed.contains('.') &&
          fixed.endsWith('0') &&
          fixed.length - fixed.indexOf('.') - 1 > minimumDecimals) {
        fixed = fixed.substring(0, fixed.length - 1);
      }
      if (fixed.endsWith('.')) fixed = fixed.substring(0, fixed.length - 1);
    }
    return decimal(fixed);
  }

  /// Input is percentage points: 2.75 displays as 2.75%, not 275%.
  String percent(num value, {int decimalPlaces = 2, int? minimumDecimals}) =>
      unit(
        number(
          value,
          decimalPlaces: decimalPlaces,
          minimumDecimals: minimumDecimals,
        ),
        '%',
      );

  String unit(String magnitude, String unit) =>
      unit == '%' && _turkish ? '%$magnitude' : '$magnitude$unit';

  /// Localizes numeric runs in English fixture copy, preserving clocks.
  String fixtureText(String text) => text.replaceAllMapped(
    RegExp(r'[−+-]?\d+(?::\d+|(?:,\d{3})*(?:\.\d+)?%?)'),
    (match) {
      final token = match[0]!;
      if (token.contains(':')) return token;
      final percent = token.endsWith('%');
      final value = decimal(token.replaceAll(',', '').replaceAll('%', ''));
      return percent ? unit(value, '%') : value;
    },
  );
}
