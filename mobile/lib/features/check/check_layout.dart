import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../l10n/number_formatter.dart';
import '../../data/models.dart';
import '../../data/providers.dart';
import '../../data/risk_level.dart';
import '../../l10n/app_strings.dart';
import '../../router.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/widgets.dart';
import 'check_strings.dart';
import 'order_measure.dart';
import 'projected_limit.dart';

/// One layout for both routes; approval always comes from the preview provider.
class CheckLayout extends ConsumerStatefulWidget {
  const CheckLayout({super.key, required this.preview, this.requestedPair});
  final OrderPreview preview;
  final String? requestedPair;
  @override
  ConsumerState<CheckLayout> createState() => _CheckLayoutState();
}

class _CheckLayoutState extends ConsumerState<CheckLayout> {
  String? _selection;
  void _discard() => context.goNamed(AppRoutes.home);
  void _feedback(String message) => ScaffoldMessenger.of(context)
    ..hideCurrentSnackBar()
    ..showSnackBar(SnackBar(content: Text(message)));

  @override
  Widget build(BuildContext context) {
    final pair = _selection ?? widget.requestedPair ?? widget.preview.pair;
    final caution = ref.watch(orderPreviewProvider);
    final blocked = ref.watch(blockedOrderPreviewProvider);
    final preview = pair == widget.preview.pair
        ? widget.preview
        : pair == caution.pair
        ? caution
        : pair == blocked.pair
        ? blocked
        : null;
    final account = ref.watch(accountProvider);
    final coin = ref.watch(coinPriceProvider(pair));
    final rule = ref
        .watch(coinClassRulesProvider)
        .where((rule) => rule.coinClass == coin?.coinClass)
        .firstOrNull;
    final time =
        '${account.cooldownEndsAt.hour.toString().padLeft(2, '0')}:'
        '${account.cooldownEndsAt.minute.toString().padLeft(2, '0')}';
    final c = context.colors;
    return WarRoomScaffold(
      appBar: AppBar(
        toolbarHeight: 56,
        leading: BackButton(
          onPressed: () {
            if (context.canPop()) {
              context.pop();
            } else {
              _discard();
            }
          },
        ),
        titleSpacing: 0,
        title: CommandTitle(
          AppStrings.of(context).check,
          style: const TextStyle(fontSize: 24),
        ),
        actions: [
          Container(
            margin: const EdgeInsets.only(right: 16),
            padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
            decoration: BoxDecoration(
              border: Border.all(color: c.line),
              borderRadius: BorderRadius.circular(MetronRadius.tag),
            ),
            child: Text(
              CheckStrings.of(context).paper,
              style: TextStyle(
                fontSize: 11,
                color: c.muted,
                letterSpacing: .66,
              ),
            ),
          ),
        ],
      ),
      body: SafeArea(
        bottom: false,
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(16, 4, 16, 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            spacing: 14,
            children: [
              Text(
                '${AppStrings.of(context).accountMarket}: ${preview?.market == AccountMarket.futures ? AppStrings.of(context).futures : AppStrings.of(context).spot}',
                key: const ValueKey('account-market'),
                style: TextStyle(fontSize: 12, color: c.muted),
              ),
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: _Field(
                      label: CheckStrings.of(context).pair,
                      child: DropdownButtonFormField<String>(
                        key: ValueKey(pair),
                        initialValue: pair,
                        isExpanded: true,
                        style: Theme.of(context).textTheme.bodyMedium,
                        decoration: const InputDecoration(
                          contentPadding: EdgeInsets.symmetric(horizontal: 10),
                        ),
                        items: {pair, caution.pair, blocked.pair}
                            .map(
                              (value) => DropdownMenuItem(
                                value: value,
                                child: Text(
                                  value,
                                  style: const TextStyle(fontSize: 14),
                                ),
                              ),
                            )
                            .toList(),
                        onChanged: (value) =>
                            setState(() => _selection = value),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _Field(
                      label: CheckStrings.of(context).price,
                      child: InputDecorator(
                        decoration: const InputDecoration(),
                        child: Semantics(
                          readOnly: true,
                          child: Num(
                            NumberFormatter.of(context)
                                .decimal(preview?.entry ?? coin?.price ?? '—'),
                            style: const TextStyle(fontSize: 16),
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              if (coin != null && rule != null)
                Wrap(
                  key: const ValueKey('check-meta'),
                  spacing: 12,
                  runSpacing: 6,
                  children: [
                    Text(switch (coin.coinClass) {
                      CoinClass.btcEth => AppStrings.of(context).coreClass,
                      CoinClass.largeCap => AppStrings.of(
                        context,
                      ).largeCapClass,
                      CoinClass.smallCap => AppStrings.of(
                        context,
                      ).smallCapClass,
                    }, style: TextStyle(fontSize: 12, color: c.subtle)),
                    for (final segment in [
                      '${AppStrings.of(context).dailyAtr} ${NumberFormatter.of(context).percent(coin.atrPercent, decimalPlaces: 1)}',
                      '${AppStrings.of(context).tradeRisk} ${NumberFormatter.of(context).percent(rule.riskPerTradePercent)}',
                      '${AppStrings.of(context).stopDistance} ${NumberFormatter.of(context).number(rule.atrMultiplier, decimalPlaces: 1)} ATR',
                    ])
                      Num(
                        segment,
                        style: TextStyle(fontSize: 12, color: c.subtle),
                      ),
                  ],
                ),
              if (preview == null)
                MetronCard(child: Text(CheckStrings.of(context).unavailable))
              else ...[
                _Verdict(
                  preview: preview,
                  remaining: account.cooldownEndsAt
                      .difference(account.asOf)
                      .inMinutes,
                ),
                OrderMeasure(preview: preview),
                MetronCard(
                  padding: EdgeInsets.zero,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Padding(
                        padding: EdgeInsets.all(16),
                        child: Text(
                          CheckStrings.of(context).limits,
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                      for (final check in preview.checks)
                        ProjectedLimit(check: check),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: c.surface,
          border: Border(top: BorderSide(color: c.line)),
        ),
        child: SafeArea(
          top: false,
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              spacing: 8,
              children: [
                if (preview != null) ...[
                  if (preview.verdict == RiskLevel.limit)
                    PrimaryButton(
                      label: preview.largestFittingOrderValue == null
                          ? AppStrings.of(context).waitUntilReset
                          : CheckStrings.of(context).reduce,
                      onPressed: preview.largestFittingOrderValue == null
                          ? null
                          : () => _showReduction(preview),
                    )
                  else if (preview.verdict == RiskLevel.caution) ...[
                    SizedBox(
                      height: MetronSize.control,
                      child: FilledButton(
                        onPressed: () =>
                            _feedback(CheckStrings.of(context).sampleReminder),
                        child: Wrap(
                          alignment: WrapAlignment.center,
                          children: [
                            Text(CheckStrings.of(context).wait),
                            Num(
                              time,
                              style: TextStyle(
                                color: c.inverse,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    SecondaryButton(
                      label: CheckStrings.of(context).anyway,
                      onPressed: () =>
                          _feedback(CheckStrings.of(context).sampleOrder),
                    ),
                    Text(
                      CheckStrings.of(context).violation,
                      textAlign: TextAlign.center,
                      style: TextStyle(fontSize: 12, color: c.subtle),
                    ),
                  ] else
                    PrimaryButton(
                      label: CheckStrings.of(context).place,
                      onPressed: () =>
                          _feedback(CheckStrings.of(context).sampleOrder),
                    ),
                ],
                SizedBox(
                  height: MetronSize.control,
                  child: TextButton(
                    onPressed: _discard,
                    child: Text(
                      CheckStrings.of(context).discard,
                      style: TextStyle(color: c.muted),
                    ),
                  ),
                ),
                if (preview?.verdict == RiskLevel.limit &&
                    preview?.largestFittingOrderValue != null)
                  _ReductionCopy(preview: preview!),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _showReduction(OrderPreview preview) {
    showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(CheckStrings.of(context).reductionTitle),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          spacing: 12,
          children: [
            _ReductionCopy(preview: preview),
            Text(CheckStrings.of(context).reductionBody),
          ],
        ),
        actions: [
          SecondaryButton(
            label: CheckStrings.of(context).close,
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
  }
}

class _Field extends StatelessWidget {
  const _Field({required this.label, required this.child});
  final String label;
  final Widget child;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    spacing: 6,
    children: [
      Text(label, style: TextStyle(fontSize: 12, color: context.colors.muted)),
      child,
    ],
  );
}

class _Verdict extends StatelessWidget {
  const _Verdict({required this.preview, required this.remaining});
  final OrderPreview preview;
  final int remaining;
  @override
  Widget build(BuildContext context) {
    final level = preview.verdict;
    final body = switch (level) {
      RiskLevel.safe => CheckStrings.of(context).safeBody,
      RiskLevel.caution => CheckStrings.of(context).cooldown(
        NumberFormatter.of(context)
            .number(remaining.clamp(0, 60), decimalPlaces: 0),
      ),
      RiskLevel.limit => AppStrings.of(context).dailyLossBlockedBody,
    };
    return Semantics(
      liveRegion: true,
      child: Container(
        padding: EdgeInsets.all(level == RiskLevel.limit ? 20 : 16),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [level.tint(context.colors), context.colors.surface],
          ),
          border: Border(
            left: BorderSide(color: level.color(context.colors), width: 5),
            top: BorderSide(color: level.color(context.colors)),
            right: BorderSide(color: level.color(context.colors)),
            bottom: BorderSide(color: level.color(context.colors)),
          ),
          boxShadow: level == RiskLevel.limit
              ? [BoxShadow(color: context.colors.limitTint, blurRadius: 32)]
              : null,
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(
              level.icon,
              size: 28,
              color: level.color(context.colors),
              semanticLabel: level.label(context),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: 4,
                children: [
                  CommandTitle(
                    switch (level) {
                      RiskLevel.safe => CheckStrings.of(context).safeTitle,
                      RiskLevel.caution => CheckStrings.of(
                        context,
                      ).cautionTitle,
                      RiskLevel.limit => AppStrings.of(
                        context,
                      ).dailyLossBlockedTitle,
                    },
                    style: TextStyle(
                      fontSize: level == RiskLevel.limit ? 30 : 26,
                      color: level.color(context.colors),
                    ),
                  ),
                  NumericCopy(body),
                  if (level == RiskLevel.limit)
                    Align(
                      alignment: Alignment.centerRight,
                      child: Padding(
                        padding: const EdgeInsets.only(top: 16, bottom: 8),
                        child: Transform.rotate(
                          angle: -.12,
                          child: Container(
                            key: const ValueKey('blocked-stamp'),
                            padding: const EdgeInsets.all(4),
                            decoration: BoxDecoration(
                              border: Border.all(
                                color: context.colors.limit,
                                width: 2,
                              ),
                            ),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 10,
                                vertical: 3,
                              ),
                              decoration: BoxDecoration(
                                border: Border.all(color: context.colors.limit),
                              ),
                              child: CommandTitle(
                                AppStrings.of(context).blockedStamp,
                                style: TextStyle(
                                  fontSize: 26,
                                  color: context.colors.limit,
                                  letterSpacing: 2,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Mixed copy keeps numeric runs in Num and words in the body typeface.
class NumericCopy extends StatelessWidget {
  const NumericCopy(this.text, {super.key, this.small = false});
  final String text;
  final bool small;
  @override
  Widget build(BuildContext context) => Semantics(
    label: text,
    excludeSemantics: true,
    child: Wrap(
      alignment: small ? WrapAlignment.center : WrapAlignment.start,
      spacing: 3,
      children: [
        for (final word in text.split(' '))
          if (RegExp(r'\d').hasMatch(word))
            Num(
              word,
              style: TextStyle(
                fontSize: small ? 12 : 14,
                color: context.colors.muted,
              ),
            )
          else
            Text(
              word,
              style: TextStyle(
                fontSize: small ? 12 : 14,
                height: small ? 17 / 12 : 20 / 14,
                color: context.colors.muted,
              ),
            ),
      ],
    ),
  );
}

class _ReductionCopy extends StatelessWidget {
  const _ReductionCopy({required this.preview});
  final OrderPreview preview;
  @override
  Widget build(BuildContext context) => NumericCopy(
    CheckStrings.of(context).reduction(
      NumberFormatter.of(context)
          .decimal(preview.largestFittingOrderValue ?? '—'),
      NumberFormatter.of(context).decimal(preview.largestFittingRisk ?? '—'),
    ),
    small: true,
  );
}
