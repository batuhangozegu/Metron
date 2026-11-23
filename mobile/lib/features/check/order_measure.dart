import 'package:flutter/material.dart';

import '../../data/live_models.dart';
import '../../data/risk_level.dart';
import '../../l10n/live_strings.dart';
import '../../l10n/number_formatter.dart';
import '../../theme/metron_colors.dart';
import '../../widgets/widgets.dart';
import '../../widgets/live_state.dart';
import 'check_strings.dart';
import 'visual_ratio.dart';

/// The existing price ruler, using server sizing and exact decimal labels.
class OrderMeasure extends StatelessWidget {
  const OrderMeasure({super.key, required this.check});
  final RiskCheck check;
  @override
  Widget build(BuildContext context) {
    final c = context.colors, s = LiveStrings.of(context);
    final f = NumberFormatter.of(context), labels = CheckStrings.of(context);
    final stop = decimalTicks(check.value('stop'));
    final entry = decimalTicks(check.value('entry_price'));
    final target = decimalTicks(check.value('target_price'));
    final fraction = visualFraction(
      (entry - stop).abs(),
      (target - stop).abs(),
    );
    final risk = check.blocked
        ? RiskLevel.limit
        : check.json['verdict'] == 'caution'
        ? RiskLevel.caution
        : RiskLevel.safe;
    return MetronCard(
      risk: risk,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: 12,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              for (final item in [
                (labels.stop, 'stop'),
                (labels.entry, 'entry_price'),
                (labels.target, 'target_price'),
              ])
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    spacing: 4,
                    children: [
                      Text(item.$1),
                      FittedBox(
                        alignment: Alignment.centerLeft,
                        fit: BoxFit.scaleDown,
                        child: Num(
                          f.decimal(check.value(item.$2)),
                          style: const TextStyle(fontSize: 18),
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
          Semantics(
            label:
                '${labels.stop} ${check.value('stop')}, ${labels.entry} ${check.value('entry_price')}, ${labels.target} ${check.value('target_price')}',
            child: SizedBox(
              height: 41,
              width: double.infinity,
              child: CustomPaint(
                painter: _PricePainter(
                  fraction,
                  c.loss,
                  c.profit,
                  c.lossTint,
                  c.profitTint,
                  c.text,
                  c.subtle,
                ),
              ),
            ),
          ),
          ValueRow(s.suggestedStop, f.decimal(check.value('suggested_stop'))),
          ValueRow(labels.size, f.decimal(check.value('position_size'))),
          ValueRow(labels.value, '${f.decimal(check.value('notional'))} USDT'),
          ValueRow(s.risk, f.decimal(check.value('actual_risk'))),
          if (check.json['initial_margin'] != null)
            ValueRow(
              s.text('Initial margin', 'Başlangıç marjini'),
              f.decimal(check.value('initial_margin')),
            ),
          if (check.json['liquidation_price'] != null)
            ValueRow(
              s.text('Estimated liquidation', 'Tahmini tasfiye'),
              f.decimal(check.value('liquidation_price')),
            ),
        ],
      ),
    );
  }
}

class _PricePainter extends CustomPainter {
  const _PricePainter(
    this.fraction,
    this.loss,
    this.profit,
    this.lossTint,
    this.profitTint,
    this.text,
    this.tick,
  );
  final double fraction;
  final Color loss, profit, lossTint, profitTint, text, tick;
  @override
  void paint(Canvas canvas, Size size) {
    final p = Paint()..strokeWidth = 1;
    for (var i = 0; i <= 10; i++) {
      final x = .5 + (size.width - 1) * i / 10;
      canvas.drawLine(
        Offset(x, i == 0 || i == 10 ? 0 : 5),
        Offset(x, 12),
        p..color = tick,
      );
    }
    final x = size.width * fraction;
    canvas.drawLine(Offset(x, 0), Offset(x, 12), p..color = text);
    canvas.save();
    canvas.clipRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(0, 17, size.width, 24),
        const Radius.circular(4),
      ),
    );
    canvas.drawRect(Rect.fromLTWH(0, 17, x, 24), p..color = lossTint);
    canvas.drawRect(
      Rect.fromLTWH(x, 17, size.width - x, 24),
      p..color = profitTint,
    );
    canvas.drawRect(Rect.fromLTWH(0, 17, 4, 24), p..color = loss);
    canvas.drawRect(
      Rect.fromLTWH(size.width - 4, 17, 4, 24),
      p..color = profit,
    );
    canvas.drawRect(Rect.fromLTWH(x - 1, 17, 2, 24), p..color = text);
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _PricePainter old) =>
      fraction != old.fraction ||
      loss != old.loss ||
      profit != old.profit ||
      lossTint != old.lossTint ||
      profitTint != old.profitTint ||
      text != old.text ||
      tick != old.tick;
}
