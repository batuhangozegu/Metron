import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileHome.dc.html. Own only features/home/.
class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.risk);
}
