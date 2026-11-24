import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { test, afterEach } from "node:test";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { loadSource } from "./load-source.mjs";

const { apiRequest, ApiError } = loadSource("src/lib/auth.ts");
const { requiresLiveConfirmation, tradableAccounts, accountLabel, selectTradingAccount, orderForExecution } = loadSource("src/lib/accounts.ts");
const { tradingError, tradingErrorCode } = loadSource("src/lib/trading.ts");
const { alertPollInterval, alertStorageKey, readAlertCursor, saveAlertCursor, unseenAlerts, pollTriggeredAlerts, validAlertPrice } = loadSource("src/lib/alerts.ts");
const { AccountBadges } = loadSource("src/components/metron/account-badges.tsx");
const { ExchangeStop } = loadSource("src/app/portfolio/_components/positions.tsx");
const { LiveConfirmation } = loadSource("src/app/check/_components/live-confirmation.tsx");
const { DELETE: proxyDelete } = loadSource("src/app/v1/[...path]/route.ts");
const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });
const account = { id: "a", type: "live", environment: "mainnet", name: "Binance", market: "spot", quote_asset: "USDT", key_hint: "…WXYZ" };

test("confirmation is required only for live mainnet; paper selection is preserved", () => {
  assert.equal(requiresLiveConfirmation(account), true);
  assert.equal(requiresLiveConfirmation({ ...account, environment: "testnet" }), false);
  assert.equal(requiresLiveConfirmation({ ...account, type: "paper", environment: null }), false);
  const paper = { ...account, id: "paper", type: "paper" };
  assert.deepEqual(tradableAccounts([paper, account, { ...account, quote_asset: "BTC" }]), [paper, account]);
  assert.equal(selectTradingAccount([account, paper], ""), paper);
  assert.equal(selectTradingAccount([account, paper], "a"), account);
  assert.equal(selectTradingAccount([account], ""), account);
});

test("live badges and native option labels include environment and only the key hint", () => {
  const html = renderToStaticMarkup(React.createElement(AccountBadges, { account, locale: "en" }));
  for (const word of ["LIVE", "MAINNET", "…WXYZ"]) { assert.ok(html.includes(word)); assert.ok(accountLabel(account, "tr").includes(word)); }
  const paper = renderToStaticMarkup(React.createElement(AccountBadges, { account: { ...account, type: "paper" }, locale: "en" }));
  assert.ok(!paper.includes("MAINNET"));
});

test("same-origin DELETE accepts 204 and preserves API conflict details", async () => {
  globalThis.fetch = async (url, init) => {
    assert.equal(url, "/v1/accounts/a"); assert.equal(init.method, "DELETE"); assert.equal(init.credentials, "same-origin"); assert.equal(init.body, undefined);
    return new Response(null, { status: 204 });
  };
  assert.equal(await apiRequest("/accounts/a", undefined, undefined, "DELETE"), undefined);
  globalThis.fetch = async () => Response.json({ error: "Open positions", error_tr: "Açık pozisyonlar" }, { status: 409 });
  await assert.rejects(apiRequest("/accounts/a", undefined, undefined, "DELETE"), error => error.status === 409 && tradingError(error, "tr") === "Açık pozisyonlar");
});

test("the proxy forwards DELETE query, cookies and a bodyless 204", async () => {
  globalThis.fetch = async (url, init) => {
    assert.ok(url.endsWith("/v1/alerts/a%2Fb?example=1")); assert.equal(init.method, "DELETE");
    assert.equal(init.headers.get("cookie"), "session=example"); assert.equal(init.body, undefined);
    return new Response(null, { status: 204 });
  };
  const response = await proxyDelete({ method: "DELETE", headers: new Headers({ cookie: "session=example" }), nextUrl: new URL("http://localhost/v1/alerts/a?example=1") }, { params: Promise.resolve({ path: ["alerts", "a/b"] }) });
  assert.equal(response.status, 204); assert.equal(await response.text(), "");
});

test("mainnet requires confirmation to build a request; paper and testnet keep their payloads", async () => {
  const preview = { account_id: "a", symbol: "BTC", side: "long", manual_stop: "60123.12345678" };
  assert.equal(orderForExecution(account, preview, false), null);
  const body = orderForExecution(account, preview, true);
  globalThis.fetch = async (url, init) => { assert.equal(url, "/v1/orders"); assert.deepEqual(JSON.parse(init.body), { ...preview, confirm_live: true }); return Response.json({ id: "position" }); };
  await apiRequest("/orders", body);
  for (const other of [{ ...account, environment: "testnet" }, { ...account, type: "paper", environment: null }]) {
    assert.deepEqual(orderForExecution(other, preview, false), preview);
    assert.ok(!("confirm_live" in orderForExecution(other, preview, false)));
  }
  assert.ok(!("confirm_live" in preview));
});

for (const code of ["live_confirmation_required", "exchange_rejected", "exchange_unavailable", "invalid_keys", "exchange_stop_failed"]) {
  test(`${code} has EN/TR messages for both error code shapes`, () => {
    const error = new ApiError(400, null, { code });
    assert.equal(tradingErrorCode(error), code);
    assert.notEqual(tradingError(error, "en"), tradingError(error, "tr"));
    assert.equal(tradingErrorCode(new ApiError(400, null, { error: { code } })), code);
  });
}

test("unknown bilingual exchange errors retain their server messages", () => {
  const error = new ApiError(422, null, { error: { message_en: "Exchange detail", message_tr: "Borsa ayrıntısı" } });
  assert.equal(tradingError(error, "en"), "Exchange detail"); assert.equal(tradingError(error, "tr"), "Borsa ayrıntısı");
});

for (const status of ["placed", "failed", "filled", "cancelled"]) {
  test(`exchange stop ${status} renders status and appropriate retry action`, () => {
    const html = renderToStaticMarkup(React.createElement(ExchangeStop, { position: { id: "p", status: "open", exchange_stop: { status, order_id: "123", error: null } }, locale: "en", busy: false, retrying: false, onRetry() {} }));
    if (status === "placed") { assert.ok(html.includes("123")); assert.ok(html.includes("text-profit")); }
    if (status === "failed") { assert.ok(html.includes('role="alert"')); assert.ok(html.includes("Retry stop")); assert.ok(html.includes("text-limit")); }
    else assert.ok(!html.includes("Retry stop"));
  });
}

test("paper positions have no exchange stop UI; closed failed stops cannot be retried", () => {
  const props = { locale: "tr", busy: false, retrying: false, onRetry() {} };
  assert.equal(renderToStaticMarkup(React.createElement(ExchangeStop, { ...props, position: { exchange_stop: null } })), "");
  const html = renderToStaticMarkup(React.createElement(ExchangeStop, { ...props, position: { status: "closed", exchange_stop: { status: "failed" } } }));
  assert.ok(!html.includes("<button"));
});

test("mainnet confirmation identifies real money and all five preview values", () => {
  const html = renderToStaticMarkup(React.createElement(LiveConfirmation, { check: { symbol: "BTC", side: "short", position_size: "0.12345678", notional: "65000.12", stop: "67000.12345678" }, locale: "en", onCancel() {}, onConfirm() {} }));
  for (const word of ["MAINNET", "Real money", "BTC", "Short", "0.12345678", "65,000.12", "67,000.12345678", "Cancel"]) assert.ok(html.includes(word), word);
  assert.ok(html.includes('aria-labelledby="live-confirm-title"'));
});

test("polling uses an encoded RFC3339 cursor and the shared signed request helper", async () => {
  const cursor = "2026-11-24T12:00:00.000Z";
  assert.equal(alertPollInterval, 15_000);
  globalThis.fetch = async (url, init) => { assert.equal(new URL(url, "http://localhost").searchParams.get("triggered_since"), cursor); assert.equal(init.credentials, "same-origin"); return Response.json([]); };
  assert.deepEqual(await pollTriggeredAlerts(cursor, new AbortController().signal), []);
});

test("alert prices preserve exact decimal precision and reject zero or formatted input", () => {
  for (const value of ["65000", "0.000000001", "9007199254740993.000000001"]) assert.equal(validAlertPrice(value), true, value);
  for (const value of ["0", "0.00", "-1", "1e3", "65,000", "1.2.3", "", " 1", "1,5"]) assert.equal(validAlertPrice(value), false, value);
});

test("triggered alerts deduplicate repeated/overlapping responses and exclude other statuses", () => {
  const cursor = "2026-11-24T12:00:00Z";
  const triggered = { id: "1", status: "triggered", triggered_at: "2026-11-24T12:00:01Z" };
  const seen = new Set();
  assert.deepEqual(unseenAlerts([triggered, triggered, { ...triggered, id: "2", status: "active" }, { ...triggered, id: "3", triggered_at: cursor }], cursor, seen), [triggered]);
  assert.deepEqual(unseenAlerts([triggered], cursor, seen), []);
});

test("cursor persistence is scoped to the user and tolerates denied storage and invalid timestamps", () => {
  const now = "2026-11-24T12:00:00.000Z";
  assert.notEqual(alertStorageKey("one"), alertStorageKey("two"));
  const denied = { getItem() { throw Error("denied"); }, setItem() { throw Error("denied"); } };
  assert.equal(readAlertCursor(denied, "key", now), now);
  assert.doesNotThrow(() => saveAlertCursor(denied, "key", now));
  for (const saved of [null, "bad date", "2099-01-01T00:00:00Z"]) assert.equal(readAlertCursor({ getItem: () => saved }, "key", now), now);
  assert.equal(readAlertCursor({ getItem: () => "2026-11-24T11:00:00Z" }, "key", now), "2026-11-24T11:00:00.000Z");
});

test("all EN/TR namespaces have matching keys and interpolation variables", () => {
  const directory = path.resolve(import.meta.dirname, "../src/messages");
  for (const filename of fs.readdirSync(path.join(directory, "en"))) {
    const en = JSON.parse(fs.readFileSync(path.join(directory, "en", filename), "utf8"));
    const tr = JSON.parse(fs.readFileSync(path.join(directory, "tr", filename), "utf8"));
    assert.deepEqual(Object.keys(en).sort(), Object.keys(tr).sort(), filename);
    for (const key of Object.keys(en)) assert.deepEqual((en[key].match(/\{\w+\}/g) ?? []).sort(), (tr[key].match(/\{\w+\}/g) ?? []).sort(), `${filename}.${key}`);
  }
});
