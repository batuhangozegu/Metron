import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileCheck.dc.html. Own only features/check/.
class CheckScreen extends StatelessWidget {
  const CheckScreen({super.key, this.pair});
  final String? pair;
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.check);
}
