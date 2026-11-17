/** Decimal strings (optionally grouped) and bigint preserve money without Number conversion. */
import type { Locale } from "./messages";

export type Decimal = string | bigint;

function parts(value: Decimal) {
  const match = /^([+−-]?)(\d+)(?:\.(\d+))?(?:[eE]([+-]?\d+))?$/.exec(String(value).replaceAll(",", "").trim());
  if (!match) throw new TypeError(`Invalid decimal: ${value}`);
  const exponent = Number(match[4] ?? 0);
  if (!Number.isInteger(exponent) || Math.abs(exponent) > 1000) throw new RangeError("Decimal exponent out of range");
  const digits = match[2] + (match[3] ?? "");
  const point = match[2].length + exponent;
  const whole = point <= 0 ? "0" : digits.slice(0, point).padEnd(point, "0");
  const fraction = point <= 0 ? "0".repeat(-point) + digits : digits.slice(point);
  return { negative: match[1] === "-" || match[1] === "−", whole, fraction };
}

function scaled(value: ReturnType<typeof parts>, scale: number) {
  const integer = BigInt(value.whole + value.fraction.padEnd(scale, "0"));
  return value.negative ? -integer : integer;
}

/** Compare exact decimals without binary floating point at risk thresholds. */
export function compareDecimals(left: Decimal, right: Decimal): -1 | 0 | 1 {
  const a = parts(left), b = parts(right);
  const scale = Math.max(a.fraction.length, b.fraction.length);
  const first = scaled(a, scale), second = scaled(b, scale);
  return first === second ? 0 : first < second ? -1 : 1;
}

/** Add exact decimals for a proposed exposure before converting to meter geometry. */
export function decimalSum(left: Decimal, right: Decimal): string {
  const a = parts(left), b = parts(right);
  const scale = Math.max(a.fraction.length, b.fraction.length);
  const total = scaled(a, scale) + scaled(b, scale);
  const digits = String(total < BigInt(0) ? -total : total).padStart(scale + 1, "0");
  return `${total < BigInt(0) ? "-" : ""}${scale ? `${digits.slice(0, -scale)}.${digits.slice(-scale)}` : digits}`;
}

/** Format a decimal with exact half-up rounding, grouping, and a real minus sign. */
export function formatAmount(value: Decimal, decimals = 2, locale: Locale = "en"): string {
  if (!Number.isInteger(decimals) || decimals < 0 || decimals > 100) throw new RangeError("Invalid decimal places");
  const { negative, whole, fraction } = parts(value);
  let digits = BigInt(whole + fraction.padEnd(decimals, "0").slice(0, decimals));
  if ((fraction[decimals] ?? "0") >= "5") digits += BigInt(1);
  const result = String(digits).padStart(decimals + 1, "0");
  const integer = (decimals ? result.slice(0, -decimals) : result).replace(/\B(?=(\d{3})+(?!\d))/g, locale === "tr" ? "." : ",");
  return `${negative && digits !== BigInt(0) ? "−" : ""}${integer}${decimals ? `${locale === "tr" ? "," : "."}${result.slice(-decimals)}` : ""}`;
}

/** Input is percentage points: "2.75" becomes "2.75%", never a fraction. */
export function formatPercent(value: Decimal, decimals = 2, locale: Locale = "en"): string {
  const amount = formatAmount(value, decimals, locale);
  return locale === "tr" ? amount.startsWith("−") ? `−%${amount.slice(1)}` : `%${amount}` : `${amount}%`;
}

/** Preserve the source price precision unless a fixed precision is supplied. */
export function formatPrice(value: Decimal, decimals?: number, locale: Locale = "en"): string {
  return formatAmount(value, decimals ?? parts(value).fraction.length, locale);
}

/** Keep editable decimal drafts and their precision when the interface language changes. */
export function reformatInput(value: string, from: Locale, to: Locale): string {
  const canonical = from === "tr" ? value.replaceAll(".", "").replace(",", ".") : value.replaceAll(",", "");
  return /^[+−-]?\d+(?:\.\d+)?$/.test(canonical) ? formatPrice(canonical, undefined, to) : value;
}

/** Exact decimal multiplication for display labels; geometry may still use number ratios. */
export function decimalProduct(left: Decimal, right: Decimal): string {
  const a = parts(left), b = parts(right);
  const scale = a.fraction.length + b.fraction.length;
  const digits = String(BigInt(a.whole + a.fraction) * BigInt(b.whole + b.fraction)).padStart(scale + 1, "0");
  return `${a.negative !== b.negative ? "-" : ""}${scale ? `${digits.slice(0, -scale)}.${digits.slice(-scale)}` : digits}`;
}
