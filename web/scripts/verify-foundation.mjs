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
const { LocaleProvider } = component("locale-provider");
const { CheckResult } = load(resolve(root, "src/app/check/_components/check-result"));
for (const check of [data.preTradeCaution, data.preTradeBlocked]) {
  const markup = renderToStaticMarkup(createElement(LocaleProvider, { locale: "tr" }, createElement(CheckResult, { locale: "tr", check })));
  const text = markup.replace(/<[^>]*>/g, "");
  assert.ok(markup.includes(`data-risk="${check.verdict}"`));
  assert.ok(text.includes(formatAmount(check.orderValue, 2, "tr")));
  assert.ok(text.includes(formatPrice(check.entry, undefined, "tr")));
  assert.ok(text.includes(formatPercent(check.riskPercent, 2, "tr")));
  assert.doesNotMatch(text, /\d+\.\d+%|%\d+\.\d+/);
}
assert.equal(getMessages("check", "tr")("transitionPercent", {
  before: formatPercent("2.75", 2, "tr"), after: formatPercent("3.25", 2, "tr"), limit: formatPercent("6", 2, "tr"),
}), "%2,75 → %3,25 / %6,00");
const { formatVolume } = load(resolve(root, "src/app/markets/_lib/market-data"));
assert.equal(formatVolume("38200000000", "tr"), "38,2B");
assert.equal(formatVolume("1200000000", "en"), "1.2B");
assert.equal(formatVolume("1000000000", "tr"), "1B");

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
  for (const name of ["text", "muted", "subtle", "profit", "loss", "safe", "caution", "limit"]) {
    for (const background of ["bg", "surface", "raised"]) {
      assert.ok(contrast(rgb(palette[name]), rgb(palette[background])) >= 4.5, `${theme}: ${name} on ${background} is AA`);
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
console.log("Passed decimal precision, risk boundaries, messages, component rendering, source fixtures, theme bootstrap, tokens, color isolation, and Tailwind compilation checks.");
