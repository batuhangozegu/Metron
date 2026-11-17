import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:google_fonts/google_fonts.dart';

import 'l10n/locale_provider.dart';
import 'router.dart';
import 'theme/metron_theme.dart';

/// Wrap with ProviderScope. ThemeMode.dark is deliberate; light is opt-in.
class MetronApp extends ConsumerWidget {
  const MetronApp({super.key, this.themeMode = ThemeMode.dark});
  final ThemeMode themeMode;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    GoogleFonts.config.allowRuntimeFetching = false;
    return MaterialApp.router(
      title: ref.watch(stringsProvider).app.appName,
      locale: ref.watch(localeProvider),
      supportedLocales: supportedLocales,
      localizationsDelegates: GlobalMaterialLocalizations.delegates,
      builder: (context, child) => LocalizedCatalogs(child: child!),
      debugShowCheckedModeBanner: false,
      theme: MetronTheme.light,
      darkTheme: MetronTheme.dark,
      themeMode: themeMode,
      routerConfig: ref.watch(routerProvider),
    );
  }
}
