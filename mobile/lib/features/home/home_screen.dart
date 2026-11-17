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
import 'home_strings.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final account = ref.watch(accountProvider);
    final limits = ref.watch(limitsProvider);
    final positions = ref.watch(positionsProvider);
    final level = [
      RiskLevel.forUsage(account.openRiskPercent, limits.openRiskPercent),
      RiskLevel.forUsage(account.dailyLossPercent, limits.dailyLossPercent),
      RiskLevel.forUsage(account.weeklyLossPercent, limits.weeklyLossPercent),
    ].reduce((a, b) => a.index > b.index ? a : b);
    final c = context.colors;
    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            spacing: 16,
            children: [
              Row(
                children: [
                  const Icon(Icons.stacked_bar_chart, size: 20),
                  const SizedBox(width: 8),
                  Text(
                    AppStrings.of(context).appName,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const Spacer(),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 7,
                      vertical: 3,
                    ),
                    decoration: BoxDecoration(
                      border: Border.all(color: c.line),
                      borderRadius: BorderRadius.circular(MetronRadius.tag),
                    ),
                    child: Text(
                      HomeStrings.of(context).paper,
                      style: TextStyle(
                        fontSize: 11,
                        color: c.muted,
                        letterSpacing: .66,
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Num(
                    NumberFormatter.of(context)
                        .decimal(account.equity, minimumDecimals: 2),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: 8,
                children: [
                  SectionHeader(
                    HomeStrings.of(context).riskStatus,
                    trailing: Num(
                      '${_time(account.asOf)} UTC',
                      style: TextStyle(fontSize: 11, color: c.subtle),
                    ),
                  ),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Semantics(
                        label: level.label(context),
                        child: Icon(
                          level.icon,
                          color: level.color(c),
                          size: 26,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          HomeStrings.of(context).headline(level),
                          style: Theme.of(context).textTheme.headlineLarge,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              _RiskCard(
                title: HomeStrings.of(context).openRisk,
                value: account.openRiskPercent,
                limit: limits.openRiskPercent,
                large: true,
                footer: Wrap(
                  spacing: 4,
                  children: [
                    Num(
                      '${NumberFormatter.of(context).decimal(account.moneyAtRisk, minimumDecimals: 2)} USDT',
                      style: const TextStyle(fontSize: 13),
                    ),
                    Text(
                      HomeStrings.of(context).stopsTrigger,
                      style: TextStyle(fontSize: 13, color: c.muted),
                    ),
                  ],
                ),
              ),
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: _RiskCard(
                      title: HomeStrings.of(context).dailyCap,
                      value: account.dailyLossPercent,
                      limit: limits.dailyLossPercent,
                      footer: _Remaining(
                        account.dailyRemaining,
                        HomeStrings.of(context).today,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _RiskCard(
                      title: HomeStrings.of(context).weeklyCap,
                      value: account.weeklyLossPercent,
                      limit: limits.weeklyLossPercent,
                      footer: _Remaining(
                        account.weeklyRemaining,
                        HomeStrings.of(context).thisWeek,
                      ),
                    ),
                  ),
                ],
              ),
              if (account.cooldownEndsAt.isAfter(account.asOf))
                _Cooldown(account: account),
              Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                spacing: 8,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          HomeStrings.of(context).positions,
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                      ),
                      PnlText(
                        double.parse(account.unrealizedPnl),
                        style: const TextStyle(fontSize: 13),
                      ),
                    ],
                  ),
                  MetronCard(
                    padding: EdgeInsets.zero,
                    child: positions.isEmpty
                        ? Padding(
                            padding: EdgeInsets.all(16),
                            child: Text(HomeStrings.of(context).noPositions),
                          )
                        : Column(
                            children: [
                              for (var i = 0; i < positions.length; i++) ...[
                                if (i > 0) const Divider(),
                                _PositionRow(position: positions[i]),
                              ],
                            ],
                          ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Reuses the shared ruler painter with the home mockup's figure hierarchy.
class _RiskCard extends StatelessWidget {
  const _RiskCard({
    required this.title,
    required this.value,
    required this.limit,
    required this.footer,
    this.large = false,
  });
  final String title;
  final double value, limit;
  final Widget footer;
  final bool large;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final level = RiskLevel.forUsage(value, limit);
    return MetronCard(
      padding: EdgeInsets.all(large ? 20 : 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: large ? 14 : 10,
        children: [
          if (large)
            Row(
              children: [
                Expanded(
                  child: Text(
                    title,
                    style: TextStyle(fontSize: 14, color: c.muted),
                  ),
                ),
                RiskBadge(level),
              ],
            )
          else ...[
            Text(title, style: TextStyle(fontSize: 13, color: c.muted)),
            Align(alignment: Alignment.centerLeft, child: RiskBadge(level)),
          ],
          Wrap(
            spacing: 8,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              Num(
                NumberFormatter.of(context).percent(value),
                style: large
                    ? MetronType.figureLarge
                    : const TextStyle(
                        fontSize: 22,
                        height: 28 / 22,
                        fontWeight: FontWeight.w500,
                      ),
              ),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    '${AppStrings.of(context).outOf} ',
                    style: TextStyle(
                      fontSize: large ? 14 : 12,
                      color: c.subtle,
                    ),
                  ),
                  Num(
                    NumberFormatter.of(context)
                        .percent(limit, decimalPlaces: large ? 2 : 0),
                    style: TextStyle(
                      fontSize: large ? 14 : 12,
                      color: c.subtle,
                    ),
                  ),
                ],
              ),
            ],
          ),
          Semantics(
            label:
                '$title, ${NumberFormatter.of(context).percent(value)} ${AppStrings.of(context).outOf} ${NumberFormatter.of(context).percent(limit, decimalPlaces: large ? 2 : 0)}, ${level.label(context)}',
            child: SizedBox(
              height: 30,
              child: CustomPaint(
                painter: MeasurePainter(
                  fraction: (value / limit).clamp(0, 1),
                  fill: level.color(c),
                  track: c.raised,
                  tick: c.subtle,
                  caution: c.caution,
                  limit: c.limit,
                ),
              ),
            ),
          ),
          footer,
        ],
      ),
    );
  }
}

class _Remaining extends StatelessWidget {
  const _Remaining(this.amount, this.period);
  final String amount, period;
  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 4,
    children: [
      Num(
        NumberFormatter.of(context).decimal(amount, minimumDecimals: 2),
        style: TextStyle(fontSize: 12, color: context.colors.subtle),
      ),
      Text(
        period,
        style: TextStyle(fontSize: 12, color: context.colors.subtle),
      ),
    ],
  );
}

class _Cooldown extends StatelessWidget {
  const _Cooldown({required this.account});
  final PaperAccount account;
  @override
  Widget build(BuildContext context) => Material(
    color: context.colors.cautionTint,
    borderRadius: BorderRadius.circular(MetronRadius.card),
    child: InkWell(
      borderRadius: BorderRadius.circular(MetronRadius.card),
      onTap: () => context.goNamed(AppRoutes.alerts),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          spacing: 6,
          children: [
            const RiskBadge(RiskLevel.caution),
            Text(
              HomeStrings.of(context).cooldown,
              style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
            ),
            Text.rich(
              TextSpan(
                children: [
                  TextSpan(text: HomeStrings.of(context).waitUntil),
                  WidgetSpan(
                    alignment: PlaceholderAlignment.baseline,
                    baseline: TextBaseline.alphabetic,
                    child: Num(
                      _time(account.cooldownEndsAt),
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                  const TextSpan(text: ', '),
                  WidgetSpan(
                    alignment: PlaceholderAlignment.baseline,
                    baseline: TextBaseline.alphabetic,
                    child: Num(
                      NumberFormatter.of(context).number(
                        account.cooldownEndsAt
                            .difference(account.asOf)
                            .inMinutes,
                        decimalPlaces: 0,
                      ),
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                  TextSpan(text: HomeStrings.of(context).minutesFromNow),
                ],
              ),
              style: TextStyle(
                fontSize: 13,
                height: 19 / 13,
                color: context.colors.muted,
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class _PositionRow extends StatelessWidget {
  const _PositionRow({required this.position});
  final Position position;
  @override
  Widget build(BuildContext context) => InkWell(
    onTap: () => context.pushNamed(
      AppRoutes.position,
      pathParameters: {'id': position.id},
    ),
    child: ConstrainedBox(
      constraints: const BoxConstraints(minHeight: MetronSize.control),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: 3,
                children: [
                  Text(
                    position.pair,
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  Wrap(
                    spacing: 4,
                    children: [
                      Text(
                        '${HomeStrings.of(context).coinClass(position.coinClass)} ·',
                        style: TextStyle(
                          fontSize: 12,
                          color: context.colors.subtle,
                        ),
                      ),
                      Num(
                        NumberFormatter.of(context)
                            .percent(position.distanceToStopPercent),
                        style: TextStyle(
                          fontSize: 12,
                          color: context.colors.subtle,
                        ),
                      ),
                      Text(
                        HomeStrings.of(context).toStop,
                        style: TextStyle(
                          fontSize: 12,
                          color: context.colors.subtle,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              spacing: 3,
              children: [
                PnlText(
                  double.parse(position.unrealizedPnl),
                  style: const TextStyle(fontSize: 15),
                ),
                Num(
                  '${NumberFormatter.of(context).decimal(position.moneyAtRisk, minimumDecimals: 2)} ${HomeStrings.of(context).atRisk}',
                  style: TextStyle(fontSize: 12, color: context.colors.subtle),
                ),
              ],
            ),
          ],
        ),
      ),
    ),
  );
}

String _time(DateTime date) {
  final utc = date.toUtc();
  return '${utc.hour.toString().padLeft(2, '0')}:${utc.minute.toString().padLeft(2, '0')}';
}
