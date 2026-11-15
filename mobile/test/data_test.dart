import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/data/providers.dart';
import 'package:metron/data/risk_level.dart';

void main() {
  test('providers expose the shared mockup account and immutable rows', () {
    final container = ProviderContainer();
    addTearDown(container.dispose);
    expect(container.read(accountProvider).equity, '25000.00');
    expect(container.read(accountProvider).asOf.isUtc, isTrue);
    expect(container.read(positionsProvider).map((p) => p.id), [
      'btc',
      'sol',
      'eth',
    ]);
    expect(container.read(positionProvider('btc'))!.detail!.stop, '93508.00');
    expect(container.read(positionProvider('missing')), isNull);
    expect(container.read(coinPriceProvider('PENDLE/USDT'))!.atrPercent, 8);
    expect(container.read(coinPriceProvider('missing')), isNull);
    expect(container.read(pricesProvider), hasLength(9));
    expect(container.read(alertsProvider), hasLength(5));
    expect(container.read(priceAlarmsProvider), hasLength(2));
    expect(container.read(coinClassRulesProvider), hasLength(3));
    expect(container.read(disciplineRulesProvider), hasLength(3));
    expect(
      () => container.read(positionsProvider).clear(),
      throwsUnsupportedError,
    );
    expect(
      () => container.read(pricesProvider).clear(),
      throwsUnsupportedError,
    );
  });

  test('caution and blocked previews retain independent explicit verdicts', () {
    final container = ProviderContainer();
    addTearDown(container.dispose);
    final caution = container.read(orderPreviewProvider);
    final blocked = container.read(blockedOrderPreviewProvider);
    expect(caution.verdict, RiskLevel.caution);
    expect(caution.checks.first.brokenRule, isTrue);
    expect(caution.checks.first.level, RiskLevel.caution);
    expect(blocked.verdict, RiskLevel.limit);
    expect(blocked.checks.first.after, 66.67);
    expect(blocked.largestFittingOrderValue, '1666.67');
  });
}
