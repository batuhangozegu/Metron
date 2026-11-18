import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../theme/risk_tokens.dart';

/// Shared segment geometry. Proposed exposure stays hatched, not solid.
void paintSegmentedBar(
  Canvas canvas,
  Rect bounds, {
  required double fraction,
  required Color track,
  required Color color,
  Color? startColor,
  double? solidFraction,
}) {
  final segments = Path();
  for (double x = bounds.left; x < bounds.right; x += MetronRisk.segmentPitch) {
    segments.addRect(
      Rect.fromLTRB(
        x,
        bounds.top,
        math.min(
          x + MetronRisk.segmentPitch - MetronRisk.segmentGap,
          bounds.right,
        ),
        bounds.bottom,
      ),
    );
  }
  final paint = Paint()..color = track;
  canvas.save();
  canvas.clipPath(segments);
  canvas.drawRect(bounds, paint);
  final filled = Rect.fromLTWH(
    bounds.left,
    bounds.top,
    bounds.width * fraction.clamp(0, 1),
    bounds.height,
  );
  paint.shader = LinearGradient(colors: [startColor ?? color, color])
      .createShader(bounds);
  canvas.save();
  if (solidFraction != null) {
    canvas.clipRect(
      Rect.fromLTWH(
        bounds.left,
        bounds.top,
        bounds.width * solidFraction.clamp(0, 1),
        bounds.height,
      ),
    );
  }
  canvas.drawRect(filled, paint);
  canvas.restore();
  if (solidFraction != null) {
    canvas.clipRect(
      Rect.fromLTRB(
        bounds.left + bounds.width * solidFraction.clamp(0, 1),
        bounds.top,
        filled.right,
        bounds.bottom,
      ),
    );
    paint.strokeWidth = 2;
    for (double x = bounds.left; x < bounds.right + bounds.height; x += 5) {
      canvas.drawLine(
        Offset(x, bounds.top),
        Offset(x - bounds.height, bounds.bottom),
        paint,
      );
    }
  }
  canvas.restore();
}
