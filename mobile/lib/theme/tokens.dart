import 'package:flutter/material.dart';

export 'risk_tokens.dart';

/// Hand-transcribed from design/tokens.json. Update alongside that source.
abstract final class MetronSpace {
  static const double xs = 4;
  static const double sm = 8;
  static const double md = 12;
  static const double lg = 16;
  static const double xl = 24;
  static const double xxl = 32;
}

abstract final class MetronRadius {
  static const double tag = 6;
  static const double control = 8;
  static const double card = 12;
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
  static const display = TextStyle(
    fontFamily: sans,
    fontSize: 40,
    height: 44 / 40,
    fontWeight: FontWeight.w600,
    letterSpacing: -.8,
  );
  static const title = TextStyle(
    fontFamily: sans,
    fontSize: 28,
    height: 34 / 28,
    fontWeight: FontWeight.w600,
    letterSpacing: -.28,
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
    fontWeight: FontWeight.w400,
    letterSpacing: .96,
  );
  static const figureLarge = TextStyle(
    fontFamily: mono,
    fontSize: 44,
    height: 48 / 44,
    fontWeight: FontWeight.w500,
    letterSpacing: -1.32,
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
