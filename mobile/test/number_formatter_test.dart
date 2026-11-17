import 'package:flutter_test/flutter_test.dart';
import 'package:metron/l10n/number_formatter.dart';

void main() {
  for (final language in ['en', 'tr']) {
    final f = NumberFormatter(language);
    final tr = language == 'tr';
    group(language, () {
      test('amounts, grouping, and real minus signs', () {
        expect(f.number(25000), tr ? '25.000,00' : '25,000.00');
        expect(f.number(-312.5), tr ? '−312,50' : '−312.50');
        expect(f.number(0), tr ? '0,00' : '0.00');
        expect(f.number(1234, decimalPlaces: 0), tr ? '1.234' : '1,234');
      });
      test('percent points and precision rules', () {
        expect(f.percent(2.75), tr ? '%2,75' : '2.75%');
        expect(f.percent(6, decimalPlaces: 0), tr ? '%6' : '6%');
        expect(f.percent(4.2, minimumDecimals: 1), tr ? '%4,2' : '4.2%');
        expect(f.percent(6.75, minimumDecimals: 1), tr ? '%6,75' : '6.75%');
        expect(f.number(2, decimalPlaces: 1, minimumDecimals: 0), '2');
      });
      test('decimal strings preserve precision and trailing zeros', () {
        expect(f.decimal('97846.00'), tr ? '97.846,00' : '97,846.00');
        expect(f.decimal('0.086450'), tr ? '0,086450' : '0.086450');
        expect(
          f.decimal('9007199254740993.010'),
          tr ? '9.007.199.254.740.993,010' : '9,007,199,254,740,993.010',
        );
        expect(f.decimal('5.200'), tr ? '5,200' : '5.200');
        expect(f.decimal('12', minimumDecimals: 2), tr ? '12,00' : '12.00');
        expect(f.decimal('12.0', trimZero: true), '12');
        expect(f.decimal('—'), '—');
      });
      test(
        'fixture copy localizes percentages and prices but preserves times',
        () {
          expect(
            f.fixtureText('2.75% → 3.25% of 6.00%'),
            tr ? '%2,75 → %3,25 of %6,00' : '2.75% → 3.25% of 6.00%',
          );
          expect(
            f.fixtureText('Closed at 31.93. Wait until 14:20.'),
            tr
                ? 'Closed at 31,93. Wait until 14:20.'
                : 'Closed at 31.93. Wait until 14:20.',
          );
          expect(
            f.fixtureText('Above 97,500.00'),
            tr ? 'Above 97.500,00' : 'Above 97,500.00',
          );
        },
      );
      test('nonfinite numeric values are rejected', () {
        expect(() => f.number(double.nan), throwsArgumentError);
        expect(() => f.percent(double.infinity), throwsArgumentError);
      });
    });
  }
}
