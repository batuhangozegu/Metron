import 'package:flutter/material.dart';

import '../../l10n/app_strings.dart';
import '../../widgets/feature_placeholder.dart';

/// TODO(screen): Replace using MobileSignIn.dc.html. Own only features/sign_in/.
class SignInScreen extends StatelessWidget {
  const SignInScreen({super.key});
  @override
  Widget build(BuildContext context) =>
      const FeaturePlaceholder(title: AppStrings.signIn);
}
