import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../l10n/app_strings.dart';
import '../../theme/metron_colors.dart';
import '../../theme/tokens.dart';
import '../../widgets/metron_button.dart';
import '../../widgets/num.dart';
import '../../widgets/command_title.dart';
import '../../widgets/section_header.dart';
import '../../widgets/war_room_scaffold.dart';
import 'sign_in_strings.dart';

/// Local-only preview. Credentials are never transmitted or persisted.
class SignInScreen extends StatefulWidget {
  const SignInScreen({super.key});

  @override
  State<SignInScreen> createState() => _SignInScreenState();
}

class _SignInScreenState extends State<SignInScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _code = TextEditingController();
  bool _codeStep = false;
  bool _recovery = false;
  bool _verified = false;
  String? _error;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _code.dispose();
    super.dispose();
  }

  void _submit() {
    FocusScope.of(context).unfocus();
    setState(() {
      if (!_codeStep) {
        if (!RegExp(r'^[^\s@]+@[^\s@]+\.[^\s@]+$')
            .hasMatch(_email.text.trim())) {
          _error = SignInStrings.of(context).emailError;
        } else if (_password.text.isEmpty) {
          _error = SignInStrings.of(context).passwordError;
        } else {
          _error = null;
          _password.clear();
          _codeStep = true;
        }
      } else if (_recovery
          ? _code.text.trim().isEmpty
          : !RegExp(r'^\d{6}$').hasMatch(_code.text)) {
        _error = _recovery
            ? SignInStrings.of(context).recoveryError
            : SignInStrings.of(context).codeError;
      } else {
        _error = null;
        _verified = true;
        _code.clear();
      }
    });
  }

  void _resetStep() {
    setState(() {
      _codeStep = false;
      _recovery = false;
      _verified = false;
      _error = null;
      _code.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return PopScope(
      canPop: !_codeStep,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop && _codeStep) _resetStep();
      },
      child: WarRoomScaffold(
        showTicker: false,
        body: SafeArea(
          child: SingleChildScrollView(
            keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
            padding: EdgeInsets.fromLTRB(
              24,
              (72 - MediaQuery.paddingOf(context).top).clamp(24, 72),
              24,
              32,
            ),
            child: AutofillGroup(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Align(
                    alignment: Alignment.centerLeft,
                    child: ExcludeSemantics(
                      child: CustomPaint(
                        size: const Size(44, 44),
                        painter: _MetronMark(colors.text),
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  SectionHeader(AppStrings.of(context).warRoom),
                  const SizedBox(height: 8),
                  CommandTitle(
                    AppStrings.of(context).appName,
                    style: MetronType.title.copyWith(
                      color: colors.text,
                      letterSpacing: -.56,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    SignInStrings.of(context).tagline,
                    style: MetronType.body.copyWith(color: colors.muted),
                  ),
                  const SizedBox(height: 40),
                  Semantics(
                    label: _codeStep
                        ? SignInStrings.of(context).twoFactor
                        : AppStrings.of(context).signIn,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: _codeStep
                          ? _codeForm(context)
                          : _signInForm(context),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  List<Widget> _signInForm(BuildContext context) => [
    _field(
      context,
      label: SignInStrings.of(context).email,
      controller: _email,
      keyboard: TextInputType.emailAddress,
      autofill: const [AutofillHints.email],
      next: true,
    ),
    const SizedBox(height: 16),
    _field(
      context,
      label: SignInStrings.of(context).password,
      controller: _password,
      obscure: true,
      autofill: const [AutofillHints.password],
    ),
    ..._feedback(context),
    const SizedBox(height: 24),
    PrimaryButton(label: AppStrings.of(context).signIn, onPressed: _submit),
    const SizedBox(height: 16),
    SecondaryButton(
      label: SignInStrings.of(context).createAccount,
      onPressed: () {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(SignInStrings.of(context).accountUnavailable)),
        );
      },
    ),
    const SizedBox(height: 16),
    Text(
      SignInStrings.of(context).paperNotice,
      textAlign: TextAlign.center,
      style: MetronType.small.copyWith(color: context.colors.subtle),
    ),
  ];

  List<Widget> _codeForm(BuildContext context) => [
    Text(
      _recovery
          ? SignInStrings.of(context).recovery
          : SignInStrings.of(context).twoFactor,
      style: MetronType.heading.copyWith(
        fontSize: 20,
        height: 26 / 20,
        color: context.colors.text,
      ),
    ),
    const SizedBox(height: 6),
    Text(
      _recovery
          ? SignInStrings.of(context).recoveryHelp
          : SignInStrings.of(context).codeHelp,
      style: MetronType.body.copyWith(
        fontSize: 14,
        height: 20 / 14,
        color: context.colors.muted,
      ),
    ),
    const SizedBox(height: 16),
    _field(
      context,
      label: _recovery
          ? SignInStrings.of(context).recovery
          : SignInStrings.of(context).code,
      controller: _code,
      numeric: !_recovery,
      keyboard: _recovery ? TextInputType.text : TextInputType.number,
      autofill: _recovery ? null : const [AutofillHints.oneTimeCode],
    ),
    ..._feedback(context),
    const SizedBox(height: 24),
    PrimaryButton(label: SignInStrings.of(context).verify, onPressed: _submit),
    const SizedBox(height: 16),
    SizedBox(
      height: MetronSize.control,
      child: TextButton(
        onPressed: () {
          FocusScope.of(context).unfocus();
          setState(() {
            _recovery = !_recovery;
            _code.clear();
            _error = null;
            _verified = false;
          });
        },
        style: TextButton.styleFrom(
          foregroundColor: context.colors.muted,
          textStyle: MetronType.body.copyWith(
            fontWeight: FontWeight.w500,
            decoration: TextDecoration.underline,
          ),
        ),
        child: Text(
          _recovery
              ? SignInStrings.of(context).useAuthenticator
              : SignInStrings.of(context).useRecovery,
        ),
      ),
    ),
  ];

  List<Widget> _feedback(BuildContext context) => [
    if (_error != null || _verified) ...[
      const SizedBox(height: 8),
      Semantics(
        liveRegion: true,
        child: Text(
          _error ?? SignInStrings.of(context).verified,
          style: MetronType.small.copyWith(color: context.colors.muted),
        ),
      ),
    ],
  ];

  Widget _field(
    BuildContext context, {
    required String label,
    required TextEditingController controller,
    TextInputType? keyboard,
    List<String>? autofill,
    bool obscure = false,
    bool numeric = false,
    bool next = false,
  }) {
    final colors = context.colors;
    // EditableText needs a style, so take it from the shared Num renderer.
    final numericStyle = (Num(
      '',
      style: TextStyle(
        fontSize: 30,
        height: 38 / 30,
        letterSpacing: 10.5,
        color: colors.text,
      ),
    ).build(context) as Text).style;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(label, style: MetronType.small.copyWith(color: colors.muted)),
        const SizedBox(height: 6),
        Semantics(
          label: label,
          child: TextField(
            key: ValueKey(label),
            controller: controller,
            obscureText: obscure,
            keyboardType: keyboard,
            autofillHints: autofill,
            autocorrect: false,
            enableSuggestions: false,
            textInputAction: next ? TextInputAction.next : TextInputAction.done,
            onSubmitted: next ? null : (_) => _submit(),
            onChanged: (_) {
              if (_error != null || _verified) {
                setState(() {
                  _error = null;
                  _verified = false;
                });
              }
            },
            inputFormatters: numeric
                ? [
                    FilteringTextInputFormatter.digitsOnly,
                    LengthLimitingTextInputFormatter(6),
                  ]
                : null,
            textAlign: numeric ? TextAlign.center : TextAlign.start,
            style: numeric
                ? numericStyle
                : MetronType.body.copyWith(fontSize: 16, color: colors.text),
            decoration: InputDecoration(
              hint: numeric
                  ? Num(
                      '000000',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 30,
                        height: 38 / 30,
                        letterSpacing: 10.5,
                        color: colors.subtle,
                      ),
                    )
                  : null,
              contentPadding: EdgeInsets.symmetric(
                horizontal: 14,
                vertical: numeric ? 12 : 10,
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(MetronRadius.control),
                borderSide: BorderSide(color: colors.muted),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

/// Neutral ruler mark from the mockup's 22-unit viewbox.
class _MetronMark extends CustomPainter {
  const _MetronMark(this.color);
  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.scale(size.width / 22, size.height / 22);
    final paint = Paint()
      ..color = color
      ..strokeWidth = 1.3
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(const Offset(2, 16), const Offset(20, 16), paint);
    for (final x in [2.0, 6.5, 11.0, 15.5, 20.0]) {
      canvas.drawLine(
        Offset(x, 16),
        Offset(x, x == 6.5 || x == 15.5 ? 12 : 7),
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(_MetronMark oldDelegate) => oldDelegate.color != color;
}
