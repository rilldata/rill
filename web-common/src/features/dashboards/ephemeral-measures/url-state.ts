import type {
  MetricsViewSpecMeasure,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { ephemeralMeasureToSpecMeasure } from "./measure-mapping";
import { fromEphemeralMeasuresParam } from "./url-param";
import type { EphemeralMeasureDef } from "./types";
import {
  isReferenceableMeasure,
  validateEphemeralMeasureDef,
} from "./validation";

/**
 * Validates ephemeral measure definitions during URL state conversion.
 * `measures` is the explore's measure map (what an expression may reference);
 * names are reserved against every field of the metrics view, matching the
 * server's collision check, so a definition can never shadow a hidden
 * measure, a dimension or the time dimension. Returns the valid definitions
 * and human-readable labels for the dropped ones.
 */
export function validateEphemeralDefsAgainstSpec(
  defs: EphemeralMeasureDef[],
  measures: Map<string, MetricsViewSpecMeasure>,
  metricsView: V1MetricsViewSpec,
): { valid: EphemeralMeasureDef[]; invalidEntries: string[] } {
  const knownMeasureNames = new Set(
    [...measures.values()]
      .filter(isReferenceableMeasure)
      .map((m) => m.name as string),
  );
  const reservedNames = new Set<string>([
    ...(metricsView.measures ?? []).map((m) => m.name as string),
    ...(metricsView.dimensions ?? []).map(
      (d) => (d.name || d.column) as string,
    ),
    ...(metricsView.timeDimension ? [metricsView.timeDimension] : []),
  ]);
  const valid: EphemeralMeasureDef[] = [];
  const invalidEntries: string[] = [];
  for (const def of defs) {
    const error = validateEphemeralMeasureDef(
      def,
      knownMeasureNames,
      reservedNames,
    );
    if (error) {
      invalidEntries.push(`${def.name} (${error})`);
    } else {
      valid.push(def);
      // Later definitions may not reuse this name either.
      reservedNames.add(def.name);
    }
  }
  return { valid, invalidEntries };
}

/**
 * Parses and validates a `ephemeral` URL param value.
 */
export function parseAndValidateEphemeralParam(
  param: string,
  measures: Map<string, MetricsViewSpecMeasure>,
  metricsView: V1MetricsViewSpec,
): { valid: EphemeralMeasureDef[]; invalidEntries: string[] } {
  const { ephemeralMeasures, invalidEntries } =
    fromEphemeralMeasuresParam(param);
  const res = validateEphemeralDefsAgainstSpec(
    ephemeralMeasures,
    measures,
    metricsView,
  );
  return {
    valid: res.valid,
    invalidEntries: [...invalidEntries, ...res.invalidEntries],
  };
}

/**
 * Adds synthetic spec measures for the definitions into the measures map used
 * during URL state conversion, so every existing name validation (visible
 * measures, leaderboards, sort, pivot columns, formatting, ...) accepts
 * ephemeral measure names without per-param special cases.
 */
export function injectEphemeralMeasuresIntoMap(
  measures: Map<string, MetricsViewSpecMeasure>,
  defs: EphemeralMeasureDef[],
): void {
  for (const def of defs) {
    measures.set(def.name, ephemeralMeasureToSpecMeasure(def));
  }
}
