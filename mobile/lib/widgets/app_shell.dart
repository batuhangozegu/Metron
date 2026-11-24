import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/providers.dart';
import '../features/alerts/alert_notifications.dart';
import '../l10n/live_strings.dart';

import '../l10n/app_strings.dart';
import '../router.dart';
import '../theme/metron_colors.dart';
import 'fixed_main_action.dart';

/// Owns persistent tabs and the home action, keeping screen work independent.
class AppShell extends ConsumerWidget {
  const AppShell({super.key, required this.shell});
  final StatefulNavigationShell shell;

  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    appBar: AppBar(
      toolbarHeight: 40,
      title: Text(
        LiveStrings.of(context).account,
        style: const TextStyle(fontSize: 13),
      ),
      actions: [
        TextButton(
          onPressed: () async {
            try {
              await ref.read(sessionProvider.notifier).logout();
            } catch (e) {
              if (context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text(LiveStrings.of(context).error(e))),
                );
              }
            }
          },
          child: Text(LiveStrings.of(context).signOut),
        ),
      ],
    ),
    body: AlertNotifications(child: shell),
    bottomNavigationBar: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (shell.currentIndex == 0)
          FixedMainAction(
            label: AppStrings.of(context).checkNewOrder,
            onPressed: () => context.pushNamed(AppRoutes.check),
          ),
        DecoratedBox(
          position: DecorationPosition.foreground,
          decoration: BoxDecoration(
            border: Border(
              top: BorderSide(color: context.colors.line, width: 1),
            ),
          ),
          child: BottomNavigationBar(
            currentIndex: shell.currentIndex,
            onTap: (index) => shell.goBranch(
              index,
              initialLocation: index == shell.currentIndex,
            ),
            items: [
              BottomNavigationBarItem(
                icon: Icon(Icons.speed_outlined),
                label: AppStrings.of(context).risk,
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.format_list_bulleted),
                label: AppStrings.of(context).markets,
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.notifications_none),
                label: AppStrings.of(context).alerts,
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.tune),
                label: AppStrings.of(context).rules,
              ),
            ],
          ),
        ),
      ],
    ),
  );
}
