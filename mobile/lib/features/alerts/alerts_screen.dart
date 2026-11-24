import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/live_models.dart';
import '../../data/providers.dart';
import '../../l10n/alert_strings.dart';
import '../../l10n/app_strings.dart';
import '../../widgets/live_state.dart';
import '../../widgets/widgets.dart';
import 'alarm_form.dart';

class AlertsScreen extends ConsumerStatefulWidget {
  const AlertsScreen({super.key});
  @override
  ConsumerState<AlertsScreen> createState() => _AlertsScreenState();
}

class _AlertsScreenState extends ConsumerState<AlertsScreen> {
  final _cancelling = <String>{};
  Future<void> _cancel(Alert alert) async {
    setState(() => _cancelling.add(alert.id));
    try {
      await ref
          .read(apiClientProvider)
          .request(
            '/alerts/${Uri.encodeComponent(alert.id)}',
            method: 'DELETE',
          );
      if (mounted) ref.invalidate(alertsProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              AlertStrings(Localizations.localeOf(context).languageCode == 'tr')
                  .error(e),
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _cancelling.remove(alert.id));
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = AlertStrings(
      Localizations.localeOf(context).languageCode == 'tr',
    );
    return WarRoomScaffold(
      appBar: AppBar(
        title: CommandTitle(AppStrings.of(context).alerts),
        actions: [
          IconButton(
            key: const ValueKey('add-alert'),
            tooltip: s.add,
            icon: const Icon(Icons.add),
            onPressed: () => showDialog<void>(
              context: context,
              builder: (_) => const AlarmForm(),
            ),
          ),
        ],
      ),
      body: AsyncPanel(
        value: ref.watch(alertsProvider),
        retry: () => ref.invalidate(alertsProvider),
        data: (alerts) => RefreshIndicator(
          onRefresh: () async {
            ref.invalidate(alertsProvider);
            await ref.read(alertsProvider.future);
          },
          child: ListView(
            padding: const EdgeInsets.all(16),
            physics: const AlwaysScrollableScrollPhysics(),
            children: [
              if (alerts.isEmpty) Text(s.empty),
              for (final alert in alerts)
                MetronCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${alert.symbol} · ${s.directionLabel(alert.direction)} ${alert.price}',
                      ),
                      Text(s.alertStatus(alert.status)),
                      if (alert.note != null) Text(alert.note!),
                      if (alert.triggeredPrice != null)
                        Text(s.triggered(alert.symbol, alert.triggeredPrice!)),
                      if (alert.status == 'active')
                        TextButton(
                          key: ValueKey('cancel-${alert.id}'),
                          onPressed: _cancelling.contains(alert.id)
                              ? null
                              : () => _cancel(alert),
                          child: Text(s.cancel),
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
