import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/providers.dart';
import '../../l10n/app_strings.dart';
import '../../l10n/live_strings.dart';
import '../../widgets/widgets.dart';
import 'sign_in_strings.dart';

class SignInScreen extends ConsumerStatefulWidget {
  const SignInScreen({super.key});
  @override
  ConsumerState<SignInScreen> createState() => _SignInScreenState();
}

class _SignInScreenState extends ConsumerState<SignInScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _code = TextEditingController();
  bool _codeStep = false, _register = false, _busy = false;
  String? _error;
  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _code.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_busy) return;
    final s = LiveStrings.of(context);
    if (_codeStep && !RegExp(r'^\d{6}$').hasMatch(_code.text)) {
      setState(() => _error = SignInStrings.of(context).codeError);
      return;
    }
    if (!_codeStep && (!_email.text.contains('@') || _password.text.isEmpty)) {
      setState(
        () => _error = s.text(
          'Enter your email and password.',
          'E-posta ve parolanızı girin.',
        ),
      );
      return;
    }
    if (_register &&
        (_password.text.runes.length < 12 ||
            _password.text.runes.length > 128)) {
      setState(
        () => _error = s.text(
          'Your password must contain 12–128 characters.',
          'Parolanız 12–128 karakter içermelidir.',
        ),
      );
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final auth = ref.read(sessionProvider.notifier);
      if (_codeStep) {
        await auth.complete(_code.text);
        _code.clear();
      } else {
        final signedIn = await auth.login(
          _email.text,
          _password.text,
          register: _register,
        );
        _password.clear();
        if (!signedIn && mounted) setState(() => _codeStep = true);
      }
    } catch (e) {
      if (mounted) {
        setState(
          () => _error = s.error(e, context: _codeStep ? 'code' : 'login'),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = LiveStrings.of(context);
    final f = SignInStrings.of(context);
    return WarRoomScaffold(
      showTicker: false,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: AutofillGroup(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                spacing: 16,
                children: [
                  CommandTitle(AppStrings.of(context).appName),
                  Text(f.tagline),
                  Text(
                    _codeStep
                        ? f.twoFactor
                        : _register
                        ? f.createAccount
                        : AppStrings.of(context).signIn,
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  if (!_codeStep) ...[
                    TextField(
                      key: const ValueKey('email'),
                      controller: _email,
                      autofillHints: const [AutofillHints.email],
                      keyboardType: TextInputType.emailAddress,
                      decoration: InputDecoration(labelText: f.email),
                    ),
                    TextField(
                      key: const ValueKey('password'),
                      controller: _password,
                      obscureText: true,
                      autofillHints: [
                        _register
                            ? AutofillHints.newPassword
                            : AutofillHints.password,
                      ],
                      decoration: InputDecoration(labelText: f.password),
                      onSubmitted: (_) => _submit(),
                    ),
                  ] else ...[
                    Text(f.codeHelp),
                    TextField(
                      key: const ValueKey('code'),
                      controller: _code,
                      autofillHints: const [AutofillHints.oneTimeCode],
                      keyboardType: TextInputType.number,
                      inputFormatters: [
                        FilteringTextInputFormatter.digitsOnly,
                        LengthLimitingTextInputFormatter(6),
                      ],
                      decoration: InputDecoration(labelText: f.code),
                      onSubmitted: (_) => _submit(),
                    ),
                  ],
                  if (_error != null)
                    Text(_error!, key: const ValueKey('auth-error')),
                  if (_busy) const LinearProgressIndicator(),
                  PrimaryButton(
                    label: _codeStep
                        ? f.verify
                        : _register
                        ? f.createAccount
                        : AppStrings.of(context).signIn,
                    onPressed: _busy ? null : _submit,
                  ),
                  TextButton(
                    onPressed: _busy
                        ? null
                        : () => setState(() {
                            if (_codeStep) {
                              _codeStep = false;
                              _code.clear();
                            } else {
                              _register = !_register;
                            }
                            _error = null;
                          }),
                    child: Text(
                      _codeStep
                          ? s.text(
                              'Start sign-in again',
                              'Girişi yeniden başlat',
                            )
                          : _register
                          ? AppStrings.of(context).signIn
                          : f.createAccount,
                    ),
                  ),
                  Text(f.paperNotice),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
