import 'package:flutter/material.dart';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/providers.dart';
import '../check/check_layout.dart';

/// Reuses the check layout with a blocked preview and no placement action.
class CheckBlockedScreen extends ConsumerWidget {
  const CheckBlockedScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) =>
      CheckLayout(preview: ref.watch(blockedOrderPreviewProvider));
}
