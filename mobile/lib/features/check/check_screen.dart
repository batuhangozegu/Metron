import 'package:flutter/material.dart';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/providers.dart';
import 'check_layout.dart';

/// Pre-trade preview backed by the shared sample providers.
class CheckScreen extends ConsumerWidget {
  const CheckScreen({super.key, this.pair});
  final String? pair;
  @override
  Widget build(BuildContext context, WidgetRef ref) => CheckLayout(
    preview: ref.watch(orderPreviewProvider),
    requestedPair: pair,
  );
}
