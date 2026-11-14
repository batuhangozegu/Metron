import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync } from "node:fs";
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
  const file = [path, `${path}.ts`, `${path}.tsx`, `${path}.json`].find(existsSync);
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
const { formatAmount, formatPrice, formatPercent, decimalProduct, decimalSum } = lib("format");
for (const [value, decimals, expected] of [
  ["1.005", 2, "1.01"], ["−1.005", 2, "−1.01"], ["999.995", 2, "1,000.00"],
  ["9007199254740993.125", 2, "9,007,199,254,740,993.13"], ["-0.004", 2, "0.00"],
  ["25,000.00", 2, "25,000.00"], [BigInt("12345678901234567890"), 0, "12,345,678,901,234,567,890"],
]) assert.equal(formatAmount(value, decimals), expected);
assert.equal(formatPercent("2.75"), "2.75%");
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

const { getMessages } = lib("messages");
assert.equal(getMessages("common")("measure.value", { value: "1", limit: "2", state: "Safe" }), "1 of 2 · Safe");
assert.throws(() => getMessages("common")("measure.value"), /Missing interpolation/);
const { PnL } = component("pnl");
assert.match(renderToStaticMarkup(createElement(PnL, { value: "-55.59" })), /text-loss.*▼ −55\.59/);
assert.match(renderToStaticMarkup(createElement(PnL, { value: "125.01" })), /text-profit.*▲ \+125\.01/);
const zero = renderToStaticMarkup(createElement(PnL, { value: "-0.004" }));
assert.match(zero, /text-muted.*0\.00/);
assert.doesNotMatch(zero, /[▲▼−]/);
const { Measure } = component("measure");
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

const data = lib("sample-data");
function mockup(name) {
  const body = readFileSync(resolve(root, `../design/mockups/${name}.dc.html`), "utf8").replaceAll("\r", "").split("renderVals() {")[1].split("\n}\n}")[0];
  return new Function(body).call({ props: {} });
}
const normalize = (value) => typeof value === "string" && value.startsWith("var(--") ? value.replace("var(--", "").replace(")", "").replace("-t", "-tint") : value;
for (const [name, field, fixture] of [
  ["Main", "gauges", data.gauges], ["Main", "positions", data.openPositions], ["Markets", "rows", data.markets],
  ["Rules", "classes", data.classRules], ["Rules", "discipline", data.disciplineRules],
  ["Portfolio", "journal", data.journal], ["Portfolio", "classes", data.classStatistics], ["Portfolio", "holdings", data.holdings],
  ["PreTrade", "checks", data.preTradeCaution.checks], ["PreTradeBlocked", "checks", data.preTradeBlocked.checks],
]) {
  const source = mockup(name)[field];
  assert.equal(fixture.length, source.length);
  source.forEach((row, index) => Object.entries(row).forEach(([key, value]) => assert.equal(fixture[index][key], normalize(value), `${name}.${field}.${index}.${key}`)));
}
assert.equal(data.preTradeCaution.entry, "5.200");
assert.equal(data.preTradeCaution.quantity, "200.32");
assert.equal(data.preTradeBlocked.reducedOrderValue, "1666.67");
assert.equal(data.equitySeries.account.length, 13);

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
const utilities = ["bg-surface", "text-muted", "border-line", "text-profit", "bg-caution-tint", "num", "p-card", "p-page", "max-w-page", "rounded-card", "text-title", "dark:bg-raised"];
const compiled = compiler.build(utilities);
for (const utility of utilities.filter((name) => !name.includes(":"))) assert.ok(compiled.includes(`.${utility} {`), `Tailwind emits ${utility}`);
assert.ok(compiled.includes(":root:not([data-theme=\"light\"])"));
console.log("Passed decimal precision, risk boundaries, messages, component rendering, source fixtures, theme bootstrap, tokens, color isolation, and Tailwind compilation checks.");
