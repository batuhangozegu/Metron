import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../l10n/live_strings.dart';
import '../../widgets/widgets.dart';

class AlertsScreen extends StatelessWidget {
  const AlertsScreen({super.key});
  @override
  Widget build(BuildContext context) => WarRoomScaffold(
    appBar: AppBar(title: CommandTitle(AppStrings.of(context).alerts)),
    body: ListView(
      padding: const EdgeInsets.all(16),
      children: [
        MetronCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            spacing: 12,
            children: [
              Text(LiveStrings.of(context).comingSoon),
              Text(LiveStrings.of(context).alarmsNote),
            ],
          ),
        ),
      ],
    ),
  );
}
