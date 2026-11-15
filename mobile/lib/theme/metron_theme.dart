import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'metron_colors.dart';
import 'tokens.dart';

abstract final class MetronTheme {
  static ThemeData get dark => _build(Brightness.dark, MetronColors.dark);
  static ThemeData get light => _build(Brightness.light, MetronColors.light);

  static ThemeData _build(Brightness brightness, MetronColors c) {
    final scheme =
        ColorScheme.fromSeed(
          seedColor: c.text,
          brightness: brightness,
        ).copyWith(
          primary: c.text,
          onPrimary: c.inverse,
          secondary: c.muted,
          surface: c.surface,
          onSurface: c.text,
          outline: c.line,
          error: c.loss,
          onError: c.inverse,
          surfaceTint: Colors.transparent,
        );
    final base = ThemeData(
      useMaterial3: true,
      brightness: brightness,
      colorScheme: scheme,
    );
    // Apply font styles individually: google_fonts' TextTheme helper targets
    // material_ui, whose TextTheme differs from Flutter Material's TextTheme.
    final text = base.textTheme
        .copyWith(
          displayLarge: GoogleFonts.inter(textStyle: MetronType.display),
          displayMedium: GoogleFonts.inter(
            textStyle: base.textTheme.displayMedium,
          ),
          displaySmall: GoogleFonts.inter(
            textStyle: base.textTheme.displaySmall,
          ),
          headlineLarge: GoogleFonts.inter(textStyle: MetronType.title),
          headlineMedium: GoogleFonts.inter(
            textStyle: base.textTheme.headlineMedium,
          ),
          headlineSmall: GoogleFonts.inter(
            textStyle: base.textTheme.headlineSmall,
          ),
          titleLarge: GoogleFonts.inter(textStyle: MetronType.heading),
          titleMedium: GoogleFonts.inter(textStyle: MetronType.heading),
          titleSmall: GoogleFonts.inter(textStyle: base.textTheme.titleSmall),
          bodyLarge: GoogleFonts.inter(textStyle: MetronType.body),
          bodyMedium: GoogleFonts.inter(textStyle: MetronType.body),
          bodySmall: GoogleFonts.inter(textStyle: MetronType.small),
          labelLarge: GoogleFonts.inter(textStyle: base.textTheme.labelLarge),
          labelMedium: GoogleFonts.inter(textStyle: base.textTheme.labelMedium),
          labelSmall: GoogleFonts.inter(textStyle: MetronType.label),
        )
        .apply(bodyColor: c.text, displayColor: c.text);
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(MetronRadius.control),
    );
    final button = ButtonStyle(
      minimumSize: const WidgetStatePropertyAll(Size(0, MetronSize.control)),
      padding: const WidgetStatePropertyAll(
        EdgeInsets.symmetric(horizontal: MetronSpace.lg),
      ),
      shape: WidgetStatePropertyAll(shape),
      elevation: const WidgetStatePropertyAll(0),
      textStyle: WidgetStatePropertyAll(
        GoogleFonts.inter(
          textStyle: text.bodyMedium,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
    return base.copyWith(
      extensions: [c],
      textTheme: text,
      scaffoldBackgroundColor: c.bg,
      splashFactory: NoSplash.splashFactory,
      appBarTheme: AppBarTheme(
        backgroundColor: c.bg,
        foregroundColor: c.text,
        elevation: 0,
        scrolledUnderElevation: 0,
        titleTextStyle: text.titleLarge,
      ),
      cardTheme: CardThemeData(
        color: c.surface,
        elevation: 0,
        shadowColor: Colors.transparent,
        surfaceTintColor: Colors.transparent,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(MetronRadius.card),
          side: BorderSide(color: c.line, width: 1),
        ),
      ),
      dividerTheme: DividerThemeData(color: c.line, thickness: 1, space: 1),
      filledButtonTheme: FilledButtonThemeData(
        style: button.copyWith(
          backgroundColor: WidgetStateProperty.resolveWith(
            (states) =>
                states.contains(WidgetState.disabled) ? c.raised : c.text,
          ),
          foregroundColor: WidgetStateProperty.resolveWith(
            (states) =>
                states.contains(WidgetState.disabled) ? c.subtle : c.inverse,
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: button.copyWith(
          foregroundColor: WidgetStatePropertyAll(c.text),
          side: WidgetStatePropertyAll(BorderSide(color: c.line)),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: c.surface,
        isDense: true,
        constraints: const BoxConstraints(minHeight: MetronSize.control),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: MetronSpace.md,
          vertical: MetronSpace.md,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(MetronRadius.control),
          borderSide: BorderSide(color: c.line),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(MetronRadius.control),
          borderSide: BorderSide(color: c.line),
        ),
      ),
      bottomNavigationBarTheme: BottomNavigationBarThemeData(
        backgroundColor: c.surface,
        selectedItemColor: c.text,
        unselectedItemColor: c.muted,
        elevation: 0,
        type: BottomNavigationBarType.fixed,
        selectedLabelStyle: text.bodySmall,
        unselectedLabelStyle: text.bodySmall,
      ),
    );
  }
}
