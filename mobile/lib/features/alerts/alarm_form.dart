import 'package:flutter/material.dart';

import '../../data/models.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/widgets.dart';
import 'alerts_strings.dart';

class AlarmForm extends StatefulWidget {
  const AlarmForm({super.key, required this.prices});
  final List<CoinPrice> prices;
  @override
  State<AlarmForm> createState() => _AlarmFormState();
}

class _AlarmFormState extends State<AlarmForm> {
  final _form = GlobalKey<FormState>();
  final _price = TextEditingController();
  late String? _pair = widget.prices.firstOrNull?.pair;
  var _direction = AlarmDirection.above;
  @override
  void dispose() {
    _price.dispose();
    super.dispose();
  }

  void _save() {
    if (!_form.currentState!.validate() || _pair == null) return;
    Navigator.of(context).pop(
      PriceAlarm(
        id: 'local-${DateTime.now().microsecondsSinceEpoch}',
        pair: _pair!,
        direction: _direction,
        targetPrice: _price.text.trim(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    backgroundColor: context.colors.surface,
    surfaceTintColor: context.colors.surface,
    elevation: 0,
    shape: RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(MetronRadius.card),
      side: BorderSide(color: context.colors.line),
    ),
    title: Text(AlertsStrings.of(context).addAlarm),
    content: SingleChildScrollView(
      child: Form(
        key: _form,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              AlertsStrings.of(context).sampleOnly,
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 16),
            if (widget.prices.isEmpty)
              Text(AlertsStrings.of(context).noPrices)
            else
              DropdownButtonFormField<String>(
                initialValue: _pair,
                isExpanded: true,
                decoration: InputDecoration(
                  labelText: AlertsStrings.of(context).pair,
                ),
                items: [
                  for (final price in widget.prices)
                    DropdownMenuItem(
                      value: price.pair,
                      child: Text(price.pair),
                    ),
                ],
                onChanged: (value) => setState(() => _pair = value),
              ),
            const SizedBox(height: 12),
            DropdownButtonFormField<AlarmDirection>(
              initialValue: _direction,
              isExpanded: true,
              decoration: InputDecoration(
                labelText: AlertsStrings.of(context).condition,
              ),
              items: [
                for (final direction in AlarmDirection.values)
                  DropdownMenuItem(
                    value: direction,
                    child: Text(AlertsStrings.of(context).direction(direction)),
                  ),
              ],
              onChanged: (value) => setState(() => _direction = value!),
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _price,
              style: MetronType.figure.copyWith(color: context.colors.text),
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              decoration: InputDecoration(
                labelText: AlertsStrings.of(context).targetPrice,
              ),
              validator: (value) {
                final text = value?.trim() ?? '';
                final number = double.tryParse(text);
                return RegExp(r'^\d+(?:\.\d+)?$').hasMatch(text) &&
                        number != null &&
                        number.isFinite &&
                        number > 0
                    ? null
                    : AlertsStrings.of(context).invalidPrice;
              },
            ),
          ],
        ),
      ),
    ),
    actions: [
      SecondaryButton(
        label: AlertsStrings.of(context).cancel,
        onPressed: () => Navigator.of(context).pop(),
      ),
      PrimaryButton(
        label: AlertsStrings.of(context).addAlarm,
        onPressed: _pair == null ? null : _save,
      ),
    ],
  );
}
