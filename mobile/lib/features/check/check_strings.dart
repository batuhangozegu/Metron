/// Feature-local copy, ready for localization.
abstract final class CheckStrings {
  static const paper = 'PAPER';
  static const limit = 'Limit';
  static const wait = 'Wait, remind me at ';
  static const safeTitle = 'Within your limits';
  static const cautionTitle = 'One rule would be broken';
  static const blockedTitle = 'Blocked: position too large';
  static const safeBody = 'Every check is within your limits.';
  static const broken = 'Broken';
  static const exceeded = 'Exceeded';
  static String cooldown(int minutes) =>
      'No limit is exceeded. Your cooldown rule asks you '
      'to wait $minutes more minutes.';
  static String blocked(String coin, String after, String cap) =>
      '$coin would be $after% of the account. Your '
      'BTC/ETH rule allows $cap%.';
  static String reduction(String value, String risk) =>
      'The largest order that fits is $value USDT, risking '
      '$risk USDT. A blocked order cannot be placed anyway.';
  static const pair = 'Pair';
  static const price = 'Limit price, USDT';
  static const limits = 'Your limits after this order';
  static const stop = 'Stop';
  static const entry = 'Entry';
  static const target = 'Target';
  static const stopResult = 'If the stop triggers';
  static const targetResult = 'If the target fills';
  static const size = 'Position size';
  static const value = 'Order value';
  static const discard = 'Discard';
  static const reduce = 'Reduce to fit the limit';
  static const place = 'Place paper order';
  static const anyway = 'Place paper order anyway';
  static const violation = 'Placing it now is recorded as a broken rule.';
  static const unavailable = 'No order preview is available for this pair.';
  static const sampleOrder = 'Sample preview only. No order was placed.';
  static const sampleReminder =
      'Sample preview only. No reminder was scheduled.';
  static const reductionTitle = 'Suggested reduction';
  static const reductionBody =
      'Recheck the reduced order before placing it. '
      'This sample cannot recalculate limits or clear the cooldown.';
  static const close = 'Close';
}
