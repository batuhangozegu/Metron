/// Decimal source amounts never become doubles. Only the bounded pixel ratio
/// is converted after integer division, at four decimal places of resolution.
BigInt decimalTicks(String value) {
  if (!RegExp(r'^-?\d+(\.\d{1,28})?$').hasMatch(value)) return BigInt.zero;
  final parts = value.split('.');
  return BigInt.parse(
    parts[0] + (parts.length == 1 ? '' : parts[1]).padRight(28, '0'),
  );
}

double visualFraction(BigInt numerator, BigInt denominator) {
  if (denominator <= BigInt.zero || numerator <= BigInt.zero) return 0;
  if (numerator >= denominator) return 1;
  return (numerator * BigInt.from(10000) ~/ denominator).toInt() / 10000;
}
