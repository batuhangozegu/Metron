import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/providers.dart';
import '../data/live_models.dart';
import '../l10n/live_strings.dart';
import 'live_state.dart';
import 'metron_button.dart';

class AccountGate extends ConsumerWidget {
  const AccountGate({super.key, required this.child});
  final Widget child;
  @override
  Widget build(BuildContext context, WidgetRef ref) => AsyncPanel(
    value: ref.watch(accountsProvider),
    retry: () => ref.invalidate(accountsProvider),
    data: (accounts) => accounts.isEmpty
        ? const AccountForm()
        : Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const AccountPicker(),
              Expanded(child: child),
            ],
          ),
  );
}

class AccountPicker extends ConsumerWidget {
  const AccountPicker({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final accounts = ref.watch(accountsProvider).value ?? [];
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: DropdownButtonFormField<String>(
              key: ValueKey(ref.watch(selectedAccountIdProvider)),
              initialValue: ref.watch(selectedAccountIdProvider),
              isExpanded: true,
              decoration: InputDecoration(
                labelText: LiveStrings.of(context).account,
              ),
              items: accounts
                  .map(
                    (a) => DropdownMenuItem(
                      value: a.id,
                      child: Text(
                        '${a.name} · ${a.market}',
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  )
                  .toList(),
              onChanged: (id) async {
                if (id == null) return;
                try {
                  await ref.read(accountsProvider.notifier).select(id);
                } catch (e) {
                  if (context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(LiveStrings.of(context).error(e))),
                    );
                  }
                }
              },
            ),
          ),
          IconButton(
            tooltip: LiveStrings.of(context).create,
            icon: const Icon(Icons.add),
            onPressed: () => showDialog<void>(
              context: context,
              builder: (_) => const Dialog(
                child: SizedBox(width: 400, child: AccountForm(dialog: true)),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class AccountForm extends ConsumerStatefulWidget {
  const AccountForm({super.key, this.dialog = false});
  final bool dialog;
  @override
  ConsumerState<AccountForm> createState() => _AccountFormState();
}

class _AccountFormState extends ConsumerState<AccountForm> {
  final _name = TextEditingController(),
      _balance = TextEditingController(text: '10000');
  String _market = 'spot';
  bool _busy = false;
  String? _error;
  @override
  void dispose() {
    _name.dispose();
    _balance.dispose();
    super.dispose();
  }

  Future<void> _create() async {
    final s = LiveStrings.of(context);
    if (_name.text.trim().isEmpty ||
        _name.text.trim().runes.length > 100 ||
        !RegExp(r'^\d{1,8}(\.\d{1,8})?$').hasMatch(_balance.text) ||
        !validPositiveDecimal(_balance.text, '10000000')) {
      setState(() => _error = s.invalidAccount);
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await ref
          .read(accountsProvider.notifier)
          .create(_name.text, _balance.text, _market);
      if (mounted && widget.dialog) Navigator.of(context).pop();
    } catch (e) {
      if (mounted) setState(() => _error = s.error(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: 16,
        children: [
          Text(s.noAccounts),
          TextField(
            key: const ValueKey('account-name'),
            controller: _name,
            decoration: InputDecoration(labelText: s.name),
          ),
          TextField(
            key: const ValueKey('account-balance'),
            controller: _balance,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(labelText: s.balance),
          ),
          DropdownButtonFormField<String>(
            initialValue: _market,
            items: const [
              DropdownMenuItem(value: 'spot', child: Text('Spot')),
              DropdownMenuItem(value: 'futures', child: Text('Futures')),
            ],
            onChanged: _busy ? null : (v) => setState(() => _market = v!),
          ),
          if (_error != null) Text(_error!),
          if (_busy) const LinearProgressIndicator(),
          PrimaryButton(label: s.create, onPressed: _busy ? null : _create),
        ],
      ),
    );
  }
}
