import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/api_client.dart';
import '../../data/live_models.dart';
import '../../data/providers.dart';
import '../../l10n/alert_strings.dart';

/// Polls for in-app notifications only while the authenticated app is foregrounded.
class AlertNotifications extends ConsumerStatefulWidget {
  const AlertNotifications({super.key, required this.child});
  final Widget child;
  @override
  ConsumerState<AlertNotifications> createState() => _AlertNotificationsState();
}

class _AlertNotificationsState extends ConsumerState<AlertNotifications>
    with WidgetsBindingObserver {
  Timer? _timer;
  DateTime _since = DateTime.now().toUtc();
  final _seen = <String>{};
  bool _busy = false;
  bool get _foreground =>
      WidgetsBinding.instance.lifecycleState == null ||
      WidgetsBinding.instance.lifecycleState == AppLifecycleState.resumed;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _start();
  }

  void _start() {
    _timer?.cancel();
    if (_foreground) {
      _timer = Timer.periodic(const Duration(seconds: 15), (_) => _poll());
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) => _start();

  Future<void> _poll() async {
    if (_busy || !_foreground || ref.read(sessionProvider).value == null) {
      return;
    }
    _busy = true;
    final started = DateTime.now().toUtc();
    try {
      final data =
          await ref
                  .read(apiClientProvider)
                  .request(
                    '/alerts?triggered_since=${Uri.encodeQueryComponent(_since.toIso8601String())}',
                  )
              as List;
      if (!mounted || !_foreground || ref.read(sessionProvider).value == null) {
        return;
      }
      // Advance only after success; overlapping responses are deduplicated by ID.
      _since = started;
      for (final json in data) {
        final alert = Alert.fromJson(json as Json);
        if (!_foreground || alert.status != 'triggered') continue;
        if (!_seen.add(alert.id)) continue;
        final s = AlertStrings(
          Localizations.localeOf(context).languageCode == 'tr',
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              s.triggered(alert.symbol, alert.triggeredPrice ?? alert.price),
            ),
          ),
        );
        ref.invalidate(alertsProvider);
      }
    } catch (_) {
      // Keep the cursor so a temporary failure cannot lose triggers.
    } finally {
      _busy = false;
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
