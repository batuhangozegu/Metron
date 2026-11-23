# Metron mobile

Flutter app for the live Metron paper-trading API, with the war-room theme and English/Turkish copy. Supports registration, password/TOTP sign-in, persistent cookie sessions, paper account creation/selection, portfolio and positions, market quotes, pre-trade checks, orders, position close and saved risk limits. Price alarms are marked coming soon until a backend exists.

## Run

Use Flutter 3.47.1 / Dart 3.13.1 or a compatible newer stable SDK. Start the API and dependencies using the root README, apply all migrations, then from `mobile/`:

```sh
flutter pub get
flutter run
```

The default API is `http://10.0.2.2:8080` on Android (emulator host) and `http://127.0.0.1:8080` elsewhere (including the iOS simulator). For a physical phone, use a reachable API address on your LAN or an HTTPS deployment:

```sh
flutter run --dart-define=METRON_API_URL=http://192.168.1.20:8080
# Deployment / release:
flutter run --release --dart-define=METRON_API_URL=https://api.example.com
```

Allow local-network access when iOS asks. Android debug/profile allow local HTTP; release uses HTTPS. The iOS runner permits local networking. Both runners use `dev.metron.metron`; iOS builds require macOS/Xcode. The cookie transport targets native mobile; Flutter Web requires a same-origin cookie proxy and is not supported by this transport.

At launch, `/v1/me` verifies the stored session before protected screens open. Full sessions (including refreshed expiry) use `flutter_secure_storage`; the five-minute pending TOTP cookie is memory-only. Logout/401 removes the full session. Account selection is remembered per user and API URL. HTTP errors and rule results display the API's EN/TR messages, with localized fallback messages for authentication/network failures. Amounts and quantities remain decimal strings through requests and rendering.

Home portfolio/positions and visible market ticker quotes refresh every five seconds. Polling pauses for hidden routes/tabs and while the app is in the background. A pre-trade check uses the current quote; placing an order repeats the server's risk evaluation with the saved account profile and current execution price. Any failed rule routes to the blocked result, including an order rejected after a successful preview.

## Checks

```sh
flutter pub get
dart format --set-exit-if-changed lib test
flutter analyze
flutter test
```

The HTTP tests use `MockClient` and an in-memory secure-store substitute; they do not need a running backend or real credentials. They cover session persistence/refresh/clear, sign-in/TOTP, account selection/creation, check/order/blocked routing, close confirmation, risk-limit validation, polling, loading/error/empty states and localization. Shared widget/format tests remain offline.

## Layout

- `lib/data/api_client.dart`: injectable HTTP client, cookie jar and secure store.
- `lib/data/live_models.dart`, `providers.dart`: live API data and Riverpod session/account state.
- `lib/router.dart`: protected routes and startup session gate.
- `lib/features/`: live screens and their existing translation catalogs.
- `lib/l10n/live_strings.dart`: shared EN/TR copy for the API flows; Rules retains the language selector.
- `lib/widgets/`: war-room scaffolds, live ticker, account gate, async states and existing visual primitives.
- `assets/fonts/`: bundled fonts and licenses; runtime fetching is disabled.
