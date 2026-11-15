import 'package:flutter/material.dart';

/// Semantic tokens, including neutral surfaces. Access via context.colors.
@immutable
class MetronColors extends ThemeExtension<MetronColors> {
  const MetronColors({
    required this.bg,
    required this.surface,
    required this.raised,
    required this.line,
    required this.text,
    required this.muted,
    required this.subtle,
    required this.inverse,
    required this.profit,
    required this.loss,
    required this.safe,
    required this.caution,
    required this.limit,
    required this.safeTint,
    required this.cautionTint,
    required this.limitTint,
    required this.profitTint,
    required this.lossTint,
  });

  final Color bg, surface, raised, line, text, muted, subtle, inverse;
  final Color profit, loss, safe, caution, limit;
  final Color safeTint, cautionTint, limitTint, profitTint, lossTint;

  static const dark = MetronColors(
    bg: Color(0xFF0D1012),
    surface: Color(0xFF14181B),
    raised: Color(0xFF1B2125),
    line: Color(0xFF2A3238),
    text: Color(0xFFE9EDEF),
    muted: Color(0xFFA3AEB5),
    subtle: Color(0xFF86929A),
    inverse: Color(0xFF0D1012),
    profit: Color(0xFF3FCF8E),
    loss: Color(0xFFF47C7C),
    safe: Color(0xFF6CB6FF),
    caution: Color(0xFFF2B84B),
    limit: Color(0xFFE879F9),
    safeTint: Color.fromRGBO(108, 182, 255, .12),
    cautionTint: Color.fromRGBO(242, 184, 75, .12),
    limitTint: Color.fromRGBO(232, 121, 249, .12),
    profitTint: Color.fromRGBO(63, 207, 142, .14),
    lossTint: Color.fromRGBO(244, 124, 124, .14),
  );
  static const light = MetronColors(
    bg: Color(0xFFF3F5F6),
    surface: Color(0xFFFFFFFF),
    raised: Color(0xFFECEFF1),
    line: Color(0xFFD3D9DD),
    text: Color(0xFF12171A),
    muted: Color(0xFF4A565E),
    subtle: Color(0xFF5F6B73),
    inverse: Color(0xFFFFFFFF),
    profit: Color(0xFF0B7446),
    loss: Color(0xFFBF2F38),
    safe: Color(0xFF1B66B5),
    caution: Color(0xFF8A5A00),
    limit: Color(0xFF9A1FA8),
    safeTint: Color.fromRGBO(27, 102, 181, .10),
    cautionTint: Color.fromRGBO(214, 150, 20, .16),
    limitTint: Color.fromRGBO(154, 31, 168, .10),
    profitTint: Color.fromRGBO(11, 116, 70, .12),
    lossTint: Color.fromRGBO(191, 47, 56, .10),
  );

  @override
  MetronColors copyWith({
    Color? bg,
    Color? surface,
    Color? raised,
    Color? line,
    Color? text,
    Color? muted,
    Color? subtle,
    Color? inverse,
    Color? profit,
    Color? loss,
    Color? safe,
    Color? caution,
    Color? limit,
    Color? safeTint,
    Color? cautionTint,
    Color? limitTint,
    Color? profitTint,
    Color? lossTint,
  }) => MetronColors(
    bg: bg ?? this.bg,
    surface: surface ?? this.surface,
    raised: raised ?? this.raised,
    line: line ?? this.line,
    text: text ?? this.text,
    muted: muted ?? this.muted,
    subtle: subtle ?? this.subtle,
    inverse: inverse ?? this.inverse,
    profit: profit ?? this.profit,
    loss: loss ?? this.loss,
    safe: safe ?? this.safe,
    caution: caution ?? this.caution,
    limit: limit ?? this.limit,
    safeTint: safeTint ?? this.safeTint,
    cautionTint: cautionTint ?? this.cautionTint,
    limitTint: limitTint ?? this.limitTint,
    profitTint: profitTint ?? this.profitTint,
    lossTint: lossTint ?? this.lossTint,
  );

  @override
  MetronColors lerp(covariant MetronColors? other, double t) {
    if (other == null) return this;
    Color mix(Color a, Color b) => Color.lerp(a, b, t)!;
    return MetronColors(
      bg: mix(bg, other.bg),
      surface: mix(surface, other.surface),
      raised: mix(raised, other.raised),
      line: mix(line, other.line),
      text: mix(text, other.text),
      muted: mix(muted, other.muted),
      subtle: mix(subtle, other.subtle),
      inverse: mix(inverse, other.inverse),
      profit: mix(profit, other.profit),
      loss: mix(loss, other.loss),
      safe: mix(safe, other.safe),
      caution: mix(caution, other.caution),
      limit: mix(limit, other.limit),
      safeTint: mix(safeTint, other.safeTint),
      cautionTint: mix(cautionTint, other.cautionTint),
      limitTint: mix(limitTint, other.limitTint),
      profitTint: mix(profitTint, other.profitTint),
      lossTint: mix(lossTint, other.lossTint),
    );
  }
}

extension MetronColorContext on BuildContext {
  MetronColors get colors => Theme.of(this).extension<MetronColors>()!;
}
