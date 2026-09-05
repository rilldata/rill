import {
  MetricsViewSpecMeasureType,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import { ComparisonModifierSuffixRegex } from "../pivot/types";
import { parseMeasureExpression } from "./expression-parser";
import type { EphemeralMeasureDef } from "./types";

// Maximum number of ephemeral measures per pivot.
export const MAX_EPHEMERAL_MEASURES = 10;

export const EPHEMERAL_MEASURE_NAME_REGEX = /^[a-zA-Z][a-zA-Z0-9_]*$/;

// Accessor suffixes appended to measure names when building requests:
// the pivot's `__delta_abs`-style suffixes (ComparisonModifierSuffixRegex) and
// the leaderboard/dimension-table `_prev`-style suffixes below. An ephemeral
// measure ending in one of these would collide with another measure's
// comparison accessor.
const RESERVED_NAME_SUFFIXES = [
  "_prev",
  "_delta",
  "_delta_perc",
  "_percent_of_total",
];

/**
 * Returns whether a measure may be referenced from an ephemeral measure's
 * expression. Window measures, measures with required dimensions and
 * time-comparison measures are excluded: the same views that filter those
 * spec measures out of their requests (totals, big numbers, ...) cannot
 * filter an ephemeral wrapper, so allowing the reference would produce
 * permanent query errors.
 */
export function isReferenceableMeasure(
  measure: MetricsViewSpecMeasure,
): boolean {
  return (
    !measure.window &&
    !measure.requiredDimensions?.length &&
    measure.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON
  );
}

/**
 * Validates an ephemeral measure's name (the query alias).
 * `reservedNames` should contain all metrics view field names (measures,
 * dimensions, the time dimension) plus the names of other ephemeral measures.
 */
export function validateEphemeralMeasureName(
  name: string,
  reservedNames: Set<string>,
): string | undefined {
  if (!EPHEMERAL_MEASURE_NAME_REGEX.test(name)) {
    return "name must start with a letter and contain only letters, numbers and underscores";
  }
  if (
    ComparisonModifierSuffixRegex.test(name) ||
    RESERVED_NAME_SUFFIXES.some((suffix) => name.endsWith(suffix))
  ) {
    return "name must not end with a comparison suffix";
  }
  if (name.includes("_rill_")) {
    return 'name must not contain "_rill_"';
  }
  if (reservedNames.has(name)) {
    return `"${name}" is already used by another field`;
  }
  return undefined;
}

/**
 * Validates a full ephemeral measure definition.
 * `knownMeasureNames` are the measure names an expression may reference
 * (the metrics view's measures available in this explore; ephemeral measures
 * cannot reference other ephemeral measures).
 */
export function validateEphemeralMeasureDef(
  def: EphemeralMeasureDef,
  knownMeasureNames: Set<string>,
  reservedNames: Set<string>,
): string | undefined {
  const nameError = validateEphemeralMeasureName(def.name, reservedNames);
  if (nameError) return nameError;

  if (def.displayName.trim() === "") {
    return "display name is required";
  }

  const parsed = parseMeasureExpression(def.expression);
  if (parsed.error) {
    return parsed.error.message;
  }
  for (const ref of parsed.refs) {
    if (!knownMeasureNames.has(ref)) {
      return `"${ref}" is not a measure in this dashboard`;
    }
  }
  return undefined;
}
