import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/api_client.dart';
import '../../data/live_models.dart';
import '../../data/providers.dart';
import '../../l10n/alert_strings.dart';
import '../../widgets/live_state.dart';

class AlarmForm extends ConsumerStatefulWidget {
  const AlarmForm({super.key});
  @override
  ConsumerState<AlarmForm> createState() => _AlarmFormState();
}

class _AlarmFormState extends ConsumerState<AlarmForm> {
  final _price = TextEditingController(), _note = TextEditingController();
  String? _symbol, _error;
  String _direction = 'above';
  bool _busy = false;
  @override
  void dispose() {
    _price.dispose();
    _note.dispose();
    super.dispose();
  }

  Future<void> _save(String symbol) async {
    final s = AlertStrings(
      Localizations.localeOf(context).languageCode == 'tr',
    );
    final price = _price.text.trim();
    if (!RegExp(r'^\d+(\.\d+)?$').hasMatch(price) ||
        !RegExp(r'[1-9]').hasMatch(price)) {
      setState(() => _error = s.invalid);
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await ref
          .read(apiClientProvider)
          .request(
            '/alerts',
            body: {
              'symbol': symbol,
              'direction': _direction,
              'price': price,
              if (_note.text.trim().isNotEmpty) 'note': _note.text.trim(),
            },
          );
      if (!mounted) return;
      ref.invalidate(alertsProvider);
      Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        setState(
          () =>
              _error = e is ApiError && e.status == 409 ? s.limit : s.error(e),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = AlertStrings(
      Localizations.localeOf(context).languageCode == 'tr',
    );
    return AlertDialog(
      title: Text(s.add),
      content: SizedBox(
        width: 350,
        child: AsyncPanel<List<MarketQuote>>(
          value: ref.watch(marketsProvider),
          retry: () => ref.invalidate(marketsProvider),
          data: (markets) {
            if (markets.isEmpty) return Text(s.noMarkets);
            final symbol = _symbol ?? markets.first.symbol;
            return SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                spacing: 12,
                children: [
                  DropdownButtonFormField<String>(
                    initialValue: symbol,
                    decoration: InputDecoration(labelText: s.symbol),
                    items: [
                      for (final m in markets)
                        DropdownMenuItem(
                          value: m.symbol,
                          child: Text(m.symbol),
                        ),
                    ],
                    onChanged: _busy
                        ? null
                        : (v) => setState(() => _symbol = v),
                  ),
                  DropdownButtonFormField<String>(
                    initialValue: _direction,
                    decoration: InputDecoration(labelText: s.direction),
                    items: [
                      for (final d in ['above', 'below'])
                        DropdownMenuItem(
                          value: d,
                          child: Text(s.directionLabel(d)),
                        ),
                    ],
                    onChanged: _busy
                        ? null
                        : (v) => setState(() => _direction = v!),
                  ),
                  TextField(
                    key: const ValueKey('alert-price'),
                    controller: _price,
                    enabled: !_busy,
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    decoration: InputDecoration(labelText: s.price),
                  ),
                  TextField(
                    key: const ValueKey('alert-note'),
                    controller: _note,
                    enabled: !_busy,
                    decoration: InputDecoration(labelText: s.note),
                  ),
                  if (_error != null) Text(_error!),
                  if (_busy) const LinearProgressIndicator(),
                  FilledButton(
                    key: const ValueKey('save-alert'),
                    onPressed: _busy ? null : () => _save(symbol),
                    child: Text(s.add),
                  ),
                ],
              ),
            );
          },
        ),
      ),
      actions: [
        TextButton(
          onPressed: _busy ? null : () => Navigator.pop(context),
          child: Text(s.cancel),
        ),
      ],
    );
  }
}
