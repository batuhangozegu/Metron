import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:metron/theme/metron_theme.dart';

Future<void> pumpTestApp(
  WidgetTester tester,
  Widget child, {
  bool light = false,
}) async {
  GoogleFonts.config.allowRuntimeFetching = false;
  await tester.pumpWidget(
    MaterialApp(
      theme: light ? MetronTheme.light : MetronTheme.dark,
      home: Scaffold(body: child),
    ),
  );
  await tester.runAsync(() async {
    await GoogleFonts.pendingFonts();
  });
  await tester.pumpAndSettle();
}
