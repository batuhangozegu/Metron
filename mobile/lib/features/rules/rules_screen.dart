import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/models.dart';
import '../../data/providers.dart';
import '../../data/risk_level.dart';
import '../../l10n/app_strings.dart';
import '../../router.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/widgets.dart';
import 'rules_strings.dart';

/// Read-only rules; the shell supplies persistent bottom navigation.
class RulesScreen extends ConsumerWidget {
  const RulesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final limits = ref.watch(limitsProvider);
    final classes = ref.watch(coinClassRulesProvider);
    final discipline = ref.watch(disciplineRulesProvider);
    final account = ref.watch(accountProvider);
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          key: const PageStorageKey('rules-scroll'),
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                AppStrings.rules,
                style: Theme.of(context).textTheme.titleLarge!.copyWith(
                  fontSize: 22,
                  height: 28 / 22,
                  letterSpacing: -.22,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                RulesStrings.introduction,
                style: Theme.of(context).textTheme.bodySmall!
                    .copyWith(color: context.colors.muted),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.accountLimits,
                child: MetronCard(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    spacing: 12,
                    children: [
                      _ValueRow(
                        RulesStrings.openRisk,
                        Num(_percent(limits.openRiskPercent)),
                      ),
                      _ValueRow(
                        RulesStrings.dailyCap,
                        Num(_percent(limits.dailyLossPercent)),
                      ),
                      _ValueRow(
                        RulesStrings.weeklyCap,
                        Num(_percent(limits.weeklyLossPercent)),
                      ),
                      _ValueRow(
                        RulesStrings.positions,
                        Num('${limits.maxOpenPositions}'),
                      ),
                      _ValueRow(
                        RulesStrings.riskReward,
                        Num(limits.targetRiskReward.toStringAsFixed(1)),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.coinClasses,
                child: _RowsCard(
                  children: [
                    for (final rule in classes) _ClassRow(rule),
                    if (classes.isEmpty)
                      const _EmptyRow(RulesStrings.noClasses),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.discipline,
                child: _RowsCard(
                  children: [
                    for (final rule in discipline) _DisciplineRow(rule),
                    if (discipline.isEmpty)
                      const _EmptyRow(RulesStrings.noDiscipline),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.account,
                child: _RowsCard(
                  children: [
                    Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      child: _ValueRow(
                        RulesStrings.tradingAccount,
                        Wrap(
                          alignment: WrapAlignment.end,
                          spacing: 4,
                          children: [
                            const Text(RulesStrings.paper),
                            Num('${_groupDecimal(account.equity)} USDT'),
                          ],
                        ),
                      ),
                    ),
                    Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      child: _ValueRow(
                        RulesStrings.twoFactor,
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              account.twoFactorEnabled
                                  ? Icons.check_circle_outline
                                  : Icons.remove_circle_outline,
                              size: 16,
                              color: account.twoFactorEnabled
                                  ? context.colors.safe
                                  : context.colors.muted,
                            ),
                            const SizedBox(width: 4),
                            Text(
                              account.twoFactorEnabled
                                  ? RulesStrings.on
                                  : RulesStrings.off,
                              style: TextStyle(
                                color: account.twoFactorEnabled
                                    ? context.colors.safe
                                    : context.colors.muted,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    // No auth service exists in the fixture foundation yet.
                    SizedBox(
                      height: MetronSize.control,
                      child: TextButton(
                        style: TextButton.styleFrom(
                          alignment: Alignment.centerLeft,
                          padding: const EdgeInsets.symmetric(horizontal: 16),
                          foregroundColor: context.colors.text,
                          textStyle: Theme.of(context).textTheme.bodyMedium!
                              .copyWith(
                                fontSize: 14,
                                fontWeight: FontWeight.w500,
                              ),
                        ),
                        onPressed: () => context.goNamed(AppRoutes.signIn),
                        child: const Text(RulesStrings.signOut),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

String _percent(double value) => '${value.toStringAsFixed(2)}%';

// Preserve decimal money strings rather than converting to floating point.
String _groupDecimal(String value) {
  final parts = value.split('.');
  final whole = parts.first.replaceAllMapped(
    RegExp(r'(\d)(?=(\d{3})+$)'),
    (match) => '${match[1]},',
  );
  return [whole, ...parts.skip(1)].join('.');
}

class _Section extends StatelessWidget {
  const _Section({required this.title, required this.child});
  final String title;
  final Widget child;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [SectionHeader(title), const SizedBox(height: 8), child],
  );
}

class _RowsCard extends StatelessWidget {
  const _RowsCard({required this.children});
  final List<Widget> children;
  @override
  Widget build(BuildContext context) => MetronCard(
    padding: EdgeInsets.zero,
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (var i = 0; i < children.length; i++) ...[
          if (i > 0) const Divider(height: 1),
          children[i],
        ],
      ],
    ),
  );
}

class _ValueRow extends StatelessWidget {
  const _ValueRow(this.label, this.value);
  final String label;
  final Widget value;
  @override
  Widget build(BuildContext context) => DefaultTextStyle(
    style: Theme.of(context).textTheme.bodyMedium!
        .copyWith(fontSize: 14, height: 20 / 14),
    child: Wrap(
      alignment: WrapAlignment.spaceBetween,
      crossAxisAlignment: WrapCrossAlignment.center,
      spacing: 16,
      runSpacing: 8,
      children: [
        Text(label, style: TextStyle(color: context.colors.muted)),
        value,
      ],
    ),
  );
}

class _ClassRow extends StatelessWidget {
  const _ClassRow(this.rule);
  final CoinClassRule rule;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          RulesStrings.coinClass(rule.coinClass),
          style: Theme.of(context).textTheme.bodyMedium!
              .copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          spacing: 8,
          children: [
            _ClassValue(
              RulesStrings.stop,
              '${rule.atrMultiplier.toStringAsFixed(1)} ATR',
            ),
            _ClassValue(
              RulesStrings.riskPerTrade,
              _percent(rule.riskPerTradePercent),
            ),
            _ClassValue(
              RulesStrings.largestPosition,
              '${rule.maxPositionPercent.toStringAsFixed(0)}%',
            ),
          ],
        ),
      ],
    ),
  );
}

class _ClassValue extends StatelessWidget {
  const _ClassValue(this.label, this.value);
  final String label, value;
  @override
  Widget build(BuildContext context) => Expanded(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall!
              .copyWith(fontSize: 11, color: context.colors.subtle),
        ),
        const SizedBox(height: 2),
        Num(value),
      ],
    ),
  );
}

class _DisciplineRow extends StatelessWidget {
  const _DisciplineRow(this.rule);
  final DisciplineRule rule;
  @override
  Widget build(BuildContext context) {
    final level = rule.mode == DisciplineMode.block
        ? RiskLevel.limit
        : RiskLevel.caution;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        spacing: 8,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            spacing: 12,
            children: [
              Expanded(
                child: _NumberedCopy(
                  rule.name,
                  style: Theme.of(context).textTheme.bodyMedium!
                      .copyWith(fontSize: 14, fontWeight: FontWeight.w500),
                ),
              ),
              // Enforcement modes are not live risk verdicts. The shared badge
              // has no label override; reuse its presentation tokens locally.
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: level.tint(context.colors),
                  borderRadius: BorderRadius.circular(MetronRadius.tag),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      level.icon,
                      size: 14,
                      color: level.color(context.colors),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      RulesStrings.mode(rule.mode),
                      style: Theme.of(context).textTheme.bodySmall!.copyWith(
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                        color: level.color(context.colors),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          _NumberedCopy(
            rule.description,
            style: Theme.of(context).textTheme.bodySmall!
                .copyWith(color: context.colors.muted),
          ),
          _NumberedCopy(
            rule.history,
            style: Theme.of(context).textTheme.bodySmall!
                .copyWith(fontSize: 12, color: context.colors.subtle),
          ),
        ],
      ),
    );
  }
}

/// Keep prose in Inter and every embedded time or count in Num.
class _NumberedCopy extends StatelessWidget {
  const _NumberedCopy(this.copy, {required this.style});
  final String copy;
  final TextStyle style;
  @override
  Widget build(BuildContext context) {
    final spans = <InlineSpan>[];
    var end = 0;
    for (final match in RegExp(r'\d+(?::\d+)?').allMatches(copy)) {
      spans.add(TextSpan(text: copy.substring(end, match.start)));
      spans.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: Num(match.group(0)!, style: style),
        ),
      );
      end = match.end;
    }
    spans.add(TextSpan(text: copy.substring(end)));
    return Semantics(
      label: copy,
      excludeSemantics: true,
      child: Text.rich(TextSpan(children: spans), style: style),
    );
  }
}

class _EmptyRow extends StatelessWidget {
  const _EmptyRow(this.copy);
  final String copy;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(16),
    child: Text(
      copy,
      style: Theme.of(context).textTheme.bodySmall!
          .copyWith(color: context.colors.muted),
    ),
  );
}
