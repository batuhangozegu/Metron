import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'catalog_scope.dart';
import 'number_formatter.dart';
import '../features/alerts/alerts_strings.dart';
import '../features/check/check_strings.dart';
import '../features/home/home_strings.dart';
import '../features/markets/markets_strings.dart';
import '../features/position/position_strings.dart';
import '../features/rules/rules_strings.dart';
import '../features/sign_in/sign_in_strings.dart';
import 'app_strings.dart';

const supportedLocales = [Locale('en'), Locale('tr')];

/// The explicit choice lives for the lifetime of the app's ProviderScope.
final localeProvider = NotifierProvider<LocaleController, Locale>(
  LocaleController.new,
);

class LocaleController extends Notifier<Locale> {
  @override
  Locale build() =>
      resolveDeviceLocale(WidgetsBinding.instance.platformDispatcher.locale);

  void select(Locale locale) {
    state = resolveDeviceLocale(locale);
  }
}

Locale resolveDeviceLocale(Locale device) =>
    device.languageCode == 'tr' ? const Locale('tr') : const Locale('en');

final stringsProvider = Provider((ref) {
  final turkish = ref.watch(localeProvider).languageCode == 'tr';
  return (
    alerts: turkish
        ? const TurkishAlertsStrings() as AlertsStrings
        : const EnglishAlertsStrings(),
    check: turkish
        ? const TurkishCheckStrings() as CheckStrings
        : const EnglishCheckStrings(),
    home: turkish
        ? const TurkishHomeStrings() as HomeStrings
        : const EnglishHomeStrings(),
    markets: turkish
        ? const TurkishMarketsStrings() as MarketsStrings
        : const EnglishMarketsStrings(),
    position: turkish
        ? const TurkishPositionStrings() as PositionStrings
        : const EnglishPositionStrings(),
    rules: turkish
        ? const TurkishRulesStrings() as RulesStrings
        : const EnglishRulesStrings(),
    signIn: turkish
        ? const TurkishSignInStrings() as SignInStrings
        : const EnglishSignInStrings(),
    app: turkish
        ? const TurkishAppStrings() as AppStrings
        : const EnglishAppStrings(),
  );
});

final numberFormatterProvider = Provider(
  (ref) => NumberFormatter(ref.watch(localeProvider).languageCode),
);

/// Allows feature widgets to depend on catalogs without changing widget types.
class LocalizedCatalogs extends ConsumerWidget {
  const LocalizedCatalogs({super.key, required this.child});
  final Widget child;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final strings = ref.watch(stringsProvider);
    return CatalogScope<AlertsStrings>(
      catalog: strings.alerts,
      child: CatalogScope<CheckStrings>(
        catalog: strings.check,
        child: CatalogScope<HomeStrings>(
          catalog: strings.home,
          child: CatalogScope<MarketsStrings>(
            catalog: strings.markets,
            child: CatalogScope<PositionStrings>(
              catalog: strings.position,
              child: CatalogScope<RulesStrings>(
                catalog: strings.rules,
                child: CatalogScope<SignInStrings>(
                  catalog: strings.signIn,
                  child: CatalogScope<AppStrings>(
                    catalog: strings.app,
                    child: CatalogScope<NumberFormatter>(
                      catalog: ref.watch(numberFormatterProvider),
                      child: child,
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
