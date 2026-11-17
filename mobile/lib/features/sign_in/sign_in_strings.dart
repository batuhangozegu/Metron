import 'package:flutter/widgets.dart';

import '../../l10n/catalog_scope.dart';

/// Every locale must implement all members; no translation falls back silently.
abstract interface class SignInStrings {
  static SignInStrings of(BuildContext context) =>
      CatalogScope.maybeOf<SignInStrings>(context) ??
      const EnglishSignInStrings();
  String get tagline;
  String get email;
  String get password;
  String get createAccount;
  String get paperNotice;
  String get twoFactor;
  String get codeHelp;
  String get code;
  String get verify;
  String get useRecovery;
  String get recovery;
  String get recoveryHelp;
  String get useAuthenticator;
  String get emailError;
  String get passwordError;
  String get codeError;
  String get recoveryError;
  String get verified;
  String get accountUnavailable;
}

final class EnglishSignInStrings implements SignInStrings {
  const EnglishSignInStrings();

  @override
  String get tagline => 'Measure the risk before every order.';
  @override
  String get email => 'Email';
  @override
  String get password => 'Password';
  @override
  String get createAccount => 'Create an account';
  @override
  String get paperNotice =>
      'A new account starts with a paper balance. No real money moves until you connect an exchange.';
  @override
  String get twoFactor => 'Two-factor code';
  @override
  String get codeHelp => 'Enter the six digits from your authenticator app.';
  @override
  String get code => 'Code';
  @override
  String get verify => 'Verify';
  @override
  String get useRecovery => 'Use a recovery code';
  @override
  String get recovery => 'Recovery code';
  @override
  String get recoveryHelp => 'Enter one of your saved recovery codes.';
  @override
  String get useAuthenticator => 'Use an authenticator code';
  @override
  String get emailError => 'Enter a valid email address.';
  @override
  String get passwordError => 'Enter your password.';
  @override
  String get codeError => 'Enter the full six-digit code.';
  @override
  String get recoveryError => 'Enter a recovery code.';
  @override
  String get verified =>
      'Code accepted in this preview. No account has been signed in.';
  @override
  String get accountUnavailable =>
      'Account creation is not available in this preview.';
}

final class TurkishSignInStrings implements SignInStrings {
  const TurkishSignInStrings();
  @override
  String get tagline => 'Her emirden önce riski ölçün.';
  @override
  String get email => 'E-posta';
  @override
  String get password => 'Şifre';
  @override
  String get createAccount => 'Hesap oluştur';
  @override
  String get paperNotice =>
      'Yeni hesap sanal bakiyeyle başlar. Bir borsa bağlayana kadar gerçek para kullanılmaz.';
  @override
  String get twoFactor => 'İki aşamalı doğrulama kodu';
  @override
  String get codeHelp =>
      'Kimlik doğrulama uygulamanızdaki altı haneli kodu girin.';
  @override
  String get code => 'Kod';
  @override
  String get verify => 'Doğrula';
  @override
  String get useRecovery => 'Kurtarma kodu kullan';
  @override
  String get recovery => 'Kurtarma kodu';
  @override
  String get recoveryHelp => 'Kaydettiğiniz kurtarma kodlarından birini girin.';
  @override
  String get useAuthenticator => 'Kimlik doğrulama kodu kullan';
  @override
  String get emailError => 'Geçerli bir e-posta adresi girin.';
  @override
  String get passwordError => 'Şifrenizi girin.';
  @override
  String get codeError => 'Altı haneli kodun tamamını girin.';
  @override
  String get recoveryError => 'Bir kurtarma kodu girin.';
  @override
  String get verified =>
      'Kod bu önizlemede kabul edildi. Hesaba giriş yapılmadı.';
  @override
  String get accountUnavailable => 'Bu önizlemede hesap oluşturulamaz.';
}
