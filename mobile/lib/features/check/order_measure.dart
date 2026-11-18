import 'package:flutter/material.dart';

import '../../l10n/number_formatter.dart';
import '../../data/models.dart';
import '../../theme/metron_colors.dart';
import '../../widgets/widgets.dart';
import 'check_strings.dart';

/// A price interval, distinct from the shared usage-versus-limit Measure.
class OrderMeasure extends StatelessWidget {
  const OrderMeasure({super.key, required this.preview});
  final OrderPreview preview;
  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final entry = double.parse(preview.entry);
    final stop = double.parse(preview.stop);
    final target = double.parse(preview.target);
    final fraction = ((entry - stop) / (target - stop)).clamp(0.0, 1.0);
    final large = preview.entry.length < 7;
    Widget price(
      String label,
      String value,
      Widget caption,
      Color color, {
      bool right = false,
    }) => Expanded(
      child: Column(
        crossAxisAlignment: right
            ? CrossAxisAlignment.end
            : CrossAxisAlignment.start,
        spacing: 3,
        children: [
          Text(
            label.toUpperCase(),
            style: TextStyle(fontSize: 11, letterSpacing: .88, color: color),
          ),
          FittedBox(
            fit: BoxFit.scaleDown,
            alignment: right ? Alignment.centerRight : Alignment.centerLeft,
            child: Num(
              NumberFormatter.of(context).decimal(value),
              style: TextStyle(
                fontSize: large ? 22 : 16,
                height: 1.4,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          caption,
        ],
      ),
    );
    Widget detail(String label, Widget value) => Expanded(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        spacing: 3,
        children: [
          Text(label, style: TextStyle(fontSize: 12, color: c.subtle)),
          value,
        ],
      ),
    );
    return MetronCard(
      risk: preview.verdict,
      padding: const EdgeInsets.symmetric(vertical: 18, horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: 16,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              price(
                CheckStrings.of(context).stop,
                preview.stop,
                PnlText(
                  (stop / entry - 1) * 100,
                  decimalPlaces: 1,
                  suffix: '%',
                  style: const TextStyle(fontSize: 12),
                ),
                c.loss,
              ),
              const SizedBox(width: 8),
              price(
                CheckStrings.of(context).entry,
                preview.entry,
                Text(
                  CheckStrings.of(context).limit,
                  style: TextStyle(fontSize: 12, color: c.muted),
                ),
                c.subtle,
              ),
              const SizedBox(width: 8),
              price(
                CheckStrings.of(context).target,
                preview.target,
                PnlText(
                  (target / entry - 1) * 100,
                  decimalPlaces: 1,
                  suffix: '%',
                  style: const TextStyle(fontSize: 12),
                ),
                c.profit,
                right: true,
              ),
            ],
          ),
          Column(
            spacing: 5,
            children: [
              Semantics(
                label:
                    '${CheckStrings.of(context).stop} ${NumberFormatter.of(context).decimal(preview.stop)}, ${CheckStrings.of(context).entry} ${NumberFormatter.of(context).decimal(preview.entry)}, '
                    '${CheckStrings.of(context).target} ${NumberFormatter.of(context).decimal(preview.target)}',
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
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Num(
                    '${NumberFormatter.of(context).number(-1.5, decimalPlaces: 1)} ATR',
                    style: TextStyle(fontSize: 11, color: c.subtle),
                  ),
                  Num(
                    '+${NumberFormatter.of(context).number(3, decimalPlaces: 0)} ATR',
                    style: TextStyle(fontSize: 11, color: c.subtle),
                  ),
                ],
              ),
            ],
          ),
          const Divider(height: 1),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              detail(
                CheckStrings.of(context).stopResult,
                PnlText(
                  double.parse(preview.lossAtStop),
                  style: const TextStyle(fontSize: 16),
                ),
              ),
              const SizedBox(width: 12),
              detail(
                CheckStrings.of(context).targetResult,
                PnlText(
                  double.parse(preview.profitAtTarget),
                  style: const TextStyle(fontSize: 16),
                ),
              ),
            ],
          ),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              detail(
                CheckStrings.of(context).size,
                Wrap(
                  spacing: 4,
                  children: [
                    Num(
                      NumberFormatter.of(context).decimal(preview.quantity),
                      style: const TextStyle(fontSize: 16),
                    ),
                    Text(
                      preview.pair.split('/').first,
                      style: TextStyle(fontSize: 12, color: c.subtle),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              detail(
                CheckStrings.of(context).value,
                Wrap(
                  spacing: 4,
                  children: [
                    Num(
                      NumberFormatter.of(context).decimal(preview.orderValue),
                      style: const TextStyle(fontSize: 16),
                    ),
                    Text(
                      'USDT',
                      style: TextStyle(fontSize: 12, color: c.subtle),
                    ),
                  ],
                ),
              ),
            ],
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
