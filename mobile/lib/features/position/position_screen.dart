import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';
import '../../widgets/num.dart';

/// TODO(screen): Replace using MobilePosition.dc.html. Own only features/position/.
class PositionScreen extends StatelessWidget {
  const PositionScreen({super.key, required this.positionId});
  final String positionId;
  @override
  Widget build(BuildContext context) =>
      FeaturePlaceholder(title: AppStrings.position, detail: Num(positionId));
}
