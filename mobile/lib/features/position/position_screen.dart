import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/providers.dart';
import '../../data/risk_level.dart';
import '../../l10n/app_strings.dart';
import '../../router.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/widgets.dart';
import 'position_format.dart';
import 'position_strings.dart';
import 'price_ruler.dart';

class PositionScreen extends ConsumerWidget {
  const PositionScreen({super.key, required this.positionId});
  final String positionId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final position = ref.watch(positionProvider(positionId));
    final detail = position?.detail;
    final price = position == null
        ? null
        : ref.watch(coinPriceProvider(position.pair));
    final rules = ref.watch(coinClassRulesProvider);
    final maxShare = rules
        .where((r) => r.coinClass == position?.coinClass)
        .firstOrNull
        ?.maxPositionPercent;
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          tooltip: AppStrings.of(context).back,
          onPressed: () {
            if (Navigator.of(context).canPop()) {
              context.pop();
            } else {
              context.goNamed(AppRoutes.home);
            }
          },
          icon: const Icon(Icons.chevron_left, size: 22),
        ),
        titleSpacing: 0,
        title: Text(position?.pair ?? AppStrings.of(context).position),
        actions: [
          if (position != null)
            Container(
              margin: const EdgeInsets.only(right: 16),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                border: Border.all(color: context.colors.line),
                borderRadius: BorderRadius.circular(MetronRadius.tag),
              ),
              child: Text(
                PositionStrings.of(context).coinClass(position.coinClass),
                style: TextStyle(fontSize: 12, color: context.colors.muted),
              ),
            ),
        ],
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(16, 4, 16, 20),
          child: position == null
              ? Text(PositionStrings.of(context).notFound)
              : Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  spacing: 14,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      spacing: 4,
                      children: [
                        Text(
                          PositionStrings.of(context).unrealized,
                          style: TextStyle(
                            fontSize: 12,
                            color: context.colors.subtle,
                          ),
                        ),
                        PnlText(
                          double.parse(position.unrealizedPnl),
                          format: (v) => formatDecimal(v.toStringAsFixed(2)),
                          style: MetronType.figureLarge.copyWith(
                            fontSize: 40,
                            height: 46 / 40,
                            letterSpacing: -1.2,
                          ),
                        ),
                        Wrap(
                          spacing: 6,
                          runSpacing: 4,
                          crossAxisAlignment: WrapCrossAlignment.center,
                          children: [
                            if (detail != null) ...[
                              PnlText(detail.returnPercent, suffix: '%'),
                              const Text('·'),
                              PnlText(
                                detail.returnR,
                                decimalPlaces: 1,
                                suffix: 'R',
                              ),
                              const Text('·'),
                            ],
                            Text(
                              PositionStrings.of(context).price,
                              style: TextStyle(color: context.colors.muted),
                            ),
                            Num(
                              price == null ? '—' : formatDecimal(price.price),
                              style: TextStyle(color: context.colors.muted),
                            ),
                          ],
                        ),
                      ],
                    ),
                    if (detail != null) ...[
                      MetronCard(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 18,
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          spacing: 14,
                          children: [
                            PriceRuler(detail: detail, price: price?.price),
                            const Divider(),
                            Wrap(
                              spacing: 8,
                              runSpacing: 8,
                              crossAxisAlignment: WrapCrossAlignment.center,
                              children: [
                                RiskBadge(
                                  detail.stopOnExchange
                                      ? RiskLevel.safe
                                      : RiskLevel.limit,
                                ),
                                Text(
                                  detail.stopOnExchange
                                      ? PositionStrings.of(context)
                                            .stopOnExchange
                                      : PositionStrings.of(context).stopMissing,
                                  style: TextStyle(
                                    fontSize: 13,
                                    color: context.colors.muted,
                                  ),
                                ),
                                if (detail.stopOnExchange)
                                  Num(
                                    '${position.distanceToStopPercent.toStringAsFixed(2)}% ${PositionStrings.of(context).away}',
                                  ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      MetronCard(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          spacing: 12,
                          children: [
                            _DetailRow(
                              PositionStrings.of(context).quantity,
                              '${detail.quantity} ${position.pair.split('/').first}',
                            ),
                            _DetailRow(
                              PositionStrings.of(context).valueNow,
                              '${formatDecimal(detail.valueNow)} ${position.pair.split('/').last}',
                            ),
                            _DetailRow(
                              PositionStrings.of(context).moneyAtRisk,
                              '${formatDecimal(position.moneyAtRisk)} ${position.pair.split('/').last}',
                            ),
                            _DetailRow(
                              PositionStrings.of(context).share,
                              '${detail.shareOfAccountPercent.toStringAsFixed(2)}%${maxShare == null ? '' : ' of ${formatDecimal(maxShare.toString())}%'}',
                            ),
                            _DetailRow(
                              PositionStrings.of(context).opened,
                              PositionStrings.of(context)
                                  .openedAt(detail.openedAt),
                            ),
                          ],
                        ),
                      ),
                      Text(
                        PositionStrings.of(context).stopRule,
                        style: TextStyle(
                          fontSize: 12,
                          height: 17 / 12,
                          color: context.colors.subtle,
                        ),
                      ),
                    ] else
                      MetronCard(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          spacing: 12,
                          children: [
                            _DetailRow(
                              PositionStrings.of(context).moneyAtRisk,
                              '${formatDecimal(position.moneyAtRisk)} ${position.pair.split('/').last}',
                            ),
                            Text(PositionStrings.of(context).detailsMissing),
                          ],
                        ),
                      ),
                  ],
                ),
        ),
      ),
      bottomNavigationBar: detail == null
          ? null
          : DecoratedBox(
              decoration: BoxDecoration(
                color: context.colors.surface,
                border: Border(top: BorderSide(color: context.colors.line)),
              ),
              child: SafeArea(
                top: false,
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                  child: Row(
                    spacing: 10,
                    children: [
                      Expanded(
                        child: SecondaryButton(
                          label: PositionStrings.of(context).tighten,
                          onPressed: () => _showAction(
                            context,
                            PositionStrings.of(context).tighten,
                          ),
                        ),
                      ),
                      Expanded(
                        child: PrimaryButton(
                          label: PositionStrings.of(context).close,
                          onPressed: () => _showAction(
                            context,
                            PositionStrings.of(context).close,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
    );
  }

  // The foundation exposes read-only fixtures, with no order mutation API.
  void _showAction(BuildContext context, String title) {
    showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: Text(
          title == PositionStrings.of(context).tighten
              ? '${PositionStrings.of(context).stopRule}\n\n${PositionStrings.of(context).previewOnly}'
              : PositionStrings.of(context).previewOnly,
        ),
        actions: [
          SecondaryButton(
            label: PositionStrings.of(context).dismiss,
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  const _DetailRow(this.label, this.value);
  final String label, value;

  @override
  Widget build(BuildContext context) => Row(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Expanded(
        child: Text(
          label,
          style: TextStyle(fontSize: 14, color: context.colors.muted),
        ),
      ),
      const SizedBox(width: 12),
      Flexible(flex: 2, child: Num(value, textAlign: TextAlign.right)),
    ],
  );
}
