import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
import type {
  V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { onDestroy } from "svelte";
import { derived, get, writable, type Readable } from "svelte/store";
import {
  BASE_TABLE_KEY,
  coverageFromTimeRangeResponse,
  rollupGrainsFromSpec,
  type TableCoverage,
} from "./rollup-coverage";

/**
 * Tracks which physical table served each widget's query on a dashboard, along with
 * per-table time coverage, so widgets can warn when they show less data than the rest
 * of the dashboard. See rollup-coverage.ts for the warning semantics.
 *
 * One instance is created per dashboard in createStateManagers. Widgets report their
 * serving table via registerWidgetServingTable and compute their own warning with
 * computeCoverageWarning, supplying the time range they queried with.
 */
export class RollupCoverageStore {
  /** Per-table time coverage; undefined when the metrics view has no rollups. */
  readonly coverage: Readable<Map<string, TableCoverage> | undefined>;
  /** Rollup table name → time grain, for grain-truncation tolerance. */
  readonly rollupGrains: Readable<Map<string, V1TimeGrain>>;
  /** Serving tables of all widgets currently on the dashboard. */
  readonly tablesInUse: Readable<Set<string>>;

  private servingTables = writable(new Map<string, string>());

  constructor(
    timeRangeSummaryStore: Readable<
      QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>
    >,
    validSpecStore: Readable<
      QueryObserverResult<ExploreValidSpecResponse, Error>
    >,
  ) {
    this.coverage = derived(timeRangeSummaryStore, ($summary) =>
      coverageFromTimeRangeResponse($summary.data),
    );
    this.rollupGrains = derived(validSpecStore, ($spec) =>
      rollupGrainsFromSpec($spec.data?.metricsView?.rollups),
    );
    this.tablesInUse = derived(
      this.servingTables,
      ($tables) => new Set($tables.values()),
    );
  }

  setServingTable(widgetId: string, servingTable: string | undefined) {
    const current = get(this.servingTables);
    if (current.get(widgetId) === servingTable) return;
    if (servingTable === undefined && !current.has(widgetId)) return;

    this.servingTables.update((tables) => {
      const next = new Map(tables);
      if (servingTable === undefined) next.delete(widgetId);
      else next.set(widgetId, servingTable);
      return next;
    });
  }
}

/**
 * Normalizes the servingTable field of a query response. Proto JSON omits empty strings,
 * so a base-served query arrives with the field absent: a successful response without it
 * means the base table. Returns undefined while there is no response yet.
 */
export function servingTableOf(
  data: { servingTable?: string } | undefined,
): string | undefined {
  return data ? (data.servingTable ?? BASE_TABLE_KEY) : undefined;
}

/**
 * Returns an updater that reports a widget's serving table to the store, handling
 * widget id changes and deregistration on component destroy. Must be called during
 * component initialization.
 */
export function registerWidgetServingTable(
  store: RollupCoverageStore,
): (widgetId: string, servingTable: string | undefined) => void {
  let lastWidgetId: string | undefined;

  onDestroy(() => {
    if (lastWidgetId !== undefined)
      store.setServingTable(lastWidgetId, undefined);
  });

  return (widgetId, servingTable) => {
    if (lastWidgetId !== undefined && lastWidgetId !== widgetId) {
      store.setServingTable(lastWidgetId, undefined);
    }
    lastWidgetId = widgetId;
    store.setServingTable(widgetId, servingTable);
  };
}
