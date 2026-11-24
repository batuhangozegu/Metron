import assert from "node:assert/strict";
import { afterEach, test } from "node:test";
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { loadSource } from "./load-source.mjs";

const { getRiskStatus, gaugeState, countState, worstState, disciplineState, overviewState, remainingAmount, distanceToStop, formatStopDistance, formatRiskTime } = loadSource("src/lib/risk-status.ts");
const { formatAmount } = loadSource("src/lib/format.ts");
const { DisciplineCounts } = loadSource("src/app/rules/_components/discipline-mode.tsx");
const gauge = { amount: "100", percent: "1", limit_percent: "6" };
const status = {
  account_id: "a", market: "spot", as_of: "2026-11-24T12:00:00Z", currency: "USDT", equity: "10000", realized_today: "-12.5", unrealized: "3.2",
  open_risk: gauge, daily_loss: gauge, weekly_loss: gauge,
  open_positions: { count: 3, limit: 5 }, losing_trades_today: { count: 0, limit: 3 }, stopped_trades_this_week: { count: 0, limit: 4 },
  checks_today: { run: 4, orders: 2, blocked: 1 },
};
const originalFetch = globalThis.fetch;
afterEach(() => { globalThis.fetch = originalFetch; });

test("risk status uses the authenticated endpoint and preserves the contract decimals", async () => {
  globalThis.fetch = async (url, init) => {
    assert.equal(url, "/v1/accounts/a%2Fb/risk-status");
    assert.equal(init.method, "GET"); assert.equal(init.credentials, "same-origin"); assert.equal(init.cache, "no-store");
    return Response.json(status);
  };
  assert.deepEqual(await getRiskStatus("a/b"), status);
});

test("gauges use exact decimal thresholds consistent with existing rulers", () => {
  for (const [percent, state] of [["4.499999999999999999", "safe"], ["4.5", "caution"], ["6", "caution"], ["6.000000000000000001", "limit"]]) {
    assert.equal(gaugeState({ ...gauge, percent }), state);
  }
  assert.equal(gaugeState({ ...gauge, limit_percent: "0" }), "limit");
});

test("discipline cautions one away and blocks at or beyond either limit", () => {
  for (const [count, state] of [[0, "safe"], [1, "safe"], [2, "caution"], [3, "limit"], [4, "limit"]]) assert.equal(countState({ count, limit: 3 }), state);
  assert.equal(countState({ count: 0, limit: 1 }), "caution");
  assert.equal(countState({ count: 0, limit: 0 }), "limit");
  assert.equal(disciplineState(status), "safe");
  assert.equal(disciplineState({ ...status, losing_trades_today: { count: 2, limit: 3 } }), "caution");
  assert.equal(disciplineState({ ...status, stopped_trades_this_week: { count: 4, limit: 4 } }), "limit");
});

test("headline derives the worst of all three gauges and both discipline counters", () => {
  assert.equal(overviewState(status), "safe");
  for (const key of ["open_risk", "daily_loss", "weekly_loss"]) {
    assert.equal(overviewState({ ...status, [key]: { ...gauge, percent: "4.5" } }), "caution");
    assert.equal(overviewState({ ...status, [key]: { ...gauge, percent: "7" } }), "limit");
  }
  assert.equal(overviewState({ ...status, losing_trades_today: { count: 3, limit: 3 } }), "limit");
  assert.equal(overviewState({ ...status, stopped_trades_this_week: { count: 3, limit: 4 } }), "caution");
  assert.equal(worstState(["caution", "limit", "safe"]), "limit");
});

test("remaining amounts use equity and amount, preserve precision, and clamp exceeded limits", () => {
  assert.equal(formatAmount(remainingAmount("25000", { ...gauge, amount: "687.5", limit_percent: "6" })), "812.50");
  assert.equal(formatAmount(remainingAmount("10000", { ...gauge, amount: "125", limit_percent: "5" })), "375.00");
  assert.equal(remainingAmount("100", { ...gauge, amount: "7", limit_percent: "6" }), "0");
  assert.equal(remainingAmount("100", { ...gauge, amount: "6", limit_percent: "6" }), "0");
  assert.equal(formatAmount(remainingAmount("9007199254740993", { ...gauge, amount: "0", limit_percent: "100" })), "9,007,199,254,740,993.00");
});

test("distance to stop respects long/short direction, crossed stops and unavailable quotes", () => {
  assert.equal(distanceToStop("100", "95", "long"), "5.00000000");
  assert.equal(distanceToStop("100", "105", "short"), "5.00000000");
  assert.equal(distanceToStop("100", "105", "long"), "-5.00000000");
  assert.equal(distanceToStop("100", "95", "short"), "-5.00000000");
  assert.equal(distanceToStop("100", "100", "short"), "0.00000000");
  assert.equal(distanceToStop("0.0000000001", "0.000000000095", "long"), "5.00000000");
  assert.equal(distanceToStop("100000000000000000000", "95000000000000000000", "long"), "5.00000000");
  assert.equal(distanceToStop("3", "2", "long"), "33.33333333");
  for (const price of [null, undefined, "0", "-1"]) assert.equal(distanceToStop(price, "95", "long"), null);
});

test("stop distances and UTC timestamps format for both locales", () => {
  assert.equal(formatStopDistance("5.125", "en"), "5.13%");
  assert.equal(formatStopDistance("-5.125", "tr"), "−%5,13");
  assert.equal(formatStopDistance(null, "tr"), "—");
  for (const locale of ["en", "tr"]) assert.ok(formatRiskTime("2026-11-24T23:59:01Z", locale).includes("23:59:01"));
});

test("both views share localized live counts with caution and limit badges", () => {
  const live = { ...status, losing_trades_today: { count: 2, limit: 3 }, stopped_trades_this_week: { count: 4, limit: 4 } };
  for (const locale of ["en", "tr"]) {
    const html = renderToStaticMarkup(React.createElement(DisciplineCounts, { status: live, locale }));
    assert.ok(html.includes('data-risk="caution"'));
    assert.ok(html.includes('data-risk="limit"'));
    assert.ok(html.includes(locale === "tr" ? "Bugünkü zararlı işlemler" : "Losing trades today"));
    assert.ok(html.includes(locale === "tr" ? "/ 3" : "of 3"));
  }
});
