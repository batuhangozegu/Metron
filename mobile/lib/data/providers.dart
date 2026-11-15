import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'models.dart';
import 'sample_data.dart';

/// Read-only, overrideable fixtures. No networking or generated provider code.
final accountProvider = Provider<PaperAccount>((ref) => SampleData.account);
final limitsProvider = Provider<AccountLimits>((ref) => SampleData.limits);
final positionsProvider = Provider<List<Position>>(
  (ref) => SampleData.positions,
);
final pricesProvider = Provider<List<CoinPrice>>((ref) => SampleData.prices);
final alertsProvider = Provider<List<AlertEvent>>((ref) => SampleData.alerts);
final priceAlarmsProvider = Provider<List<PriceAlarm>>(
  (ref) => SampleData.priceAlarms,
);
final coinClassRulesProvider = Provider<List<CoinClassRule>>(
  (ref) => SampleData.coinClassRules,
);
final disciplineRulesProvider = Provider<List<DisciplineRule>>(
  (ref) => SampleData.disciplineRules,
);
final orderPreviewProvider = Provider<OrderPreview>(
  (ref) => SampleData.orderPreview,
);
final blockedOrderPreviewProvider = Provider<OrderPreview>(
  (ref) => SampleData.blockedOrderPreview,
);

/// Unknown IDs resolve to null so a detail screen can show a missing state.
final positionProvider = Provider.family<Position?, String>((ref, id) {
  for (final position in ref.watch(positionsProvider)) {
    if (position.id == id) return position;
  }
  return null;
});

final coinPriceProvider = Provider.family<CoinPrice?, String>((ref, pair) {
  for (final coin in ref.watch(pricesProvider)) {
    if (coin.pair == pair) return coin;
  }
  return null;
});
