import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'data/api_client.dart';
import 'data/live_models.dart';
import 'data/providers.dart';
import 'widgets/live_state.dart';

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
  final refresh = ValueNotifier(0);
  ref.listen(sessionProvider, (_, next) => refresh.value++);
  final router = createRouter(
    session: () => ref.read(sessionProvider),
    refresh: refresh,
  );
  ref.onDispose(router.dispose);
  ref.onDispose(refresh.dispose);
  return router;
});

/// Detail/check flows sit above the tab shell, as in the mobile mockups.
/// A separate factory supports deep-link and smoke tests without shared state.
GoRouter createRouter({
  String initialLocation = '/',
  AsyncValue<Json?> Function()? session,
  Listenable? refresh,
}) {
  final root = GlobalKey<NavigatorState>();
  return GoRouter(
    navigatorKey: root,
    initialLocation: initialLocation,
    refreshListenable: refresh,
    redirect: (context, state) {
      if (session == null) return null;
      final auth = session();
      final path = state.uri.path;
      if (auth.isLoading || auth.hasError) {
        return path == '/session' ? null : '/session';
      }
      if (auth.value == null) return path == '/sign-in' ? null : '/sign-in';
      return path == '/sign-in' || path == '/session' ? '/' : null;
    },
    errorBuilder: (context, state) =>
        FeaturePlaceholder(title: AppStrings.of(context).notFound),
    routes: [
      GoRoute(
        path: '/session',
        builder: (context, state) => const _SessionScreen(),
      ),
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
                        builder: (context, state) => CheckBlockedScreen(
                          check: state.extra is RiskCheck
                              ? state.extra as RiskCheck
                              : null,
                        ),
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

class _SessionScreen extends ConsumerWidget {
  const _SessionScreen();
  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    body: SafeArea(
      child: Center(
        child: AsyncPanel(
          value: ref.watch(sessionProvider),
          retry: () => ref.invalidate(sessionProvider),
          data: (_) => const SizedBox.shrink(),
        ),
      ),
    ),
  );
}
