import type {
  MetricsViewSpecMeasure,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { ephemeralMeasureToSpecMeasure } from "./measure-mapping";
import { fromEphemeralMeasuresParam } from "./url-param";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
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
 * Parses and validates an `adhoc_m` URL param value.
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

/**
 * Names of the ephemeral measures any page of the explore state uses: visible,
 * leaderboard and sort measures, the TDD measure and pivot chips.
 */
export function ephemeralMeasureNamesInUse(
  exploreState: Partial<ExploreState>,
): Set<string> {
  const inUse = new Set<string>([
    ...(exploreState.visibleMeasures ?? []),
    ...(exploreState.leaderboardMeasureNames ?? []),
    ...(exploreState.pivot?.rows ?? []).map((chip) => chip.id),
    ...(exploreState.pivot?.columns ?? []).map((chip) => chip.id),
    ...(exploreState.pivot?.sorting ?? []).map((sort) => sort.id),
  ]);
  if (exploreState.leaderboardSortByMeasureName) {
    inUse.add(exploreState.leaderboardSortByMeasureName);
  }
  if (exploreState.tdd?.expandedMeasureName) {
    inUse.add(exploreState.tdd.expandedMeasureName);
  }
  return inUse;
}

/**
 * Returns the definitions the active page actually shows: visible and
 * leaderboard measures plus the sort measure on the explore page, the expanded
 * measure on the TDD page, and the row, column and sort chips on the pivot
 * page. Only these go into the URL, matching the other params each page
 * emits; the rest stay in the per-metrics-view library (see `library.ts`) and
 * in the per-view session store, so hidden definitions are never lost and
 * shared links do not grow with every definition the user has ever created.
 */
export function referencedEphemeralMeasures(
  exploreState: Partial<ExploreState>,
): EphemeralMeasureDef[] {
  const defs = exploreState.ephemeralMeasures ?? [];
  if (!defs.length) return defs;

  const referenced = new Set<string>();
  switch (exploreState.activePage) {
    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      if (exploreState.tdd?.expandedMeasureName) {
        referenced.add(exploreState.tdd.expandedMeasureName);
      }
      break;
    case DashboardState_ActivePage.PIVOT:
      for (const chip of [
        ...(exploreState.pivot?.rows ?? []),
        ...(exploreState.pivot?.columns ?? []),
      ]) {
        referenced.add(chip.id);
      }
      for (const sort of exploreState.pivot?.sorting ?? []) {
        referenced.add(sort.id);
      }
      break;
    default:
      // Explore and dimension table pages.
      if (exploreState.allMeasuresVisible) return defs;
      for (const name of [
        ...(exploreState.visibleMeasures ?? []),
        ...(exploreState.leaderboardMeasureNames ?? []),
      ]) {
        referenced.add(name);
      }
      if (exploreState.leaderboardSortByMeasureName) {
        referenced.add(exploreState.leaderboardSortByMeasureName);
      }
  }
  return defs.filter((def) => referenced.has(def.name));
}
