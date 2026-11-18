import 'package:flutter/material.dart';

export 'risk_tokens.dart';

/// Shared spacing and the web war-room presentation tokens.
abstract final class MetronSpace {
  static const double xs = 4;
  static const double sm = 8;
  static const double md = 12;
  static const double lg = 16;
  static const double xl = 24;
  static const double xxl = 32;
}

abstract final class MetronRadius {
  static const double tag = 1;
  static const double control = 3;
  static const double card = 3;
}

abstract final class MetronSize {
  static const double control = 44;
  static const double cardPadding = 24;
  static const double pageMaxWidth = 1280;
  static const double pagePadding = 32;
}

/// Letter spacing in the JSON is in em; Flutter expects logical pixels.
abstract final class MetronType {
  static const sans = 'Inter';
  static const mono = 'JetBrains Mono';
  static const condensed = 'Oswald';
  static const display = TextStyle(
    fontFamily: condensed,
    fontSize: 40,
    height: 44 / 40,
    fontWeight: FontWeight.w700,
    letterSpacing: -.8,
  );
  static const title = TextStyle(
    fontFamily: condensed,
    fontSize: 32,
    height: 1.12,
    fontWeight: FontWeight.w700,
    letterSpacing: -.8,
  );
  static const heading = TextStyle(
    fontFamily: sans,
    fontSize: 16,
    height: 22 / 16,
    fontWeight: FontWeight.w600,
    letterSpacing: 0,
  );
  static const body = TextStyle(
    fontFamily: sans,
    fontSize: 15,
    height: 22 / 15,
    fontWeight: FontWeight.w400,
    letterSpacing: 0,
  );
  static const small = TextStyle(
    fontFamily: sans,
    fontSize: 13,
    height: 19 / 13,
    fontWeight: FontWeight.w400,
    letterSpacing: 0,
  );
  // Callers apply uppercase to label copy (for example SectionHeader).
  static const label = TextStyle(
    fontFamily: sans,
    fontSize: 12,
    height: 16 / 12,
    fontWeight: FontWeight.w600,
    letterSpacing: 1.1,
  );
  static const figureLarge = TextStyle(
    fontFamily: mono,
    fontSize: 44,
    height: 48 / 44,
    fontWeight: FontWeight.w500,
    letterSpacing: -1.32,
    fontFeatures: [FontFeature.tabularFigures()],
  );
  static const gauge = TextStyle(
    fontFamily: condensed,
    fontSize: 56,
    height: 1.1,
    fontWeight: FontWeight.w700,
    letterSpacing: -1.4,
    fontFeatures: [FontFeature.tabularFigures()],
  );
  static const figureMedium = TextStyle(
    fontFamily: mono,
    fontSize: 26,
    height: 32 / 26,
    fontWeight: FontWeight.w500,
    letterSpacing: -.52,
    fontFeatures: [FontFeature.tabularFigures()],
  );
  static const figure = TextStyle(
    fontFamily: mono,
    fontSize: 14,
    height: 20 / 14,
    fontWeight: FontWeight.w400,
    letterSpacing: 0,
    fontFeatures: [FontFeature.tabularFigures()],
  );
}
