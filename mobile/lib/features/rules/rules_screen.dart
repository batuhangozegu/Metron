import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/api_client.dart';
import '../../data/live_models.dart';
import '../../data/providers.dart';
import '../../l10n/app_strings.dart';
import '../../l10n/live_strings.dart';
import '../../l10n/language_selector.dart';
import '../../widgets/widgets.dart';
import '../../widgets/account_gate.dart';
import '../../widgets/live_state.dart';

class RulesScreen extends ConsumerWidget {
  const RulesScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) => WarRoomScaffold(
    appBar: AppBar(
      title: CommandTitle(AppStrings.of(context).rules),
      actions: const [LanguageSelector()],
    ),
    body: AccountGate(
      child: AsyncPanel(
        value: ref.watch(riskProfileProvider),
        retry: () => ref.invalidate(riskProfileProvider),
        data: (profile) => profile == null
            ? Text(LiveStrings.of(context).noAccounts)
            : _ProfileForm(
                key: ValueKey(ref.watch(selectedAccountIdProvider)),
                profile: profile,
              ),
      ),
    ),
  );
}

class _ProfileForm extends ConsumerStatefulWidget {
  const _ProfileForm({super.key, required this.profile});
  final Json profile;
  @override
  ConsumerState<_ProfileForm> createState() => _ProfileFormState();
}

class _ProfileFormState extends ConsumerState<_ProfileForm> {
  final _form = GlobalKey<FormState>();
  late final _fields = {
    for (final key in profileFields)
      key: TextEditingController(text: widget.profile[key].toString()),
  };
  bool _busy = false;
  String? _notice;
  @override
  void dispose() {
    for (final c in _fields.values) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    if (!_form.currentState!.validate()) return;
    final s = LiveStrings.of(context);
    final account = ref.read(selectedAccountProvider)!;
    final body = <String, dynamic>{
      for (final key in profileFields)
        key: profileCounts.contains(key)
            ? int.parse(_fields[key]!.text)
            : _fields[key]!.text,
    };
    setState(() {
      _busy = true;
      _notice = null;
    });
    try {
      await ref
          .read(apiClientProvider)
          .request(
            '/accounts/${account.id}/risk-profile',
            method: 'PUT',
            body: body,
          );
      if (mounted) {
        setState(() => _notice = s.saved);
        ref.invalidate(riskProfileProvider);
      }
    } catch (e) {
      if (mounted) setState(() => _notice = s.error(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context);
    return Form(
      key: _form,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (widget.profile['updated_at'] == null) Text(s.defaults),
          for (final key in profileFields)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 10),
              child: TextFormField(
                key: ValueKey(key),
                controller: _fields[key],
                enabled: !_busy,
                decoration: InputDecoration(labelText: s.profileLabel(key)),
                keyboardType: TextInputType.numberWithOptions(
                  decimal: !profileCounts.contains(key),
                ),
                validator: (value) => profileCounts.contains(key)
                    ? validRiskCount(value ?? '')
                          ? null
                          : s.invalidCount
                    : validPositiveDecimal(value ?? '', '100')
                    ? null
                    : s.invalidDecimal,
              ),
            ),
          if (_notice != null) Text(_notice!),
          if (_busy) const LinearProgressIndicator(),
          PrimaryButton(label: s.save, onPressed: _busy ? null : _save),
        ],
      ),
    );
  }
}
