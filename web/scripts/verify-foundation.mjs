import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { createRequire } from "node:module";
import { dirname, extname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { runInNewContext } from "node:vm";
import ts from "typescript";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

const root = fileURLToPath(new URL("../", import.meta.url));
const require = createRequire(import.meta.url);
const cache = new Map();
// Compile to memory, so checks require neither a build nor another dependency.
function load(path) {
  const file = [path, `${path}.ts`, `${path}.tsx`, `${path}.json`, join(path, "index.ts"), join(path, "index.tsx")]
    .find(candidate => existsSync(candidate) && statSync(candidate).isFile());
  assert.ok(file, `Module exists: ${path}`);
  if (cache.has(file)) return cache.get(file).exports;
  const source = readFileSync(file, "utf8");
  if (extname(file) === ".json") return JSON.parse(source);
  const loaded = { exports: {} };
  cache.set(file, loaded);
  const output = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, jsx: ts.JsxEmit.ReactJSX, esModuleInterop: true } }).outputText;
  const localRequire = (id) => id.startsWith("@/") ? load(resolve(root, "src", id.slice(2))) : id.startsWith(".") ? load(resolve(dirname(file), id)) : require(id);
  new Function("require", "module", "exports", output)(localRequire, loaded, loaded.exports);
  return loaded.exports;
}
const lib = (name) => load(resolve(root, "src/lib", name));
const component = (name) => load(resolve(root, "src/components/metron", name));
const { formatAmount, formatPrice, formatPercent, decimalProduct, decimalSum, reformatInput } = lib("format");
for (const [value, decimals, expected] of [
  ["1.005", 2, "1.01"], ["−1.005", 2, "−1.01"], ["999.995", 2, "1,000.00"],
  ["9007199254740993.125", 2, "9,007,199,254,740,993.13"], ["-0.004", 2, "0.00"],
  ["25,000.00", 2, "25,000.00"], [BigInt("12345678901234567890"), 0, "12,345,678,901,234,567,890"],
]) assert.equal(formatAmount(value, decimals), expected);
assert.equal(formatPercent("2.75"), "2.75%");
assert.equal(formatAmount("-1234.5", 2, "tr"), "−1.234,50");
assert.equal(formatAmount("25000", 2, "tr"), "25.000,00");
assert.equal(formatAmount("999.995", 2, "tr"), "1.000,00");
assert.equal(formatAmount("9007199254740993.125", 2, "tr"), "9.007.199.254.740.993,13");
assert.equal(formatAmount("-0.004", 2, "tr"), "0,00");
assert.equal(formatPercent("2.75", 2, "tr"), "%2,75");
assert.equal(formatPercent("-2.75", 2, "tr"), "−%2,75");
assert.equal(formatPrice("96400.00500", undefined, "tr"), "96.400,00500");
assert.equal(formatPrice("1e-7", undefined, "tr"), "0,0000001");
assert.equal(reformatInput("96,400.00500", "en", "tr"), "96.400,00500");
assert.equal(reformatInput("96.400,00500", "tr", "en"), "96,400.00500");
assert.equal(reformatInput("", "tr", "en"), "");
assert.equal(formatPrice("5.200"), "5.200");
assert.equal(formatPrice("1.0001"), "1.0001");
assert.equal(decimalProduct("0.1", "0.2"), "0.02");
assert.equal(decimalProduct("6.00", "0.75"), "4.5000");
assert.equal(decimalSum("0.1", "0.2"), "0.3");
assert.equal(formatPrice("1e-7"), "0.0000001");
assert.throws(() => formatAmount("not-money"), TypeError);

const { riskState } = lib("risk");
assert.equal(riskState(74.99, 100), "safe");
assert.equal(riskState(75, 100), "caution");
assert.equal(riskState(100, 100), "caution");
assert.equal(riskState(100.01, 100), "limit");
assert.equal(riskState(6.4, 6), "limit");
assert.equal(riskState(0, 0), "limit");
assert.equal(riskState(NaN, 100), "limit");
assert.equal(riskState(50, 100, 0.5), "caution");
assert.equal(riskState(0.075, 0.1), "caution");

// Compare the complete catalogs, including placeholder multiplicity, before rendering.
const messageRoot = resolve(root, "src/messages");
const englishFiles = readdirSync(join(messageRoot, "en")).sort();
assert.deepEqual(readdirSync(join(messageRoot, "tr")).sort(), englishFiles, "Locale namespace files match");
const placeholders = (text) => [...text.matchAll(/\{(\w+)\}/g)].map((match) => match[1]).sort();
for (const file of englishFiles) {
  const en = JSON.parse(readFileSync(join(messageRoot, "en", file), "utf8"));
  const tr = JSON.parse(readFileSync(join(messageRoot, "tr", file), "utf8"));
  assert.deepEqual(Object.keys(tr).sort(), Object.keys(en).sort(), `${file}: English/Turkish keys match`);
  for (const key of Object.keys(en)) {
    assert.equal(typeof tr[key], "string", `${file}.${key}: Turkish message is text`);
    assert.ok(tr[key].trim().length, `${file}.${key}: Turkish message is not empty`);
    assert.deepEqual(placeholders(tr[key]), placeholders(en[key]), `${file}.${key}: placeholders match`);
  }
}
const { getMessages, parseLocale } = lib("messages");
assert.equal(parseLocale(undefined), "en");
assert.equal(parseLocale("invalid"), "en");
assert.equal(parseLocale("tr"), "tr");
assert.equal(getMessages("common", "tr")("nav.check"), "İşlem öncesi kontrol");
assert.equal(getMessages("common", "tr")("measure.value", { value: "1", limit: "2", state: "Uygun" }), "1 / 2 · Uygun");
assert.throws(() => getMessages("common", "tr")("measure.value"), /Missing interpolation/);
assert.equal(getMessages("common")("measure.value", { value: "1", limit: "2", state: "Safe" }), "1 of 2 · Safe");
assert.throws(() => getMessages("common")("measure.value"), /Missing interpolation/);
const { PnL } = component("pnl");
assert.match(renderToStaticMarkup(createElement(PnL, { value: "-55.59" })), /text-loss.*▼ −55\.59/);
assert.match(renderToStaticMarkup(createElement(PnL, { value: "125.01" })), /text-profit.*▲ \+125\.01/);
assert.match(renderToStaticMarkup(createElement(PnL, { value: "-1234.5", locale: "tr" })), /text-loss.*▼ −1\.234,50/);
assert.match(renderToStaticMarkup(createElement(PnL, { value: "2.75", locale: "tr", suffix: "%" })), /▲ \+%2,75/);
assert.match(renderToStaticMarkup(createElement(PnL, { value: "-0.004", locale: "tr", suffix: "%" })), /text-muted.*%0,00/);
const zero = renderToStaticMarkup(createElement(PnL, { value: "-0.004" }));
assert.match(zero, /text-muted.*0\.00/);
assert.doesNotMatch(zero, /[▲▼−]/);
const { Measure } = component("measure");
assert.match(renderToStaticMarkup(createElement(Measure, { value: 70, limit: 100, locale: "tr" })), /70,00 \/ 100,00 · Uygun/);
const ruler = renderToStaticMarkup(createElement(Measure, { value: 70, added: 40, limit: 100 }));
assert.match(ruler, /role="meter"/);
assert.match(ruler, /aria-valuenow="100"/);
assert.match(ruler, /70\.00 now, 40\.00 added, limit 100\.00 · Limit exceeded/);
assert.match(ruler, /measure-added[^>]*left:70%;width:30%/);
assert.match(ruler, /text-limit/);
const percentRuler = renderToStaticMarkup(createElement(Measure, { value: 2.75, limit: 6, formatValue: formatPercent }));
assert.match(percentRuler, /4\.50%/);
const fractionalRuler = renderToStaticMarkup(createElement(Measure, { value: 0.1, added: 0.2, limit: 0.3 }));
assert.match(fractionalRuler, /aria-valuenow="0.3"/);
assert.doesNotMatch(fractionalRuler, /Limit exceeded/);
const rowRuler = renderToStaticMarkup(createElement(Measure, { value: 70, limit: 100, showStatus: false }));
assert.doesNotMatch(rowRuler, /measure-level|risk-badge/);
assert.match(rowRuler, /aria-valuetext="70.00 of 100.00 · Safe"/);

const data = lib("sample-data");
assert.equal(data.getSampleData("tr").journal[0].how, "Stop tetiklendi");
assert.equal(data.getSampleData("tr").equitySeries.dates[0].label, "16 Ağu");
assert.equal(data.getSampleData("tr").account.equity, data.account.equity);
function mockup(name) {
  const body = readFileSync(resolve(root, `../design/mockups/${name}.dc.html`), "utf8").replaceAll("\r", "").split("renderVals() {")[1].split("\n}\n}")[0];
  return new Function(body).call({ props: {} });
}
const normalize = (value) => typeof value === "string" && value.startsWith("var(--") ? value.replace("var(--", "").replace(")", "").replace("-t", "-tint") : value;
for (const [name, field, fixture] of [
  ["Main", "gauges", data.gauges], ["Main", "positions", data.openPositions], ["Markets", "rows", data.markets],
  ["Rules", "classes", data.classRules], ["Rules", "discipline", data.disciplineRules],
  ["Portfolio", "journal", data.journal], ["Portfolio", "classes", data.classStatistics], ["Portfolio", "holdings", data.holdings],
  ["PreTrade", "checks", data.preTradeCaution.checks],
]) {
  const source = mockup(name)[field];
  assert.equal(fixture.length, source.length);
  source.forEach((row, index) => Object.entries(row).forEach(([key, value]) => assert.equal(fixture[index][key], normalize(value), `${name}.${field}.${index}.${key}`)));
}
assert.equal(data.preTradeCaution.entry, "5.200");
assert.equal(data.preTradeCaution.quantity, "200.32");
assert.equal(data.preTradeCaution.market, "spot");
assert.equal(data.preTradeBlocked.market, "futures");
assert.equal(data.preTradeBlocked.checks[0].value, "3 → 3 of 3");
assert.equal(data.preTradeBlocked.checks[0].over, true);
assert.deepEqual([data.limits.riskPerTrade, data.limits.totalOpenRisk, data.limits.maxPositions, data.limits.losingTradesPerDay, data.limits.dailyLoss, data.limits.stoppedTradesPerWeek, data.limits.weeklyLoss], ["2.00", "6.00", 5, 3, "5.00", 5, "10.00"]);
assert.equal(data.equitySeries.account.length, 13);
const { LocaleProvider } = component("locale-provider");
const { CheckResult } = load(resolve(root, "src/app/check/_components/check-result"));
const liveCheck = {
  symbol: "BTCUSDT", side: "long", coin_class: "core", market: "spot", leverage: 1,
  entry_price: "100.12345678", stop: "95", target_price: "110.37037034",
  account_equity: "10000", atr: "3.415637853333333333", atr_multiplier: "1.5",
  stop_percent: "5.117", risk_percent: "1", risk_amount: "100", actual_risk: "99.99999999",
  position_size: "19.517" , notional: "1954.109", target_risk_reward: "2",
  allowed: true, verdict: "pass", passed: 1, cautions: 0, failed: 0,
  checks: [{ rule_id: "total_open_risk", status: "pass", values: { current_percent: "1", proposed_percent: "2", limit_percent: "6" }, message_en: "Total open risk includes the proposed trade.", message_tr: "Toplam açık risk önerilen işlemi içerir." }],
};
for (const locale of ["en", "tr"]) for (const status of ["pass", "caution", "fail"]) {
  // A failed row must block even if allowed/verdict are inconsistent.
  const check = { ...liveCheck, checks: [{ ...liveCheck.checks[0], status }] };
  const markup = renderToStaticMarkup(createElement(LocaleProvider, { locale }, createElement(CheckResult, { locale, check, onOpen() {} })));
  const text = markup.replace(/<[^>]*>/g, "");
  const state = { pass: "safe", caution: "caution", fail: "limit" }[status];
  assert.ok(markup.includes(`data-verdict="${state}"`));
  assert.ok(text.includes(formatAmount(check.notional, 2, locale)));
  assert.ok(text.includes(formatPrice(check.entry_price, undefined, locale)));
  assert.ok(text.includes(formatPercent(check.risk_percent, 2, locale)));
  assert.ok(text.includes(check.checks[0][`message_${locale}`]));
  assert.equal(markup.includes('class="blocked-stamp"'), status === "fail");
  assert.doesNotMatch(markup, /measure-level/, "Check rows retain only the trailing verdict");
  assert.equal((markup.match(/class="risk-badge /g) ?? []).length, 1, "Exactly one status badge per rule row");
  assert.equal(/<button[^>]* disabled=""/.test(markup), status === "fail", "A failed check disables the order action");
}
const { PositionTable } = load(resolve(root, "src/app/portfolio/_components/positions"));
const position = {
  id: "position-1", symbol: "BTCUSDT", side: "short", quantity: "9007199254740993.00000001",
  entry_price: "100.12345678", stop: "105.12345678", leverage: 10, coin_class: "core",
  initial_risk: "100", status: "open", opened_at: "2026-11-19T00:00:00Z", closed_at: null,
  exit_price: null, realized_pnl: null, close_reason: null, last_price: null, unrealized_pnl: null,
};
for (const locale of ["en", "tr"]) {
  const props = { locale, closingId: null, onClose() {} };
  const open = renderToStaticMarkup(createElement(PositionTable, { ...props, rows: [position], closed: false }));
  assert.ok(open.includes(formatPrice(position.quantity, undefined, locale)), "Position quantities retain precision beyond JS Number");
  assert.ok(open.includes(getMessages("portfolio", locale)("unavailable")), "Missing quotes are explicit");
  assert.match(open, /<button/);
  const closing = renderToStaticMarkup(createElement(PositionTable, { ...props, closingId: position.id, rows: [position], closed: false }));
  assert.match(closing, /<button[^>]* disabled=""/, "Close cannot be submitted twice while pending");
  for (const reason of ["manual", "stop"]) {
    const closed = renderToStaticMarkup(createElement(PositionTable, { ...props, closed: true, rows: [{ ...position, status: "closed", exit_price: "95.12345678", realized_pnl: "-10.005", close_reason: reason, closed_at: "2026-11-19T01:02:03Z" }] }));
    assert.ok(closed.includes(getMessages("portfolio", locale)(`reason.${reason}`)));
    assert.ok(closed.includes("01:02:03"), "Close times render in UTC regardless of machine timezone");
    assert.match(closed, /data-pnl="negative"/);
    assert.doesNotMatch(closed, /<button/, "Closed positions have no close action");
  }
  for (const closed of [false, true]) {
    const empty = renderToStaticMarkup(createElement(PositionTable, { ...props, closed, rows: [] }));
    assert.ok(empty.includes(getMessages("portfolio", locale)(closed ? "emptyClosed" : "emptyOpen")));
  }
}
assert.equal(getMessages("check", "tr")("transitionPercent", {
  before: formatPercent("2.75", 2, "tr"), after: formatPercent("3.25", 2, "tr"), limit: formatPercent("6", 2, "tr"),
}), "%2,75 → %3,25 / %6,00");
const { formatVolume } = load(resolve(root, "src/app/markets/_lib/market-data"));
assert.equal(formatVolume("38200000000", "tr"), "38,2B");
assert.equal(formatVolume("1200000000", "en"), "1.2B");
assert.equal(formatVolume("1000000000", "tr"), "1B");

// Exercise the ticker route against the Markets API contract without a network.
const { GET: getTicker } = load(resolve(root, "src/app/api/market-ticker/route"));
const { sampleTicker } = lib("market-ticker");
const liveQuote = {
  symbol: "BTCUSDT", base_asset: "BTC", quote_asset: "USDT", coin_class: "core",
  price: "9007199254740993.125", change_24h_percent: "-2.10",
  quote_volume_24h: null, price_time: null, atr_percent: null, atr: null, suggested_stop_percent: null,
};
const originalFetch = globalThis.fetch;
try {
  for (const [upstream, expected] of [
    [{ markets: [liveQuote, { ...liveQuote, symbol: "ETHUSDT", base_asset: "ETH", price: null, change_24h_percent: null }] }, {
      source: "live", quotes: [
        { pair: "BTC/USDT", price: liveQuote.price, change: "-2.10" },
        { pair: "ETH/USDT", price: null, change: null },
      ],
    }],
    [{ markets: [] }, sampleTicker],
    [{ markets: [{ ...liveQuote, price: null }] }, sampleTicker],
    [{ markets: [{ ...liveQuote, price: "bad" }] }, sampleTicker],
    [null, sampleTicker],
  ]) {
    globalThis.fetch = async url => {
      assert.ok(String(url).endsWith("/v1/markets"), "Ticker uses the same API as Markets");
      if (upstream === null) throw new Error("Offline");
      return Response.json(upstream);
    };
    const response = await getTicker();
    assert.equal(response.headers.get("Cache-Control"), "no-store");
    assert.deepEqual(await response.json(), expected);
  }
  globalThis.fetch = async () => new Response(null, { status: 503 });
  assert.deepEqual(await (await getTicker()).json(), sampleTicker);
} finally {
  globalThis.fetch = originalFetch;
}

// Auth regressions: exact API boundaries, JSON mutations and both session cookies.
const { apiRequest, ApiError, apiError, errorText, validPassword, validCode, validPaperBalance } = lib("auth");
const { checkBlocked, rejectedCheck, tradingError, validStop, meterWidth } = lib("trading");
assert.equal(checkBlocked(liveCheck), false);
assert.equal(checkBlocked({ ...liveCheck, allowed: false }), true);
assert.equal(checkBlocked({ ...liveCheck, checks: [{ ...liveCheck.checks[0], status: "fail" }] }), true);
for (const value of ["0.00000001", "1000000000000", "100.12345678"]) assert.equal(validStop(value), true);
for (const value of ["0", "-1", "1e2", "1,25", " 1", "1000000000000.00000001", "1.123456789"]) assert.equal(validStop(value), false);
assert.equal(meterWidth("9007199254740993", "18014398509481986"), 50);
assert.equal(meterWidth("0.00000001", "0.00000002"), 50);
assert.equal(meterWidth("7", "6"), 100);
assert.equal(meterWidth("1", "0"), 0);
assert.equal(rejectedCheck(new ApiError(422, null, { check: liveCheck })).allowed, false);
assert.equal(rejectedCheck(new ApiError(503)), null);
const bilingualError = new ApiError(422, null, { error: { id: "risk_blocked", message_en: "Risk blocked", message_tr: "Risk engeli" } });
assert.equal(tradingError(bilingualError, "en"), "Risk blocked");
assert.equal(tradingError(bilingualError, "tr"), "Risk engeli");
assert.equal(tradingError(new ApiError(400, null, { error: "Wrong stop", error_tr: "Hatalı stop" }), "tr"), "Hatalı stop");
for (const value of ["0.00000001", "10000000", "10000000.00000000", "00000001.12345678"]) assert.equal(validPaperBalance(value), true, value);
for (const value of ["0", "0.00000000", "-1", "+1", " 1", "1 ", "1e2", ".1", "1.", "1,25", "10000000.00000001", "99999999", "000000001", "1.123456789"]) assert.equal(validPaperBalance(value), false, value);
assert.equal(validPassword("😀".repeat(12)), true, "Password length counts Unicode characters");
assert.equal(validPassword("x".repeat(11)), false);
assert.equal(validPassword("x".repeat(128)), true);
assert.equal(validPassword("x".repeat(129)), false);
assert.equal(validCode("000001"), true, "Leading-zero codes stay strings");
for (const value of ["12345", "1234567", " 123456", "12345a"]) assert.equal(validCode(value), false);
assert.equal(apiError(new ApiError(400)).key, "error.badRequest");
assert.equal(apiError(new ApiError(401), "login").key, "error.credentials");
assert.equal(apiError(new ApiError(401), "code").key, "error.code");
assert.equal(apiError(new ApiError(401), "disable").key, "error.disable");
assert.equal(apiError(new ApiError(401), "accounts").key, "error.session");
assert.equal(apiError(new ApiError(409), "register").key, "error.duplicate");
assert.equal(apiError(new ApiError(409), "settings").key, "error.conflict");
assert.equal(apiError(new ApiError(503)).key, "error.unavailable");
const limited = apiError(new ApiError(429, "900"));
assert.equal(limited.seconds, "900");
assert.ok(errorText(limited, "en").includes("900 seconds"));
assert.ok(errorText(limited, "tr").includes("900 saniye"));
try {
  for (const [path, body] of [["/auth/register", { email: "test@example.com", password: " password with spaces " }], ["/auth/login/2fa", { code: "000001" }], ["/auth/logout", {}], ["/auth/2fa/setup", {}], ["/accounts", { name: "Practice", paper_balance: "0.00000001" }]]) {
    globalThis.fetch = async (url, options) => {
      assert.equal(url, `/v1${path}`);
      assert.equal(options.credentials, "same-origin");
      assert.equal(options.cache, "no-store");
      assert.equal(options.method, "POST");
      assert.equal(options.headers["Content-Type"], "application/json");
      assert.deepEqual(JSON.parse(options.body), body);
      return Response.json({ ok: true });
    };
    await apiRequest(path, body);
  }
  globalThis.fetch = async () => Response.json({ error: "rate limited" }, { status: 429, headers: { "Retry-After": "120" } });
  await assert.rejects(apiRequest("/auth/login", {}), error => error instanceof ApiError && error.retryAfter === "120");
  globalThis.fetch = async () => Response.json({ error: { id: "risk_blocked" }, check: liveCheck }, { status: 422 });
  await assert.rejects(apiRequest("/orders", { account_id: "account", symbol: "BTCUSDT", side: "long" }), error => {
    assert.equal(error.status, 422);
    assert.deepEqual(error.data.check, liveCheck, "Order rejection retains the authoritative check");
    assert.equal(rejectedCheck(error).allowed, false);
    return true;
  });
  const { GET: proxyGet, POST: proxyPost } = load(resolve(root, "src/app/v1/[...path]/route"));
  const cookies = ["metron_session=full; Path=/; HttpOnly; SameSite=Lax", "metron_pending=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax"];
  const request = { method: "POST", headers: new Headers({ cookie: "metron_pending=pending", "content-type": "application/json", "user-agent": "metron-test" }), text: async () => '{"code":"000001"}', nextUrl: { search: "" } };
  globalThis.fetch = async (url, options) => {
    assert.ok(url.endsWith("/v1/auth/login/2fa"));
    assert.equal(options.headers.get("cookie"), "metron_pending=pending");
    assert.equal(options.headers.get("user-agent"), "metron-test");
    assert.equal(options.body, '{"code":"000001"}');
    assert.equal(options.cache, "no-store");
    assert.equal(options.redirect, "manual");
    const headers = new Headers();
    cookies.forEach(cookie => headers.append("Set-Cookie", cookie));
    return Response.json({ id: "user" }, { headers });
  };
  const completed = await proxyPost(request, { params: Promise.resolve({ path: ["auth", "login", "2fa"] }) });
  assert.deepEqual(completed.headers.getSetCookie(), cookies, "Full and pending cookies survive the proxy separately");
  assert.equal(completed.headers.get("Cache-Control"), "no-store");
  globalThis.fetch = async (_url, options) => {
    assert.equal(options.headers.get("content-type"), "text/plain", "Proxy does not upgrade cross-site form content to JSON");
    return Response.json({ error: "JSON required" }, { status: 400 });
  };
  await proxyPost({ ...request, headers: new Headers({ "content-type": "text/plain" }) }, { params: Promise.resolve({ path: ["auth", "logout"] }) });
  globalThis.fetch = async () => Response.json({ error: "limited" }, { status: 429, headers: { "Retry-After": "900" } });
  const throttled = await proxyGet({ ...request, method: "GET" }, { params: Promise.resolve({ path: ["me"] }) });
  assert.equal(throttled.status, 429);
  assert.equal(throttled.headers.get("Retry-After"), "900");
  globalThis.fetch = async () => { throw new Error("offline"); };
  const offline = await proxyGet({ ...request, method: "GET" }, { params: Promise.resolve({ path: ["me"] }) });
  assert.equal(offline.status, 503);
  assert.equal(offline.headers.get("Cache-Control"), "no-store");
} finally {
  globalThis.fetch = originalFetch;
}
assert.equal(sampleTicker.quotes[0].price, "97846.00");
assert.equal(sampleTicker.quotes.length, data.markets.length);
for (const locale of ["en", "tr"]) {
  const { MarketTicker } = component("market-ticker");
  const markup = renderToStaticMarkup(createElement(LocaleProvider, { locale }, createElement(MarketTicker)));
  assert.ok(markup.includes(getMessages("common", locale)("ticker.sample")));
  assert.ok(markup.includes('aria-hidden="true"'));
  assert.ok(markup.includes(formatPrice("97846.00", undefined, locale)));
}

const { themeBootstrap } = lib("theme");
for (const stored of ["light", "dark", "invalid", null]) {
  const document = { documentElement: { dataset: {} } };
  runInNewContext(themeBootstrap, { document, localStorage: { getItem: () => stored } });
  assert.equal(document.documentElement.dataset.theme, stored === "light" ? "light" : "dark");
}
const document = { documentElement: { dataset: {} } };
runInNewContext(themeBootstrap, { document, localStorage: { getItem: () => { throw new Error("Storage denied"); } } });
assert.equal(document.documentElement.dataset.theme, "dark");

const tokens = JSON.parse(readFileSync(resolve(root, "../design/tokens.json"), "utf8"));
// Small text and semantic labels must remain AA on both surface hierarchies.
const rgb = (hex) => hex.slice(1).match(/../g).map(value => parseInt(value, 16));
const luminance = (color) => color.map(value => value / 255)
  .map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4)
  .reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
const contrast = (foreground, background) => {
  const a = luminance(foreground), b = luminance(background);
  return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
};
for (const [theme, palette] of Object.entries(tokens.color)) {
  const [r, g, b, alpha] = palette.limitTint.match(/[\d.]+/g).map(Number);
  const mix = (front, back, amount) => back.map((value, index) => value * (1 - amount) + front[index] * amount);
  const redWash = mix([r, g, b], rgb(palette.surface), alpha);
  const roomBackgrounds = {
    blockedBanner: theme === "dark" ? mix(rgb(palette.limitEdge), rgb(palette.surface), 0.32) : redWash,
    hoveredRow: theme === "dark" ? mix([r, g, b], rgb(palette.raised), alpha) : redWash,
  };
  for (const name of ["text", "muted", "subtle", "profit", "loss", "safe", "caution", "limit"]) {
    for (const background of ["bg", "surface", "raised"]) {
      assert.ok(contrast(rgb(palette[name]), rgb(palette[background])) >= 4.5, `${theme}: ${name} on ${background} is AA`);
    }
    for (const [background, color] of Object.entries(roomBackgrounds)) {
      assert.ok(contrast(rgb(palette[name]), color) >= 4.5, `${theme}: ${name} on ${background} is AA`);
    }
    if (palette[`${name}Tint`]) {
      const [r, g, b, alpha] = palette[`${name}Tint`].match(/[\d.]+/g).map(Number);
      const background = rgb(palette.surface).map((value, index) => value * (1 - alpha) + [r, g, b][index] * alpha);
      assert.ok(contrast(rgb(palette[name]), background) >= 4.5, `${theme}: ${name} on its tint is AA`);
    }
  }
}
const css = readFileSync(resolve(root, "src/app/globals.css"), "utf8");
for (const palette of Object.values(tokens.color)) for (const [key, color] of Object.entries(palette)) {
  const name = key.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`);
  assert.ok(css.includes(`--${name}: ${color};`));
  assert.ok(css.includes(`--color-${name}: var(--${name});`));
}
function checkColors(dir) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) checkColors(path);
    else if (entry.name !== "globals.css") assert.doesNotMatch(readFileSync(path, "utf8"), /#[0-9A-Fa-f]{3,8}|rgba\(/, path);
  }
}
checkColors(resolve(root, "src"));
const { compile } = require("@tailwindcss/node");
const compiler = await compile(css, { base: resolve(root, "src/app"), from: resolve(root, "src/app/globals.css"), onDependency() {} });
const utilities = ["bg-surface", "text-muted", "border-line", "text-profit", "bg-caution-tint", "num", "p-card", "p-page", "max-w-canvas", "rounded-card", "text-title", "dark:bg-raised"];
const compiled = compiler.build(utilities);
for (const utility of utilities.filter((name) => !name.includes(":"))) assert.ok(compiled.includes(`.${utility} {`), `Tailwind emits ${utility}`);
// Tailwind resolves a named max-w-* against --spacing-* before --container-*,
// so a shared name would cap the page at the padding width.
assert.match(compiled, new RegExp(`\\.max-w-canvas \\{\\s*max-width: ${tokens.size.pageMaxWidth}px;`), "max-w-canvas uses the page width, not the page padding");
assert.ok(compiled.includes(":root:not([data-theme=\"light\"])"));
console.log("Passed decimal precision, auth validation/requests/proxy cookies, localized API errors, risk boundaries, messages, component rendering, ticker live/fallback responses, source fixtures, theme bootstrap, tokens, AA surface/wash contrast, color isolation, and Tailwind compilation checks.");
