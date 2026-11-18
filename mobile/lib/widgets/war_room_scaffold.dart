import 'package:flutter/material.dart';

import '../theme/metron_colors.dart';
import 'market_ticker.dart';

/// Shared viewport treatment, independent of routes and screen state.
class WarRoomScaffold extends StatelessWidget {
  const WarRoomScaffold({
    super.key,
    required this.body,
    this.appBar,
    this.bottomNavigationBar,
    this.showTicker = true,
  });
  final Widget body;
  final PreferredSizeWidget? appBar;
  final Widget? bottomNavigationBar;
  final bool showTicker;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: showTicker ? null : appBar,
    bottomNavigationBar: bottomNavigationBar,
    body: CustomPaint(
      painter: WarRoomTexture(context.colors),
      child: showTicker
          ? Column(
              children: [
                const SafeArea(bottom: false, child: MarketTicker()),
                if (appBar != null)
                  MediaQuery.removePadding(
                    context: context,
                    removeTop: true,
                    child: appBar!,
                  ),
                Expanded(
                  child: MediaQuery.removePadding(
                    context: context,
                    removeTop: true,
                    child: body,
                  ),
                ),
              ],
            )
          : body,
    ),
  );
}

/// Static, low-opacity paint avoids animation and never intercepts input.
class WarRoomTexture extends CustomPainter {
  const WarRoomTexture(this.colors);
  final MetronColors colors;

  @override
  void paint(Canvas canvas, Size size) {
    final bounds = Offset.zero & size;
    final paint = Paint()
      ..shader = RadialGradient(
        center: const Alignment(-1.1, -.4),
        radius: 1.3,
        colors: [colors.limitEdge.withValues(alpha: .12), Colors.transparent],
      ).createShader(bounds);
    canvas.drawRect(bounds, paint);
    paint.shader = RadialGradient(
      center: const Alignment(1.2, .5),
      radius: 1.1,
      colors: [colors.limitEdge.withValues(alpha: .08), Colors.transparent],
    ).createShader(bounds);
    canvas.drawRect(bounds, paint);
    paint
      ..shader = null
      ..color = colors.text.withValues(alpha: .018)
      ..strokeWidth = .5;
    for (double y = 0; y < size.height; y += 4) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant WarRoomTexture oldDelegate) =>
      colors != oldDelegate.colors;
}
