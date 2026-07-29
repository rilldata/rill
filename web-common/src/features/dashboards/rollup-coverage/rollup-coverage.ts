import {
  V1TimeGrain,
  type V1MetricsViewSpecRollup,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";

/**
 * Rollup coverage warnings.
 *
 * When a metrics view has rollups, different widgets on the same dashboard can be served
 * from different physical tables with different time coverage (e.g. a base table with
 * 4 months of retention and a rollup with 2 years). A widget whose query references a
 * dimension/measure only present in the narrow table silently shows truncated data while
 * the rest of the dashboard shows the full range.
 *
 * The trigger implemented here: warn on a widget when its serving table's coverage starts
 * later than the widest coverage among all serving tables currently in use on the dashboard,
 * within the requested time range. Comparing against tables in use (rather than all tables)
 * keeps a uniformly-truncated dashboard silent: if every widget routes to the narrow table,
 * the dashboard is internally consistent and no warning is shown.
 *
 * Only the start side of the range is checked: rollup min/max timestamps are grain-truncated
 * bucket starts, so end-side comparisons are polluted by up to one bucket of truncation,
 * while retention differences (the case that matters) show up at the start. For the same
 * reason, a start-side gap must exceed one bucket of the widest table's grain to count:
 * a monthly rollup built from base data starting Jan 15 reports min = Jan 1, and that
 * sub-bucket difference is a truncation artifact, not extra data.
 */

/** Key used for the metrics view's base table in coverage maps. Matches the empty
 * `servingTable` returned by query APIs when no rollup was selected. */
export const BASE_TABLE_KEY = "";

export interface TableCoverage {
  minMs: number;
  maxMs: number;
}

export interface CoverageWarning {
  /** Timestamp (ms) where the widget's serving table's data actually starts. */
  servedStartMs: number;
  /** Timestamp (ms) where data is available from other tables in use, clamped to the requested range. */
  availableStartMs: number;
  /** The widget's serving table ("" = base table). */
  servingTable: string;
}

/**
 * Builds a table name → coverage map from a MetricsViewTimeRange response.
 * The base table is keyed by BASE_TABLE_KEY. Returns undefined when the response
 * has no rollups, in which case no warnings are possible.
 */
export function coverageFromTimeRangeResponse(
  resp: V1MetricsViewTimeRangeResponse | undefined,
): Map<string, TableCoverage> | undefined {
  const rollups = resp?.rollupTimeRanges;
  if (!rollups || Object.keys(rollups).length === 0) return undefined;

  const coverage = new Map<string, TableCoverage>();
  if (resp.timeRangeSummary?.min && resp.timeRangeSummary?.max) {
    coverage.set(BASE_TABLE_KEY, {
      minMs: Date.parse(resp.timeRangeSummary.min),
      maxMs: Date.parse(resp.timeRangeSummary.max),
    });
  }
  for (const [table, summary] of Object.entries(rollups)) {
    if (!summary.min || !summary.max) continue;
    coverage.set(table, {
      minMs: Date.parse(summary.min),
      maxMs: Date.parse(summary.max),
    });
  }
  return coverage;
}

/**
 * Upper bound of one grain bucket in milliseconds, used as tolerance for
 * grain-truncation artifacts in rollup min timestamps.
 */
export function grainPeriodMs(grain: V1TimeGrain | undefined): number {
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_MILLISECOND:
      return 1;
    case V1TimeGrain.TIME_GRAIN_SECOND:
      return 1000;
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return 7 * 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return 31 * 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_QUARTER:
      return 92 * 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return 366 * 24 * 60 * 60 * 1000;
    default:
      return 0;
  }
}

/**
 * Computes a coverage warning for a single widget, or undefined if the widget's
 * serving table covers the requested range as well as any other table in use.
 *
 * @param servingTable the widget's serving table ("" = base table)
 * @param coverage per-table coverage from coverageFromTimeRangeResponse
 * @param tablesInUse serving tables of all widgets currently on the dashboard (including this one)
 * @param rollupGrains rollup table name → time grain, from the metrics view spec
 * @param requestedStartMs start of the requested time range; when a comparison range is
 *        active, pass the earlier of the two starts since the query must cover both
 */
export function computeCoverageWarning(
  servingTable: string,
  coverage: Map<string, TableCoverage> | undefined,
  tablesInUse: Iterable<string>,
  rollupGrains: Map<string, V1TimeGrain>,
  requestedStartMs: number,
): CoverageWarning | undefined {
  if (!coverage) return undefined;

  const served = coverage.get(servingTable);
  if (!served) return undefined;

  // Find the widest (earliest-starting) coverage among tables in use
  let widestMin = Infinity;
  let widestTable: string | undefined;
  for (const table of tablesInUse) {
    const c = coverage.get(table);
    if (c && c.minMs < widestMin) {
      widestMin = c.minMs;
      widestTable = table;
    }
  }
  if (widestTable === undefined || widestTable === servingTable)
    return undefined;

  // The widget is missing data from max(requested start, widest coverage start) up to
  // its own coverage start. Only warn when the gap exceeds one grain bucket of the
  // widest table (grain-truncation tolerance; 0 for the base table).
  const availableStartMs = Math.max(requestedStartMs, widestMin);
  const tolerance = grainPeriodMs(rollupGrains.get(widestTable));
  if (served.minMs <= availableStartMs + tolerance) return undefined;

  return {
    servedStartMs: served.minMs,
    availableStartMs,
    servingTable,
  };
}

/**
 * Computes the effective requested start (ms) of a widget's query for coverage checks:
 * the earlier of the primary and comparison range starts, since the serving table must
 * cover both. An absent start (all-time query) resolves to -Infinity so the check falls
 * back to comparing table coverage directly.
 */
export function requestedStartMs(
  start: string | undefined,
  comparisonStart?: string,
): number {
  const startMs = start ? Date.parse(start) : -Infinity;
  const comparisonStartMs = comparisonStart
    ? Date.parse(comparisonStart)
    : Infinity;
  return Math.min(startMs, comparisonStartMs);
}

/** Builds a rollup table name → time grain map from the metrics view spec's rollups. */
export function rollupGrainsFromSpec(
  rollups: V1MetricsViewSpecRollup[] | undefined,
): Map<string, V1TimeGrain> {
  const grains = new Map<string, V1TimeGrain>();
  for (const rollup of rollups ?? []) {
    if (rollup.table && rollup.timeGrain)
      grains.set(rollup.table, rollup.timeGrain);
  }
  return grains;
}
