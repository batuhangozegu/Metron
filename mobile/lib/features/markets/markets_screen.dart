import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileMarkets.dc.html. Own only features/markets/.
class MarketsScreen extends StatelessWidget {
  const MarketsScreen({super.key});
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.markets);
}
