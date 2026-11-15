import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileRules.dc.html. Own only features/rules/.
class RulesScreen extends StatelessWidget {
  const RulesScreen({super.key});
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.rules);
}
