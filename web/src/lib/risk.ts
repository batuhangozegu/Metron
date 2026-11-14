export type RiskState = "safe" | "caution" | "limit";

/** Caution includes its threshold; a limit is exceeded only strictly above it. Zero capacity fails closed. */
export function riskState(value: number, limit: number, cautionAt = 0.75): RiskState {
  if (!Number.isFinite(value) || !Number.isFinite(limit) || limit <= 0 || value < 0) return "limit";
  if (!Number.isFinite(cautionAt) || cautionAt < 0 || cautionAt > 1) throw new RangeError("Invalid caution fraction");
  if (value > limit) return "limit";
  return compareDecimals(String(value), decimalProduct(String(limit), String(cautionAt))) >= 0 ? "caution" : "safe";
}
import { compareDecimals, decimalProduct } from "@/lib/format";

