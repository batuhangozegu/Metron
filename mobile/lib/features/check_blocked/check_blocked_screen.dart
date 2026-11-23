import 'package:flutter/material.dart';

import '../../data/live_models.dart';
import '../../l10n/live_strings.dart';
import '../../widgets/widgets.dart';
import '../check/check_layout.dart';

class CheckBlockedScreen extends StatelessWidget {
  const CheckBlockedScreen({super.key, this.check});
  final RiskCheck? check;
  @override
  Widget build(BuildContext context) => WarRoomScaffold(
    appBar: AppBar(title: CommandTitle(LiveStrings.of(context).blocked)),
    body: ListView(
      padding: const EdgeInsets.all(16),
      children: [
        if (check == null)
          Text(LiveStrings.of(context).noCheck)
        else
          CheckLayout(check: check!),
      ],
    ),
  );
}
