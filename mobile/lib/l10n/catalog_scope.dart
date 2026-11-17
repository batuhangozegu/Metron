import 'package:flutter/widgets.dart';

/// Exposes a Riverpod-selected catalog to ordinary and consumer widgets alike.
class CatalogScope<T> extends InheritedWidget {
  const CatalogScope({super.key, required this.catalog, required super.child});

  final T catalog;

  static T? maybeOf<T>(BuildContext context) =>
      context.dependOnInheritedWidgetOfExactType<CatalogScope<T>>()?.catalog;

  @override
  bool updateShouldNotify(CatalogScope<T> oldWidget) =>
      catalog != oldWidget.catalog;
}
