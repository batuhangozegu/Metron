# Metron web foundation

Next.js App Router, Tailwind v4, React Server Components by default. The root layout loads Inter and JetBrains Mono, initializes the persisted theme before paint, and renders `AppShell` once. Pages provide their content inside its `<main>`.

## Sign-in and accounts

`/sign-in`, `/register` and `/markets` are public. Other app pages check `/v1/me` in a client guard before displaying content. The API authorizes every private data request. Session checks run on navigation and window focus; the UI shows a retry action when the service is unavailable.

The `/v1/*` route handler forwards requests to `METRON_API_URL` at runtime, defaulting to `http://127.0.0.1:8080`. Compose already supplies `http://api:8080`; no deployment changes are needed. Set this server-side environment variable when using another API address. The proxy preserves separate full/pending `Set-Cookie` headers, refreshed cookies, JSON content types, and `Retry-After`, and disables caching. Browser requests always use the web origin. Use the API's `COOKIE_SECURE=true` setting when serving over HTTPS.

`/settings` manages authenticator setup, enable/disable and sign-out. QR codes are generated locally with `qrcode.react`; setup secrets are kept only in component memory. Authenticator codes cannot be reused after a successful operation. `/accounts` lists user-owned accounts and creates Spot or Futures paper accounts using exact decimal strings. Accounts without a market field display as Spot. Balance entry uses a decimal point in either interface language; displayed balances use locale-aware formatting and retain up to eight fractional digits.

Run `npm run lint`, `npx tsc --noEmit`, `node scripts/verify-foundation.mjs`, and `npm run build`. The verification script includes auth request/proxy cookie checks, localized errors and exact balance validation alongside the existing foundation checks.

## Paper trading

`/check` loads the signed-in user's USDT paper accounts and `/v1/markets`. Spot accounts permit long orders only; futures accounts expose both directions and integer leverage from 1 to 125. An optional manual stop uses a decimal point in either language. The preview refreshes the ticker and sends its exact price to `/v1/risk/check`; editing the form clears the result. The API's bilingual rule list supplies every verdict, and any failed rule disables opening an order.

Opening an order posts only account, symbol, side, optional stop and futures leverage to `/v1/orders`. Go prices the fill and checks risk again. A 422 response replaces the preview with the returned blocked check; success opens `/portfolio` with that account selected. No mutation is automatically retried.

`/portfolio` reads open and closed positions for the selected paper account. Both lists refresh every five seconds while the tab is visible and when visibility returns. Closing posts to `/v1/positions/{id}/close`, then refreshes both lists; an already-closed response also triggers a refresh. Switching accounts cancels old reads. Missing quotes show an unavailable state, and close timestamps use UTC. Amounts stay decimal strings, including quantities beyond JavaScript's integer precision. Only bounded meter geometry uses numbers. The equity chart remains explicitly labeled sample data until step 6.

Run `npm run lint`, `npm run typecheck`, `node scripts/verify-foundation.mjs`, and `npm run build`. The verification script covers live check rendering, fail-closed order controls, one status per rule, retained 422 payloads, localized open/closed/empty position tables, exact decimal quantities and UTC timestamps. These checks use contract fixtures; testing fills and stop execution against Go and the engine requires the other step 4 lanes.

## Ownership and folders

Each screen agent owns **only its route folder and its English message file**:

| Screen | Route folder | Messages |
| --- | --- | --- |
| Overview | `src/app/` (`page.tsx` and screen-specific files only) | `src/messages/en/overview.json` |
| Pre-trade check | `src/app/check/` | `src/messages/en/check.json` |
| Markets | `src/app/markets/` | `src/messages/en/markets.json` |
| Portfolio | `src/app/portfolio/` | `src/messages/en/portfolio.json` |
| Rules | `src/app/rules/` | `src/messages/en/rules.json` |

The overview agent must leave `layout.tsx` and `globals.css` alone. Keep route-specific components beside their page. Shared code lives in `src/components/metron/`, existing primitives in `src/components/ui/`, and shared helpers/fixtures in `src/lib/`. Do not change shared files when implementing a screen. Consult the installed Next.js docs in `node_modules/next/dist/docs/` when needed.

## Shared component API

Import from individual files or `@/components/metron`. All props/types are exported. These components work on the server; only navigation and the theme toggle need a client boundary.

| Component | Props |
| --- | --- |
| `AppShell` | `children`; already applied by the layout |
| `Measure` | `value`, `limit`: numbers in the same real unit; `added = 0`; `cautionAt = 0.75`; `size = "md"` (`"md"` or `"sm"`); `showMarks = true`; `showStatus = true` (disable when a row already has a verdict); optional `label`, `className`, `formatValue(decimalString)` |
| `RiskBadge` | `state: "safe" \| "caution" \| "limit"`, optional `className`; translated icon/word/color |
| `PnL` | `value: string \| bigint`, `decimals = 2`, `suffix = ""`, optional `className`; arrow/sign and semantic color; zero is neutral |
| `Card` | Standard `<article>` props; add `className="p-card"` for a padded card |
| `CardHeader` | `title`, optional `note`, `action`, `children`, and standard `<div>` props |
| `Stat` | `label`, `figure`, optional `note`, and standard `<div>` props |
| `ClassTag` | `coinClass: "core" \| "large" \| "small" \| "stable"`, optional `className` |

`riskState(value, limit, cautionAt = 0.75)` is exported from `@/lib/risk` and the component barrel. Caution includes the threshold; limit means strictly greater than capacity. Zero/invalid capacity fails closed. `Measure` uses current plus added exposure for its state, clips geometry at the limit, and retains the actual quantity in accessible text. Use a sibling `RiskBadge` wherever a visible risk verdict is needed. Discipline verdicts belong to rule logic. The mockup's open-position check says Safe at 4/5; the generic risk function consistently returns Caution at 75% and above.

Use `formatValue={formatPercent}` for percentage rulers, or `(value) => formatAmount(value, 0)` for counts. `value` is the existing quantity, `added` is an increment, not the final total. Pass translated copy to `label`, `title`, `note`, and other content props. Wrap any numbers inside mixed prose in `<span className="num">`.

## Theme, numbers, and sample data

`globals.css` is generated from `../design/tokens.json` with `node scripts/generate-theme.mjs`. Both palettes, typography, spacing and radii are exposed through Tailwind: `bg-surface`, `bg-raised`, `text-muted`, `border-line`, `text-profit`, `bg-caution-tint`, `rounded-card`, `p-card`, `p-page`, `max-w-page`, `text-title`, and `text-figure-large`. Use `num` for every number. Add `uppercase` to `text-label`. Colors belong only in `globals.css`. Existing UI primitives use raised backgrounds for muted surfaces because Metron's `muted` token is a text color. Controls have a 44 px minimum target.

`formatAmount(decimal, decimals = 2)`, `formatPercent(decimal, decimals = 2)` and `formatPrice(decimal, decimals?)` accept strings or bigint, group digits, and round exactly using decimal digits/BigInt. Percent inputs are percentage points (`"2.75"` → `2.75%`). Price preserves source precision by default. `decimalProduct(a, b)` is an exact decimal display helper. Keep money as decimal strings; do not calculate displayed monetary results using floating point. Number ratios inside the ruler are only geometry.

`@/lib/sample-data` exports typed `account`, `limits`, `openPositions`, `gauges`, `markets`, `classRules`, `disciplineRules`, `journal`, `classStatistics`, `holdings`, `portfolioStatistics`, `equitySeries`, `preTradeCaution`, `preTradeBlocked`, and the aggregate `sampleData`. Source display fields are copied exactly; use `plAmount`, `changePercent`, and `netAmount` with formatting components. Color fields are semantic names, so map them to complete Tailwind class strings rather than constructing class names dynamically. Equity series contains the exact exported SVG coordinates; the mockup has no underlying dated money series to recover. Sample copy lives in `common.json`. Optional regeneration: `node scripts/generate-sample-data.mjs` (reads only the trusted repository mockups).

## Messages and checks

Add flat string entries to your screen JSON, then use `const t = getMessages("overview")` from `@/lib/messages`. `t("greeting", { name: "…" })` replaces `{name}`; keys are inferred from the imported JSON, with no central key list to edit. Shared messages use `getMessages("common")`. All visible copy and accessible labels come from messages; financial figures and instrument identifiers come from data. Empty screen message files are ready for their owners.

Run `npx tsc --noEmit` and `npx eslint`. In PowerShell with script execution disabled, use `npx.cmd`. Run `node scripts/verify-foundation.mjs` for decimal, risk, fixture and rendering checks. Google fonts are downloaded by Next.js at build time: a full initial build needs those font resources available; do not add a runtime Google stylesheet or install dependencies.
