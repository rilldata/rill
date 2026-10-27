import type {
  MetricsViewSpecMeasure,
  V1MetricsViewAggregationMeasure,
} from "@rilldata/web-common/runtime-client";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import type { EphemeralMeasureDef } from "./types";

/**
 * Builds a synthetic metrics view spec measure for an ephemeral measure so UI
 * code (labels, formatters, tooltips, selectors) can treat it like any other
 * measure.
 */
export function ephemeralMeasureToSpecMeasure(
  def: EphemeralMeasureDef,
): MetricsViewSpecMeasure {
  return {
    name: def.name,
    displayName: def.displayName,
    expression: def.expression,
    // Description-driven tooltips show the calculation unless the user
    // described the measure.
    description: def.description || def.expression,
    formatPreset: def.formatPreset ?? (FormatPreset.HUMANIZE as string),
  };
}

/**
 * Appends synthetic spec measures for the given ephemeral measure
 * definitions to a list of spec measures.
 */
export function appendEphemeralSpecMeasures(
  measures: MetricsViewSpecMeasure[],
  defs: EphemeralMeasureDef[] | undefined,
): MetricsViewSpecMeasure[] {
  if (!defs?.length) return measures;
  return [...measures, ...defs.map(ephemeralMeasureToSpecMeasure)];
}

/**
 * Attaches the `expression` compute to request measures whose name matches a
 * ephemeral measure definition. Leaves other measures untouched.
 */
export function mapEphemeralMeasuresForRequest(
  measures: V1MetricsViewAggregationMeasure[],
  defs: EphemeralMeasureDef[] | undefined,
): V1MetricsViewAggregationMeasure[] {
  if (!defs?.length) return measures;
  const byName = new Map(defs.map((def) => [def.name, def]));
  return measures.map((measure) => {
    const def = measure.name ? byName.get(measure.name) : undefined;
    // Never attach an expression to a measure that already carries a compute
    // (e.g. a comparison accessor whose alias happens to match a definition).
    const hasCompute =
      measure.expression ||
      measure.count ||
      measure.countDistinct ||
      measure.comparisonValue ||
      measure.comparisonDelta ||
      measure.comparisonRatio ||
      measure.percentOfTotal ||
      measure.uri ||
      measure.comparisonTime;
    if (!def || hasCompute) return measure;
    return {
      ...measure,
      expression: {
        expression: def.expression,
        displayName: def.displayName,
      },
    };
  });
}

/**
 * Splits measure names for the time series API: names of spec measures go in
 * `measureNames`, ephemeral measures go in `measures` with their expression.
 */
export function splitTimeSeriesMeasures(
  names: string[],
  defs: EphemeralMeasureDef[] | undefined,
): {
  measureNames: string[];
  ephemeralMeasures: V1MetricsViewAggregationMeasure[] | undefined;
} {
  if (!defs?.length)
    return { measureNames: names, ephemeralMeasures: undefined };
  const byName = new Map(defs.map((def) => [def.name, def]));
  const measureNames: string[] = [];
  const measures: V1MetricsViewAggregationMeasure[] = [];
  for (const name of names) {
    const def = byName.get(name);
    if (def) {
      measures.push({
        name: def.name,
        expression: {
          expression: def.expression,
          displayName: def.displayName,
        },
      });
    } else {
      measureNames.push(name);
    }
  }
  return {
    measureNames,
    ephemeralMeasures: measures.length ? measures : undefined,
  };
}

/**
 * The spec measure a request measure is computed from: the base measure of a
 * comparison or percent-of-total compute, otherwise the measure itself.
 */
function requestMeasureBase(
  measure: V1MetricsViewAggregationMeasure,
): string | undefined {
  return (
    measure.comparisonValue?.measure ??
    measure.comparisonDelta?.measure ??
    measure.comparisonRatio?.measure ??
    measure.percentOfTotal?.measure ??
    measure.name
  );
}

/**
 * Flags request measures with `exact: true` when their measure, or the base
 * measure of their compute, is in `exactNames` (see `requestExactNames`).
 * The runtime then computes those measures, and the measures they reference,
 * with their `exact_expression`. Leaves other measures untouched.
 */
export function applyExactMeasures(
  measures: V1MetricsViewAggregationMeasure[],
  exactNames: ReadonlySet<string> | undefined,
): V1MetricsViewAggregationMeasure[] {
  if (!exactNames?.size) return measures;
  return measures.map((measure) => {
    const base = requestMeasureBase(measure);
    return base && exactNames.has(base) ? { ...measure, exact: true } : measure;
  });
}

/**
 * The measure names to send as `exactMeasureNames` on a time series request:
 * the requested names that are in `exactNames`. Undefined when there are none.
 */
export function exactTimeSeriesMeasureNames(
  names: string[],
  exactNames: ReadonlySet<string> | undefined,
): string[] | undefined {
  if (!exactNames?.size) return undefined;
  const result = names.filter((name) => exactNames.has(name));
  return result.length ? result : undefined;
}

/**
 * Recovers ephemeral measure definitions from request measures that carry an
 * `expression` compute, e.g. when editing a saved report outside an explore.
 */
export function ephemeralDefsFromRequestMeasures(
  measures: V1MetricsViewAggregationMeasure[] | undefined,
): EphemeralMeasureDef[] | undefined {
  const defs = (measures ?? []).flatMap((measure) =>
    measure.name && measure.expression?.expression
      ? [
          {
            name: measure.name,
            displayName: measure.expression.displayName || measure.name,
            expression: measure.expression.expression,
          },
        ]
      : [],
  );
  return defs.length ? defs : undefined;
}

/**
 * Returns the set of ephemeral measure names, for quick membership checks.
 */
export function ephemeralMeasureNameSet(
  defs: EphemeralMeasureDef[] | undefined,
): Set<string> {
  return new Set(defs?.map((def) => def.name) ?? []);
}
