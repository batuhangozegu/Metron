import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../l10n/app_strings.dart';
import '../router.dart';
import '../theme/metron_colors.dart';
import 'fixed_main_action.dart';

/// Owns persistent tabs and the home action, keeping screen work independent.
class AppShell extends StatelessWidget {
  const AppShell({super.key, required this.shell});
  final StatefulNavigationShell shell;

  @override
  Widget build(BuildContext context) => Scaffold(
    body: shell,
    bottomNavigationBar: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (shell.currentIndex == 0)
          FixedMainAction(
            label: AppStrings.checkNewOrder,
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
            items: const [
              BottomNavigationBarItem(
                icon: Icon(Icons.speed_outlined),
                label: AppStrings.risk,
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.format_list_bulleted),
                label: AppStrings.markets,
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.notifications_none),
                label: AppStrings.alerts,
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.tune),
                label: AppStrings.rules,
              ),
            ],
          ),
        ),
      ],
    ),
  );
}
