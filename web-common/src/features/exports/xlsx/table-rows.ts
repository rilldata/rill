import type { ConnectError } from "@connectrpc/connect";
import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { createTableCellQuery } from "@rilldata/web-common/features/dashboards/pivot/pivot-data-store";
import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
import type {
  PivotDataRow,
  PivotDataState,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { V1MetricsViewAggregationResponse } from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import { derived, readable, type Readable } from "svelte/store";
import type { ExportTable } from "./payload";
import { flattenPivotData, isPlaceholderRow } from "./pivot-flatten";

// Rows a flat table exports at most. Flat tables page their rows in as the
// viewer scrolls, so the export queries them again up to this cap rather than
// exporting only the pages the viewer happened to see. The cap is the runtime's
// default limit on the rows of one interactive query
// (`rill.interactive_sql_row_limit`), which rejects any request above it, so no
// request here may ask for more.
export const MAX_FLAT_TABLE_ROWS = 10_000;

const DEFAULT_TIMEOUT_MS = 60_000;

// The table behind a table or pivot block. A flat table that has not loaded
// every row gets them queried again, up to the cap; a nested pivot exports the
// rows as the viewer sees them (expanded rows expanded, up to the row limit or
// the pages loaded), since its expansion lives in the browser. `moreRows` tells
// the reader when the source has more rows than the sheet.
export async function tableExportTable(
  component: BaseCanvasComponent,
  state: PivotDataState,
  config: PivotDataStoreConfig,
  timeoutMs = DEFAULT_TIMEOUT_MS,
): Promise<ExportTable> {
  if (config.isFlat && hasMoreRows(state, config)) {
    try {
      const rows = await loadFlatRows(
        component,
        config,
        MAX_FLAT_TABLE_ROWS,
        0,
        timeoutMs,
      );
      // A full page may be the whole table: one row past it settles that. If
      // that row can't be loaded, the sheet says it may not be complete.
      const moreRows =
        rows.length >= MAX_FLAT_TABLE_ROWS &&
        (await loadFlatRows(
          component,
          config,
          1,
          MAX_FLAT_TABLE_ROWS,
          timeoutMs,
        ).then(
          (next) => next.length > 0,
          () => true,
        ));
      return {
        ...flattenPivotData({ ...state, data: rows }, config),
        moreRows,
      };
    } catch (e) {
      // For example an instance with a lower row limit. The rows on screen
      // still make a valid sheet, marked as partial, so the workbook downloads.
      console.warn(
        `Failed to load the rows of table "${component.id}"; exporting the rows on screen`,
        e,
      );
    }
  }
  return {
    ...flattenPivotData(state, config),
    moreRows: hasMoreRows(state, config),
  };
}

// Whether the source has more top-level rows than the pivot state holds. The
// state only says so outright through a pending "show more" row; otherwise a
// loaded row count at the row limit, or on a page boundary before the store has
// seen the end, means more rows may follow.
export function hasMoreRows(
  state: PivotDataState,
  config: PivotDataStoreConfig,
): boolean {
  const loaded = state.data.filter((row) => !isPlaceholderRow(row)).length;
  if (loaded < state.data.length) return true;
  const rowLimit = config.pivot.rowLimit;
  if (rowLimit !== undefined && loaded >= rowLimit) return true;
  return (
    !state.reachedEndForRowData &&
    loaded > 0 &&
    loaded % NUM_ROWS_PER_PAGE === 0
  );
}

// Queries `limit` rows of a flat table from `offset`, sorted like the table,
// through the same query builder the table uses.
function loadFlatRows(
  component: BaseCanvasComponent,
  config: PivotDataStoreConfig,
  limit: number,
  offset: number,
  timeoutMs: number,
): Promise<PivotDataRow[]> {
  // The builder's type is a union of its store types; this is the one the flat table gets.
  const query = createTableCellQuery(
    {
      runtimeClient: component.parent.client,
      metricsViewName: derived(
        component.specStore,
        (spec) => (spec as { metrics_view?: string }).metrics_view ?? "",
      ),
      queryClient,
      enabled: readable(true),
    },
    config,
    undefined,
    {},
    [],
    String(limit),
    String(offset),
  ) as Readable<QueryObserverResult<
    V1MetricsViewAggregationResponse,
    ConnectError
  > | null>;
  return new Promise((resolve, reject) => {
    let settled = false;
    let unsubscribe = () => {};
    const finish = (done: () => void) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      // The subscription may still be in progress when the result arrives synchronously.
      queueMicrotask(unsubscribe);
      done();
    };
    const timer = setTimeout(
      () =>
        finish(() => reject(new Error("Timed out loading the table's rows"))),
      timeoutMs,
    );
    unsubscribe = query.subscribe((result) => {
      if (!result) return;
      if (result.isError) finish(() => reject(result.error));
      else if (result.isSuccess) {
        finish(() => resolve((result.data?.data ?? []) as PivotDataRow[]));
      }
    });
  });
}
