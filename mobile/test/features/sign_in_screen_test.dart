import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:metron/features/sign_in/sign_in_screen.dart';
import 'package:metron/features/sign_in/sign_in_strings.dart';
import 'package:metron/theme/metron_colors.dart';
import 'package:metron/widgets/num.dart';

import '../test_app.dart';

Future<void> enterCredentials(WidgetTester tester) async {
  await tester.enterText(
    find.byKey(const ValueKey('Email')),
    'trader@example.com',
  );
  await tester.enterText(
    find.byKey(const ValueKey('Password')),
    'preview-password',
  );
  await tester.tap(find.text('Sign in'));
  await tester.pumpAndSettle();
}

void main() {
  for (final light in [false, true]) {
    testWidgets(
      'Sign-in and code layout at 390px in ${light ? "light" : "dark"}',
      (tester) async {
        tester.view.physicalSize = const Size(390, 844);
        tester.view.devicePixelRatio = 1;
        addTearDown(tester.view.resetPhysicalSize);
        addTearDown(tester.view.resetDevicePixelRatio);
        await pumpTestApp(tester, const SignInScreen(), light: light);
        expect(find.text('Metron'), findsOneWidget);
        expect(find.text(const EnglishSignInStrings().tagline), findsOneWidget);
        expect(
          find.text(const EnglishSignInStrings().paperNotice),
          findsOneWidget,
        );
        expect(find.byType(BottomNavigationBar), findsNothing);
        expect(
          tester
              .widget<TextField>(find.byKey(const ValueKey('Password')))
              .obscureText,
          isTrue,
        );
        expect(tester.getSize(find.byType(FilledButton)).height, 44);
        expect(tester.getTopLeft(find.byType(TextField).first).dx, 24);
        expect(tester.takeException(), isNull);
        await enterCredentials(tester);
        expect(
          find.text(const EnglishSignInStrings().twoFactor),
          findsOneWidget,
        );
        expect(
          find.text(const EnglishSignInStrings().codeHelp),
          findsOneWidget,
        );
        expect(find.byKey(const ValueKey('Email')), findsNothing);
        expect(find.byType(Num), findsOneWidget);
        final field = tester.widget<TextField>(
          find.byKey(const ValueKey('Code')),
        );
        expect(field.style!.fontFamily, contains('JetBrainsMono'));
        expect(
          field.style!.fontFeatures,
          contains(const FontFeature.tabularFigures()),
        );
        expect(
          field.style!.color,
          light ? MetronColors.light.text : MetronColors.dark.text,
        );
        await tester.enterText(find.byKey(const ValueKey('Code')), '12ab34567');
        expect(field.controller!.text, '123456');
        await tester.tap(find.text('Verify'));
        await tester.pumpAndSettle();
        expect(
          find.text(const EnglishSignInStrings().verified),
          findsOneWidget,
        );
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('Validation, recovery and back preserve local form state', (
    tester,
  ) async {
    await pumpTestApp(tester, const SignInScreen());
    await tester.tap(find.text('Sign in'));
    await tester.pumpAndSettle();
    expect(find.text(const EnglishSignInStrings().emailError), findsOneWidget);
    await tester.enterText(
      find.byKey(const ValueKey('Email')),
      'trader@example.com',
    );
    await tester.tap(find.text('Sign in'));
    await tester.pumpAndSettle();
    expect(
      find.text(const EnglishSignInStrings().passwordError),
      findsOneWidget,
    );
    await enterCredentials(tester);
    await tester.enterText(find.byKey(const ValueKey('Code')), '123');
    await tester.tap(find.text('Verify'));
    await tester.pumpAndSettle();
    expect(find.text(const EnglishSignInStrings().codeError), findsOneWidget);
    await tester.tap(find.text(const EnglishSignInStrings().useRecovery));
    await tester.pumpAndSettle();
    expect(
      find.text(const EnglishSignInStrings().recoveryHelp),
      findsOneWidget,
    );
    await tester.enterText(
      find.byKey(const ValueKey('Recovery code')),
      'backup-code',
    );
    await tester.tap(find.text('Verify'));
    await tester.pumpAndSettle();
    expect(find.text(const EnglishSignInStrings().verified), findsOneWidget);
    await tester.tap(find.text(const EnglishSignInStrings().useAuthenticator));
    await tester.pumpAndSettle();
    expect(find.text(const EnglishSignInStrings().codeHelp), findsOneWidget);
    expect(
      tester
          .widget<TextField>(find.byKey(const ValueKey('Code')))
          .controller!
          .text,
      isEmpty,
    );
    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('Email')), findsOneWidget);
    expect(
      tester
          .widget<TextField>(find.byKey(const ValueKey('Email')))
          .controller!
          .text,
      'trader@example.com',
    );
    expect(
      tester
          .widget<TextField>(find.byKey(const ValueKey('Password')))
          .controller!
          .text,
      isEmpty,
    );
  });

  testWidgets('Form scrolls with keyboard and enlarged text', (tester) async {
    tester.view.physicalSize = const Size(390, 844);
    tester.view.devicePixelRatio = 1;
    tester.view.viewInsets = const FakeViewPadding(bottom: 320);
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    addTearDown(tester.view.resetViewInsets);
    await pumpTestApp(
      tester,
      const MediaQuery(
        data: MediaQueryData(
          size: Size(390, 844),
          viewInsets: EdgeInsets.only(bottom: 320),
          textScaler: TextScaler.linear(1.5),
        ),
        child: SignInScreen(),
      ),
    );
    await tester.ensureVisible(
      find.text(const EnglishSignInStrings().createAccount),
    );
    await tester.tap(find.text(const EnglishSignInStrings().createAccount));
    await tester.pumpAndSettle();
    expect(
      find.text(const EnglishSignInStrings().accountUnavailable),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);
  });
}
