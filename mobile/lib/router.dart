import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'features/alerts/alerts_screen.dart';
import 'features/check/check_screen.dart';
import 'features/check_blocked/check_blocked_screen.dart';
import 'features/home/home_screen.dart';
import 'features/markets/markets_screen.dart';
import 'features/position/position_screen.dart';
import 'features/rules/rules_screen.dart';
import 'features/sign_in/sign_in_screen.dart';
import 'l10n/app_strings.dart';
import 'widgets/app_shell.dart';
import 'widgets/feature_placeholder.dart';

/// Stable route names; prefer named navigation to constructing paths.
abstract final class AppRoutes {
  static const home = 'home';
  static const markets = 'markets';
  static const alerts = 'alerts';
  static const rules = 'rules';
  static const check = 'check';
  static const checkBlocked = 'check-blocked';
  static const position = 'position';
  static const signIn = 'sign-in';
}

final routerProvider = Provider<GoRouter>((ref) {
  final router = createRouter();
  ref.onDispose(router.dispose);
  return router;
});

/// Detail/check flows sit above the tab shell, as in the mobile mockups.
/// A separate factory supports deep-link and smoke tests without shared state.
GoRouter createRouter({String initialLocation = '/'}) {
  final root = GlobalKey<NavigatorState>();
  return GoRouter(
    navigatorKey: root,
    initialLocation: initialLocation,
    errorBuilder: (context, state) =>
        const FeaturePlaceholder(title: AppStrings.notFound),
    routes: [
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) => AppShell(shell: shell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/',
                name: AppRoutes.home,
                builder: (context, state) => const HomeScreen(),
                routes: [
                  GoRoute(
                    path: 'check',
                    name: AppRoutes.check,
                    parentNavigatorKey: root,
                    builder: (context, state) =>
                        CheckScreen(pair: state.uri.queryParameters['pair']),
                    routes: [
                      GoRoute(
                        path: 'blocked',
                        name: AppRoutes.checkBlocked,
                        parentNavigatorKey: root,
                        builder: (context, state) => const CheckBlockedScreen(),
                      ),
                    ],
                  ),
                  GoRoute(
                    path: 'positions/:id',
                    name: AppRoutes.position,
                    parentNavigatorKey: root,
                    builder: (context, state) =>
                        PositionScreen(positionId: state.pathParameters['id']!),
                  ),
                ],
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/markets',
                name: AppRoutes.markets,
                builder: (context, state) => const MarketsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/alerts',
                name: AppRoutes.alerts,
                builder: (context, state) => const AlertsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/rules',
                name: AppRoutes.rules,
                builder: (context, state) => const RulesScreen(),
              ),
            ],
          ),
        ],
      ),
      GoRoute(
        path: '/sign-in',
        name: AppRoutes.signIn,
        builder: (context, state) => const SignInScreen(),
      ),
    ],
  );
}
