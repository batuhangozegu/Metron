import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../l10n/number_formatter.dart';
import '../../data/models.dart';
import '../../data/providers.dart';
import '../../data/risk_level.dart';
import '../../l10n/app_strings.dart';
import '../../l10n/language_selector.dart';
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
    return WarRoomScaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          key: const PageStorageKey('rules-scroll'),
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Align(
                alignment: Alignment.centerRight,
                child: LanguageSelector(),
              ),
              CommandTitle(
                AppStrings.of(context).rules,
                style: MetronType.title.copyWith(
                  fontSize: 28,
                  height: 1.12,
                  letterSpacing: -.22,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                RulesStrings.of(context).introduction,
                style: Theme.of(context).textTheme.bodySmall!
                    .copyWith(color: context.colors.muted),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.of(context).accountLimits,
                child: MetronCard(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    spacing: 12,
                    children: [
                      _ValueRow(
                        AppStrings.of(context).riskPerTradeLimit,
                        Num(
                          NumberFormatter.of(context)
                              .percent(limits.riskPerTradePercent),
                        ),
                      ),
                      _ValueRow(
                        AppStrings.of(context).dailyLosingTradesLimit,
                        Num(
                          NumberFormatter.of(context).number(
                            limits.maxLosingTradesPerDay,
                            decimalPlaces: 0,
                          ),
                        ),
                      ),
                      _ValueRow(
                        AppStrings.of(context).weeklyStoppedTradesLimit,
                        Num(
                          NumberFormatter.of(context).number(
                            limits.maxStoppedTradesPerWeek,
                            decimalPlaces: 0,
                          ),
                        ),
                      ),
                      _ValueRow(
                        RulesStrings.of(context).openRisk,
                        Num(
                          NumberFormatter.of(context)
                              .percent(limits.openRiskPercent),
                        ),
                      ),
                      _ValueRow(
                        RulesStrings.of(context).dailyCap,
                        Num(
                          NumberFormatter.of(context)
                              .percent(limits.dailyLossPercent),
                        ),
                      ),
                      _ValueRow(
                        RulesStrings.of(context).weeklyCap,
                        Num(
                          NumberFormatter.of(context)
                              .percent(limits.weeklyLossPercent),
                        ),
                      ),
                      _ValueRow(
                        RulesStrings.of(context).positions,
                        Num(
                          NumberFormatter.of(
                            context,
                          ).number(limits.maxOpenPositions, decimalPlaces: 0),
                        ),
                      ),
                      _ValueRow(
                        RulesStrings.of(context).riskReward,
                        Num(
                          NumberFormatter.of(
                            context,
                          ).number(limits.targetRiskReward, decimalPlaces: 1),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.of(context).coinClasses,
                child: _RowsCard(
                  children: [
                    for (final rule in classes) _ClassRow(rule),
                    if (classes.isEmpty)
                      _EmptyRow(RulesStrings.of(context).noClasses),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.of(context).discipline,
                child: _RowsCard(
                  children: [
                    for (final rule in discipline) _DisciplineRow(rule),
                    if (discipline.isEmpty)
                      _EmptyRow(RulesStrings.of(context).noDiscipline),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              _Section(
                title: RulesStrings.of(context).account,
                child: _RowsCard(
                  children: [
                    Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      child: _ValueRow(
                        RulesStrings.of(context).tradingAccount,
                        Wrap(
                          alignment: WrapAlignment.end,
                          spacing: 4,
                          children: [
                            Text(RulesStrings.of(context).paper),
                            Num(
                              '${NumberFormatter.of(context).decimal(account.equity)} USDT',
                            ),
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
                        RulesStrings.of(context).twoFactor,
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
                                  ? RulesStrings.of(context).on
                                  : RulesStrings.of(context).off,
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
                        child: Text(RulesStrings.of(context).signOut),
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
          RulesStrings.of(context).coinClass(rule.coinClass),
          style: Theme.of(context).textTheme.bodyMedium!
              .copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          spacing: 8,
          children: [
            _ClassValue(
              RulesStrings.of(context).stop,
              '${NumberFormatter.of(context).number(rule.atrMultiplier, decimalPlaces: 1)} ATR',
            ),
            _ClassValue(
              RulesStrings.of(context).riskPerTrade,
              NumberFormatter.of(context).percent(rule.riskPerTradePercent),
            ),
            _ClassValue(
              AppStrings.of(context).spotAllocationCap,
              NumberFormatter.of(context)
                  .percent(rule.maxPositionPercent, decimalPlaces: 0),
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
                  RulesStrings.of(context).ruleName(rule),
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
                      RulesStrings.of(context).mode(rule.mode),
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
            RulesStrings.of(context).ruleDescription(rule),
            style: Theme.of(context).textTheme.bodySmall!
                .copyWith(color: context.colors.muted),
          ),
          _NumberedCopy(
            RulesStrings.of(context).ruleHistory(rule),
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
    final localized = NumberFormatter.of(context).fixtureText(copy);
    final spans = <InlineSpan>[];
    var end = 0;
    for (final match in RegExp(r'\d+(?:[,:.]\d+)*').allMatches(localized)) {
      spans.add(TextSpan(text: localized.substring(end, match.start)));
      spans.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: Num(match.group(0)!, style: style),
        ),
      );
      end = match.end;
    }
    spans.add(TextSpan(text: localized.substring(end)));
    return Semantics(
      label: localized,
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
