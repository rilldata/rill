import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { TimeUnit } from "vega-lite/types_unstable/timeunit.js";

export const timeGrainToVegaTimeUnitMap: Record<V1TimeGrain, TimeUnit> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "yearmonthdatehoursminutes",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "yearmonthdatehours",
  [V1TimeGrain.TIME_GRAIN_DAY]: "yearmonthdate",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "yearweek",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "yearmonth",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "yearquarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "yearmonthdate",
};

export function sanitizeValueForVega(value: unknown) {
  if (typeof value === "string") {
    // Escape all special characters including quotes, brackets, operators, etc.
    return value.replace(
      /[!@#$%^&*()+=\-[\]\\';,./{}|:<>?~]/g,
      (match) => `\\${match}`,
    );
  } else {
    return String(value);
  }
}

export function sanitizeValuesForSpec(values: unknown[]) {
  return values.map((value) => sanitizeValueForVega(value));
}

export function sanitizeFieldName(fieldName: string) {
  const sanitizedFieldName = Array.from(fieldName)
    .map((char) => {
      if (/[a-zA-Z0-9_$]/.test(char)) return char;
      return `_u${char.codePointAt(0)?.toString(16) ?? "0"}_`;
    })
    .join("");

  /**
   * Vega-Lite compiles custom formatType values as expression function calls.
   * Keep this value to a JavaScript/Vega identifier-safe subset so measure
   * names with spaces or operators can still be used as formatter names.
   */
  return `rill_${sanitizedFieldName || "field"}`;
}
