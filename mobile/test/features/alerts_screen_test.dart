import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/router.dart';

import '../live_test_app.dart';

void main() {
  testWidgets('Alarms are clearly coming soon with no fixture activity', (
    tester,
  ) async {
    final fake = FakeBackend();
    final container = await bootLive(tester, backend: fake);
    container.read(routerProvider).go('/alerts');
    await tester.pumpAndSettle();
    expect(find.text('Price alarms — coming soon'), findsOneWidget);
    expect(
      find.text('Price alarms and alert history are not available yet.'),
      findsOneWidget,
    );
    expect(find.byType(FilledButton), findsNothing);
    expect(fake.requests.any((r) => r.url.path.contains('alarm')), isFalse);
  });
}
