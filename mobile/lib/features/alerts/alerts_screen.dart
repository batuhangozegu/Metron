import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../l10n/number_formatter.dart';
import '../../data/models.dart';
import '../../data/providers.dart';
import '../../data/risk_level.dart';
import '../../l10n/app_strings.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/widgets.dart';
import 'alarm_form.dart';
import 'alerts_strings.dart';

class AlertsScreen extends ConsumerStatefulWidget {
  const AlertsScreen({super.key});
  @override
  ConsumerState<AlertsScreen> createState() => _AlertsScreenState();
}

class _AlertsScreenState extends ConsumerState<AlertsScreen> {
  final _addedAlarms = <PriceAlarm>[];
  Future<void> _addAlarm() async {
    final alarm = await showDialog<PriceAlarm>(
      context: context,
      builder: (_) => AlarmForm(prices: ref.read(pricesProvider)),
    );
    if (mounted && alarm != null) setState(() => _addedAlarms.add(alarm));
  }

  @override
  Widget build(BuildContext context) {
    // Anchor today to the sample account clock, not the device's real date.
    final day = ref.watch(accountProvider).asOf.toUtc();
    final events = ref.watch(alertsProvider).where((event) {
      final at = event.at.toUtc();
      return at.year == day.year && at.month == day.month && at.day == day.day;
    }).toList()..sort((a, b) => b.at.compareTo(a.at));
    final alarms = [...ref.watch(priceAlarmsProvider), ..._addedAlarms];
    return WarRoomScaffold(
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  CommandTitle(
                    AppStrings.of(context).alerts,
                    style: MetronType.title.copyWith(
                      fontSize: 28,
                      height: 1.12,
                      letterSpacing: -.22,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 7,
                      vertical: 3,
                    ),
                    decoration: BoxDecoration(
                      border: Border.all(color: context.colors.line),
                      borderRadius: BorderRadius.circular(MetronRadius.tag),
                    ),
                    child: Text(
                      AlertsStrings.of(context).paper,
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w500,
                        letterSpacing: .66,
                        color: context.colors.muted,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.fromLTRB(16, 4, 16, 20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    SectionHeader(AlertsStrings.of(context).today),
                    const SizedBox(height: 8),
                    _Rows(
                      children: [
                        if (events.isEmpty)
                          _Empty(AlertsStrings.of(context).noEvents),
                        for (final event in events) _EventRow(event),
                      ],
                    ),
                    const SizedBox(height: 20),
                    SectionHeader(
                      AlertsStrings.of(context).priceAlarms,
                      trailing: TextButton(
                        onPressed: _addAlarm,
                        style: TextButton.styleFrom(
                          minimumSize: const Size(44, 44),
                          padding: const EdgeInsets.symmetric(horizontal: 4),
                          foregroundColor: context.colors.text,
                          textStyle: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            decoration: TextDecoration.underline,
                          ),
                        ),
                        child: Text(AlertsStrings.of(context).addAlarm),
                      ),
                    ),
                    const SizedBox(height: 8),
                    _Rows(
                      children: [
                        if (alarms.isEmpty)
                          _Empty(AlertsStrings.of(context).noAlarms),
                        for (final alarm in alarms) _AlarmRow(alarm),
                      ],
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

class _Rows extends StatelessWidget {
  const _Rows({required this.children});
  final List<Widget> children;
  @override
  Widget build(BuildContext context) => MetronCard(
    padding: EdgeInsets.zero,
    child: Column(
      children: [
        for (var i = 0; i < children.length; i++) ...[
          if (i > 0) const Divider(height: 1),
          children[i],
        ],
      ],
    ),
  );
}

class _EventRow extends StatelessWidget {
  const _EventRow(this.event);
  final AlertEvent event;
  @override
  Widget build(BuildContext context) {
    final level = switch (event.kind) {
      AlertKind.cooldown => RiskLevel.caution,
      AlertKind.orderBlocked => RiskLevel.limit,
      _ => null,
    };
    final at = event.at.toUtc();
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.only(top: 1),
            child: Icon(
              level?.icon ??
                  (event.kind == AlertKind.stopTriggered
                      ? Icons.vertical_align_bottom
                      : Icons.notifications_none),
              size: 18,
              color: level?.color(context.colors) ?? context.colors.muted,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: _NumberedCopy(
                        AlertsStrings.of(context).eventTitle(event),
                        style: const TextStyle(
                          fontSize: 14,
                          height: 20 / 14,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Num(
                      '${at.hour.toString().padLeft(2, '0')}:${at.minute.toString().padLeft(2, '0')}',
                      style: TextStyle(
                        fontSize: 12,
                        color: context.colors.subtle,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 3),
                _NumberedCopy(
                  AlertsStrings.of(context).eventBody(event),
                  style: MetronType.small.copyWith(color: context.colors.muted),
                  pnl: event.pnl,
                ),
                if (level != null) ...[
                  const SizedBox(height: 4),
                  RiskBadge(level),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// Words stay in Inter; embedded prices and times use the shared Num widget.
class _NumberedCopy extends StatelessWidget {
  const _NumberedCopy(this.text, {required this.style, this.pnl});
  final String text;
  final TextStyle style;
  final String? pnl;
  @override
  Widget build(BuildContext context) {
    final localized = NumberFormatter.of(context).fixtureText(text);
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
    if (pnl != null) {
      spans.add(const TextSpan(text: ' '));
      spans.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: PnlText(double.parse(pnl!), style: style),
        ),
      );
    }
    return Text.rich(TextSpan(children: spans), style: style);
  }
}

class _AlarmRow extends ConsumerWidget {
  const _AlarmRow(this.alarm);
  final PriceAlarm alarm;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final price = ref.watch(coinPriceProvider(alarm.pair));
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      child: LayoutBuilder(
        builder: (context, constraints) => Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${alarm.pair} ${AlertsStrings.of(context).direction(alarm.direction)}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Row(
                    children: [
                      Text(
                        AlertsStrings.of(context).now,
                        style: TextStyle(
                          fontSize: 12,
                          color: context.colors.subtle,
                        ),
                      ),
                      Flexible(
                        child: Num(
                          price == null
                              ? AlertsStrings.of(context).unavailable
                              : NumberFormatter.of(context)
                                    .decimal(price.price, minimumDecimals: 2),
                          style: TextStyle(
                            fontSize: 12,
                            color: context.colors.subtle,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            ConstrainedBox(
              constraints: BoxConstraints(maxWidth: constraints.maxWidth / 2),
              child: Num(
                NumberFormatter.of(context)
                    .decimal(alarm.targetPrice, minimumDecimals: 2),
                textAlign: TextAlign.end,
                style: const TextStyle(fontSize: 15),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _Empty extends StatelessWidget {
  const _Empty(this.message);
  final String message;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(16),
    child: Text(
      message,
      style: MetronType.small.copyWith(color: context.colors.muted),
    ),
  );
}
