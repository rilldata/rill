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
    // Surfaces the calculation in description-driven tooltips.
    description: def.expression,
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
 * Returns the set of ephemeral measure names, for quick membership checks.
 */
export function ephemeralMeasureNameSet(
  defs: EphemeralMeasureDef[] | undefined,
): Set<string> {
  return new Set(defs?.map((def) => def.name) ?? []);
}
