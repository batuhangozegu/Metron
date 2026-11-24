import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../data/api_client.dart';
import '../../data/live_models.dart';
import '../../data/providers.dart';
import '../../l10n/app_strings.dart';
import '../../l10n/live_strings.dart';
import '../../router.dart';
import '../../widgets/widgets.dart';
import '../../widgets/account_gate.dart';
import '../../widgets/live_state.dart';
import 'check_layout.dart';
import 'check_strings.dart';

class CheckScreen extends ConsumerStatefulWidget {
  const CheckScreen({super.key, this.pair});
  final String? pair;
  @override
  ConsumerState<CheckScreen> createState() => _CheckScreenState();
}

class _CheckScreenState extends ConsumerState<CheckScreen> {
  final _leverage = TextEditingController(text: '1');
  String? _symbol;
  String _side = 'long';
  RiskCheck? _check;
  Json? _checkedOrder;
  String? _error;
  bool _busy = false;
  bool _confirming = false;
  @override
  void dispose() {
    _leverage.dispose();
    super.dispose();
  }

  void _reset() {
    _check = null;
    _checkedOrder = null;
    _error = null;
  }

  Future<void> _run(MarketQuote quote, TradingAccount account) async {
    final s = LiveStrings.of(context);
    final leverage = int.tryParse(_leverage.text);
    if (account.futures &&
        (leverage == null || leverage < 1 || leverage > 125)) {
      setState(() => _error = s.invalidLeverage);
      return;
    }
    final order = <String, dynamic>{
      'account_id': account.id,
      'symbol': quote.symbol,
      'side': _side,
      if (account.futures) 'leverage': leverage,
    };
    setState(() {
      _busy = true;
      _reset();
    });
    try {
      final data =
          await ref
                  .read(apiClientProvider)
                  .request(
                    '/risk/check',
                    body: {...order, 'entry_price': quote.price},
                  )
              as Json;
      if (!mounted || ref.read(selectedAccountProvider)?.id != account.id) {
        return;
      }
      final check = RiskCheck(data);
      if (check.blocked) {
        context.pushNamed(AppRoutes.checkBlocked, extra: check);
      } else {
        setState(() {
          _check = check;
          _checkedOrder = order;
        });
      }
    } catch (e) {
      if (mounted) setState(() => _error = s.error(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _place() async {
    if (_busy || _confirming || _checkedOrder == null) return;
    final s = LiveStrings.of(context);
    final order = Map<String, dynamic>.from(_checkedOrder!);
    final account = ref.read(selectedAccountProvider);
    if (account?.id != order['account_id']) return;
    if (account!.mainnet) {
      setState(() => _confirming = true);
      final confirmed = await showModalBottomSheet<bool>(
        context: context,
        builder: (context) => SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              spacing: 16,
              children: [
                Text(s.realMoney, style: const TextStyle(color: Colors.red)),
                Text(s.confirmLive),
                FilledButton(
                  key: const ValueKey('confirm-live'),
                  style: FilledButton.styleFrom(backgroundColor: Colors.red),
                  onPressed: () => Navigator.pop(context, true),
                  child: Text(s.placeLive),
                ),
                TextButton(
                  onPressed: () => Navigator.pop(context, false),
                  child: Text(s.cancel),
                ),
              ],
            ),
          ),
        ),
      );
      if (!mounted) return;
      setState(() => _confirming = false);
      if (confirmed != true ||
          ref.read(selectedAccountProvider)?.id != account.id) {
        return;
      }
      order['confirm_live'] = true;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final result =
          await ref.read(apiClientProvider).request('/orders', body: order)
              as Json;
      if (!mounted) return;
      refreshPortfolio(ref);
      final position = result['position'] as Json;
      context.goNamed(
        AppRoutes.position,
        pathParameters: {'id': position['id'] as String},
      );
    } catch (e) {
      if (!mounted) return;
      if (e is ApiError && e.id == 'risk_blocked' && e.data['check'] is Json) {
        setState(_reset);
        context.pushNamed(
          AppRoutes.checkBlocked,
          extra: RiskCheck(e.data['check'] as Json),
        );
      } else if (e is ApiError && e.id == 'live_confirmation_required') {
        setState(() => _error = s.confirmLive);
        ref.invalidate(accountsProvider);
      } else {
        setState(() => _error = s.error(e));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context);
    ref.listen(selectedAccountIdProvider, (_, next) {
      setState(() {
        _reset();
        _side = 'long';
      });
    });
    final account = ref.watch(selectedAccountProvider);
    return WarRoomScaffold(
      appBar: AppBar(title: CommandTitle(AppStrings.of(context).check)),
      bottomNavigationBar: _check == null
          ? null
          : SafeArea(
              top: false,
              child: FixedMainAction(
                label: account?.live == true
                    ? s.placeLive
                    : CheckStrings.of(context).place,
                onPressed: _busy ? null : _place,
              ),
            ),
      body: AccountGate(
        child: AsyncPanel(
          value: ref.watch(marketsProvider),
          retry: () => ref.invalidate(marketsProvider),
          data: (quotes) {
            if (quotes.isEmpty || account == null) {
              return Center(child: Text(s.noMarkets));
            }
            final quote =
                quotes.where((q) => q.symbol == _symbol).firstOrNull ??
                quotes.where((q) => q.pair == widget.pair).firstOrNull ??
                quotes.first;
            return ListView(
              padding: const EdgeInsets.all(16),
              children: [
                if (account.live && account.environment == 'testnet')
                  Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: Text(
                      s.testnetNotice,
                      style: const TextStyle(color: Colors.amber),
                    ),
                  ),
                DropdownButtonFormField<String>(
                  key: ValueKey(quote.symbol),
                  initialValue: quote.symbol,
                  decoration: InputDecoration(
                    labelText: CheckStrings.of(context).pair,
                  ),
                  items: quotes
                      .map(
                        (q) => DropdownMenuItem(
                          value: q.symbol,
                          child: Text(q.pair),
                        ),
                      )
                      .toList(),
                  onChanged: _busy
                      ? null
                      : (v) => setState(() {
                          _symbol = v;
                          _reset();
                        }),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  key: ValueKey('${account.id}-$_side'),
                  initialValue: _side,
                  decoration: InputDecoration(labelText: s.side),
                  items: [
                    DropdownMenuItem(value: 'long', child: Text(s.long)),
                    if (account.futures)
                      DropdownMenuItem(value: 'short', child: Text(s.short)),
                  ],
                  onChanged: _busy
                      ? null
                      : (v) => setState(() {
                          _side = v!;
                          _reset();
                        }),
                ),
                if (account.futures)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: TextField(
                      key: const ValueKey('leverage'),
                      enabled: !_busy,
                      controller: _leverage,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(labelText: s.leverage),
                      onChanged: (_) => setState(_reset),
                    ),
                  ),
                const SizedBox(height: 16),
                PrimaryButton(
                  label: s.runCheck,
                  onPressed: _busy || quote.price == null
                      ? null
                      : () => _run(quote, account),
                ),
                if (_busy) const LinearProgressIndicator(),
                if (_error != null)
                  Padding(
                    padding: const EdgeInsets.all(12),
                    child: Text(_error!),
                  ),
                const SizedBox(height: 16),
                if (_check == null)
                  Text(quote.price == null ? s.unavailable : s.noCheck)
                else
                  CheckLayout(check: _check!),
              ],
            );
          },
        ),
      ),
    );
  }
}
