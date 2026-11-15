import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileCheckBlocked.dc.html.
/// Own only features/check_blocked/. No place-order action in this state.
class CheckBlockedScreen extends StatelessWidget {
  const CheckBlockedScreen({super.key});
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.checkBlocked);
}
