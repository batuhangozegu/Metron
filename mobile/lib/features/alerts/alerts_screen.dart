import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileAlerts.dc.html. Own only features/alerts/.
class AlertsScreen extends StatelessWidget {
  const AlertsScreen({super.key});
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.alerts);
}
