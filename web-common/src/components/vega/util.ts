import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { TimeUnit } from "vega-lite/build/src/timeunit";

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
  const specialCharactersRemoved = sanitizeValueForVega(fieldName);
  const sanitizedFieldName = specialCharactersRemoved.replace(" ", "__");

  /**
   * Add a prefix to the beginning of the field
   * name to avoid variables starting with a special
   * character or number.
   */
  return `rill_${sanitizedFieldName}`;
}
