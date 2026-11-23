import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../l10n/live_strings.dart';
import 'metron_button.dart';
import 'num.dart';

class AsyncPanel<T> extends StatelessWidget {
  const AsyncPanel({
    super.key,
    required this.value,
    required this.retry,
    required this.data,
  });
  final AsyncValue<T> value;
  final VoidCallback retry;
  final Widget Function(T) data;
  @override
  Widget build(BuildContext context) => value.when(
    skipLoadingOnRefresh: true,
    data: data,
    error: (error, _) => ErrorPanel(error: error, retry: retry),
    loading: () => Padding(
      padding: const EdgeInsets.all(24),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const CircularProgressIndicator(),
            Text(LiveStrings.of(context).loading),
          ],
        ),
      ),
    ),
  );
}

class ErrorPanel extends StatelessWidget {
  const ErrorPanel({super.key, required this.error, required this.retry});
  final Object error;
  final VoidCallback retry;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(16),
    child: Column(
      mainAxisSize: MainAxisSize.min,
      spacing: 12,
      children: [
        Text(LiveStrings.of(context).error(error)),
        SecondaryButton(label: LiveStrings.of(context).retry, onPressed: retry),
      ],
    ),
  );
}

/// Indexed shell branches remain mounted. TickerMode and the current route,
/// together with app lifecycle, ensure hidden screens never poll.
class VisibleRefresh extends StatefulWidget {
  const VisibleRefresh({super.key, required this.refresh, required this.child});
  final VoidCallback refresh;
  final Widget child;
  @override
  State<VisibleRefresh> createState() => _VisibleRefreshState();
}

class _VisibleRefreshState extends State<VisibleRefresh>
    with WidgetsBindingObserver {
  Timer? _timer;
  bool _resumed = true;
  @override
  void initState() {
    super.initState();
    final lifecycle = WidgetsBinding.instance.lifecycleState;
    _resumed = lifecycle == null || lifecycle == AppLifecycleState.resumed;
    WidgetsBinding.instance.addObserver(this);
    _timer = Timer.periodic(const Duration(seconds: 5), (_) {
      if (_resumed &&
          TickerMode.valuesOf(context).enabled &&
          (ModalRoute.of(context)?.isCurrent ?? true)) {
        widget.refresh();
      }
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _resumed = state == AppLifecycleState.resumed;
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

class ValueRow extends StatelessWidget {
  const ValueRow(this.label, this.value, {super.key});
  final String label, value;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: Row(
      children: [
        Expanded(child: Text(label)),
        const SizedBox(width: 12),
        Flexible(child: Num(value, textAlign: TextAlign.end)),
      ],
    ),
  );
}
