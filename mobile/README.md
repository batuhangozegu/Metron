# Metron mobile

Flutter foundation for risk monitoring and pre-trade checks. Screens are TODO placeholders. Dark is default; pass `ThemeMode.light` to `MetronApp` for light.

## Run

Use Flutter 3.47.1 / Dart 3.13.1 or a compatible newer stable SDK. From `mobile/`:

```sh
flutter pub get
flutter run
flutter analyze
flutter test
```

Android and iOS runners use `dev.metron.metron`. iOS builds require macOS/Xcode.

## Layout and ownership

- `lib/main.dart`, `app.dart`: ProviderScope and app setup.
- `lib/router.dart`: named routes and four persistent tab branches.
- `lib/theme/`: hand-transcribed design tokens, dark/light themes and the `MetronColors` extension (`context.colors`).
- `lib/widgets/`: Measure, RiskBadge, PnlText, Num, MetronCard, SectionHeader, PrimaryButton, SecondaryButton and FixedMainAction. Import individually or through `widgets.dart`.
- `lib/data/`: immutable plain Dart models, fixtures and read-only Riverpod providers; no networking.
- `lib/l10n/app_strings.dart`: shared English translation catalog.
- `lib/features/<name>/<name>_screen.dart`: replaceable screen placeholders.
- `assets/fonts/`: Inter 400/500/600 and JetBrains Mono 400/500, with OFL licenses. Runtime fetching is disabled.
- `test/`: widget, layout, fixture and navigation tests.

**Each screen owns only `lib/features/<name>/`.** Replace its screen and add helpers, feature copy and tests inside that folder; preserve its public constructor. Coordinate shared changes with the foundation owner. Run feature-local tests explicitly: `flutter test lib/features/<name>/`.

Tab screens supply their own Scaffold, app bar and body. The shared shell owns the bottom bar and home fixed action. Check, blocked-check, position and sign-in supply their own Scaffold and actions, without a tab bar. Rules are read-only.

## Routes

Use `AppRoutes` constants with `pushNamed` for details and `goNamed` for tabs:

| Name | Path | Screen input |
| --- | --- | --- |
| `home` | `/` | — |
| `markets` | `/markets` | — |
| `alerts` | `/alerts` | — |
| `rules` | `/rules` | — |
| `check` | `/check` | Optional `pair` query |
| `check-blocked` | `/check/blocked` | — |
| `position` | `/positions/:id` | Required `positionId` constructor argument |
| `sign-in` | `/sign-in` | — |

Example: `context.pushNamed(AppRoutes.position, pathParameters: {'id': 'btc'})`. Blocked-check back returns to check; tabs retain their state.

## Data and display

Providers expose account, limits, positions (`btc`, `sol`, `eth`), nine prices, alerts, price alarms, class/discipline rules and caution/blocked order previews. `positionProvider(id)` and `coinPriceProvider(pair)` return null for unknown keys. Override providers in ProviderScope for tests.

Fixtures match all eight mobile mockups' renderVals and static markup. UTC dates use 14 November 2026; only BTC has detailed position data. Samples are illustrative, never risk-engine defaults.

Money/quantities are decimal strings; parse only for display, never financial arithmetic. Percentages are percentage points (`2.75` means `2.75%`). Measure takes finite, nonnegative usage and a finite positive limit in the same unit. Caution starts at 75%; exceeding 100% is limit. The fill clamps, while labels retain the actual value. Cooldown uses the explicit fixture verdict.

Use Num for numbers, PnlText for results (arrow/sign/color), and RiskBadge for risk (icon/word/color). PnlText accepts a magnitude formatter and suffix. Num accepts token text styles such as `MetronType.figureLarge`; use bundled font weights. Cards have a one-pixel line and no shadow; buttons are 44 px. Keep shared copy in the translation catalog and feature copy in feature-local translation files.
