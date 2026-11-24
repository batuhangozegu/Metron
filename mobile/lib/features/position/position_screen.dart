import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/providers.dart';
import '../../l10n/live_strings.dart';
import '../../l10n/number_formatter.dart';
import '../../router.dart';
import '../../widgets/widgets.dart';
import '../../widgets/account_gate.dart';
import '../../widgets/live_state.dart';

class PositionScreen extends ConsumerStatefulWidget {
  const PositionScreen({super.key, required this.positionId});
  final String positionId;
  @override
  ConsumerState<PositionScreen> createState() => _PositionScreenState();
}

class _PositionScreenState extends ConsumerState<PositionScreen> {
  bool _busy = false;
  String? _error;
  Future<void> _retryStop() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await ref
          .read(apiClientProvider)
          .request(
            '/positions/${Uri.encodeComponent(widget.positionId)}/stop/retry',
            body: {},
          );
      if (mounted) refreshPortfolio(ref);
    } catch (e) {
      if (mounted) setState(() => _error = LiveStrings.of(context).error(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _close() async {
    final s = LiveStrings.of(context);
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(s.close),
        content: Text(
          ref.read(selectedAccountProvider)?.live == true
              ? s.text(
                  'Close this live position at market price?',
                  'Bu canlı pozisyon piyasa fiyatından kapatılsın mı?',
                )
              : s.confirmClose,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(s.cancel),
          ),
          TextButton(
            key: const ValueKey('confirm-close'),
            onPressed: () => Navigator.pop(context, true),
            child: Text(s.close),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await ref
          .read(apiClientProvider)
          .request(
            '/positions/${Uri.encodeComponent(widget.positionId)}/close',
            body: {},
          );
      if (!mounted) return;
      refreshPortfolio(ref);
      context.goNamed(AppRoutes.home);
    } catch (e) {
      if (mounted) setState(() => _error = s.error(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context), f = NumberFormatter.of(context);
    return WarRoomScaffold(
      appBar: AppBar(title: CommandTitle(s.text('Position', 'Pozisyon'))),
      body: AccountGate(
        child: VisibleRefresh(
          refresh: () => ref.invalidate(positionsProvider),
          child: AsyncPanel(
            value: ref.watch(positionsProvider),
            retry: () => ref.invalidate(positionsProvider),
            data: (positions) {
              final p = positions
                  .where((p) => p.id == widget.positionId)
                  .firstOrNull;
              if (p == null) return Center(child: Text(s.missingPosition));
              return ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  Text(
                    '${p.symbol} · ${p.side}',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 16),
                  MetronCard(
                    child: Column(
                      children: [
                        ValueRow(
                          s.text('Quantity', 'Miktar'),
                          f.decimal(p.value('quantity')),
                        ),
                        ValueRow(
                          s.text('Entry', 'Giriş'),
                          f.decimal(p.value('entry_price')),
                        ),
                        ValueRow(
                          s.text('Latest price', 'Son fiyat'),
                          f.decimal(p.value('last_price')),
                        ),
                        ValueRow(
                          s.text('Stop', 'Stop'),
                          f.decimal(p.value('stop')),
                        ),
                        ValueRow(s.risk, f.decimal(p.value('initial_risk'))),
                        ValueRow(
                          s.text('Unrealized PnL', 'Gerçekleşmemiş K/Z'),
                          f.decimal(p.value('unrealized_pnl')),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (p.exchangeStop case final stop?) ...[
                    ValueRow(s.exchangeStop, s.stopStatus(stop.status)),
                    if (stop.orderId != null)
                      ValueRow(
                        s.text('Order ID', 'Emir kimliği'),
                        stop.orderId!,
                      ),
                    if (stop.error != null) Text(stop.error!),
                    if (stop.status == 'failed' && p.value('status') == 'open')
                      TextButton(
                        onPressed: _busy ? null : _retryStop,
                        child: Text(s.retry),
                      ),
                  ],
                  if (p.json['close_reason'] case final String reason)
                    ValueRow(
                      s.text('Close reason', 'Kapanış nedeni'),
                      s.closeReason(reason),
                    ),
                  if (_error != null) Text(_error!),
                  if (_busy) const LinearProgressIndicator(),
                  if (p.value('status') == 'open')
                    PrimaryButton(
                      label: s.close,
                      onPressed: _busy ? null : _close,
                    ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }
}
