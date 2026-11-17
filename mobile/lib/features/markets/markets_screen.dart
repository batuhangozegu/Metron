import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../l10n/number_formatter.dart';
import '../../data/models.dart';
import '../../data/providers.dart';
import '../../l10n/app_strings.dart';
import '../../router.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/widgets.dart';
import 'markets_strings.dart';

class MarketsScreen extends ConsumerStatefulWidget {
  const MarketsScreen({super.key});
  @override
  ConsumerState<MarketsScreen> createState() => _MarketsScreenState();
}

class _MarketsScreenState extends ConsumerState<MarketsScreen> {
  String _query = '';
  CoinClass? _class;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final prices = ref
        .watch(pricesProvider)
        .where(
          (coin) =>
              (_class == null || coin.coinClass == _class) &&
              coin.pair.toLowerCase().contains(_query.trim().toLowerCase()),
        )
        .toList();
    final note = TextStyle(fontSize: 12, height: 17 / 12, color: c.subtle);
    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    AppStrings.of(context).markets,
                    style: MetronType.heading.copyWith(
                      fontSize: 22,
                      height: 28 / 22,
                      letterSpacing: -.22,
                    ),
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 7,
                        height: 7,
                        decoration: BoxDecoration(
                          color: c.safe,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 6),
                      Text(
                        MarketsStrings.of(context).live,
                        style: TextStyle(fontSize: 12, color: c.muted),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                keyboardDismissBehavior:
                    ScrollViewKeyboardDismissBehavior.onDrag,
                padding: const EdgeInsets.fromLTRB(16, 4, 16, 20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Semantics(
                      label: MarketsStrings.of(context).search,
                      child: TextField(
                        onChanged: (value) => setState(() => _query = value),
                        textInputAction: TextInputAction.search,
                        decoration: InputDecoration(
                          hintText: MarketsStrings.of(context).search,
                          contentPadding: EdgeInsets.symmetric(
                            horizontal: 14,
                            vertical: 14,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    Semantics(
                      label: MarketsStrings.of(context).coinClass,
                      child: SingleChildScrollView(
                        scrollDirection: Axis.horizontal,
                        child: Row(
                          children: [
                            for (final value in [
                              null,
                              ...CoinClass.values,
                            ]) ...[
                              if (value != null) const SizedBox(width: 6),
                              _ClassFilter(
                                label: value == null
                                    ? MarketsStrings.of(context).all
                                    : MarketsStrings.of(context)
                                          .className(value),
                                selected: value == _class,
                                onPressed: () => setState(() => _class = value),
                              ),
                            ],
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    MetronCard(
                      padding: EdgeInsets.zero,
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(MetronRadius.card),
                        child: prices.isEmpty
                            ? Padding(
                                padding: const EdgeInsets.all(24),
                                child: Text(
                                  MarketsStrings.of(context).noResults,
                                  style: note,
                                ),
                              )
                            : Column(
                                children: [
                                  for (var i = 0; i < prices.length; i++) ...[
                                    if (i > 0) const Divider(height: 1),
                                    _PriceRow(coin: prices[i]),
                                  ],
                                ],
                              ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text.rich(
                      TextSpan(
                        children: [
                          TextSpan(text: MarketsStrings.of(context).barStart),
                          WidgetSpan(
                            child: Num(
                              NumberFormatter.of(context)
                                  .number(0, decimalPlaces: 0),
                              style: note,
                            ),
                          ),
                          TextSpan(text: MarketsStrings.of(context).barTo),
                          WidgetSpan(
                            child: Num(
                              NumberFormatter.of(context)
                                  .percent(10, decimalPlaces: 0),
                              style: note,
                            ),
                          ),
                          TextSpan(text: MarketsStrings.of(context).barEnd),
                        ],
                      ),
                      style: note,
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ClassFilter extends StatelessWidget {
  const _ClassFilter({
    required this.label,
    required this.selected,
    required this.onPressed,
  });
  final String label;
  final bool selected;
  final VoidCallback onPressed;
  @override
  Widget build(BuildContext context) => Semantics(
    selected: selected,
    child: SizedBox(
      height: MetronSize.control,
      child: OutlinedButton(
        onPressed: onPressed,
        style: OutlinedButton.styleFrom(
          backgroundColor: selected ? context.colors.text : context.colors.bg,
          foregroundColor: selected
              ? context.colors.inverse
              : context.colors.muted,
          side: BorderSide(
            color: selected ? context.colors.text : context.colors.line,
          ),
          padding: const EdgeInsets.symmetric(horizontal: 14),
          textStyle: MetronType.small.copyWith(fontWeight: FontWeight.w500),
          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
        ),
        child: Text(label),
      ),
    ),
  );
}

class _PriceRow extends StatelessWidget {
  const _PriceRow({required this.coin});
  final CoinPrice coin;
  @override
  Widget build(BuildContext context) {
    final subtle = TextStyle(
      fontSize: 12,
      height: 16 / 12,
      color: context.colors.subtle,
    );
    return InkWell(
      onTap: () => context.pushNamed(
        AppRoutes.check,
        queryParameters: {'pair': coin.pair},
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    coin.pair,
                    style: MetronType.body.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Wrap(
                    children: [
                      Text(
                        '${MarketsStrings.of(context).className(coin.coinClass)} · ${MarketsStrings.of(context).stop} ',
                        style: subtle,
                      ),
                      Num(
                        NumberFormatter.of(context).percent(
                          coin.suggestedStopPercent,
                          minimumDecimals: 1,
                        ),
                        style: subtle,
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            SizedBox(
              width: 84,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(2),
                    child: LinearProgressIndicator(
                      value: (coin.atrPercent / 10).clamp(0, 1),
                      minHeight: 5,
                      color: context.colors.muted,
                      backgroundColor: context.colors.raised,
                      semanticsLabel: MarketsStrings.of(context).atr,
                      semanticsValue: NumberFormatter.of(context)
                          .percent(coin.atrPercent, decimalPlaces: 1),
                    ),
                  ),
                  const SizedBox(height: 4),
                  Num(
                    '${MarketsStrings.of(context).atr} ${NumberFormatter.of(context).percent(coin.atrPercent, decimalPlaces: 1)}',
                    style: subtle.copyWith(fontSize: 11),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Num(
                  NumberFormatter.of(context).decimal(coin.price),
                  style: const TextStyle(fontSize: 15),
                ),
                const SizedBox(height: 3),
                PnlText(
                  coin.changePercent,
                  suffix: '%',
                  style: const TextStyle(fontSize: 12, height: 16 / 12),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
