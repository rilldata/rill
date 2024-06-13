import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { mapTimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import {
  V1ExportFormat,
  V1TimeGrain,
  createQueryServiceExport,
  V1TimeRange,
  V1MetricsViewAggregationRequest,
} from "@rilldata/web-common/runtime-client";
import { derived, get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import type { StateManagers } from "../state-managers/state-managers";
import { PivotChipType, PivotColumns, PivotRows } from "./types";

export default async function exportPivot({
  ctx,
  query,
  format,
  timeDimension,
}: {
  ctx: StateManagers;
  query: ReturnType<typeof createQueryServiceExport>;
  format: V1ExportFormat;
  timeDimension: string | undefined;
}) {
  const instanceId = get(runtime).instanceId;
  const metricsViewName = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const selectedTimeRange = get(
    ctx.selectors.timeRangeSelectors.selectedTimeRangeState,
  );
  const rows = get(ctx.selectors.pivot.rows);
  const columns = get(ctx.selectors.pivot.columns);

  const timeRange = {
    start: selectedTimeRange?.start.toISOString(),
    end: selectedTimeRange?.end.toISOString(),
  };

  const request = getPivotAggregationRequest(
    metricsViewName,
    timeDimension ?? "",
    dashboard,
    timeRange,
    rows,
    columns,
  );

  const result = await get(query).mutateAsync({
    instanceId,
    data: {
      format,
      query: {
        metricsViewAggregationRequest: request,
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}

export function getPivotExportArgs(ctx: StateManagers) {
  return derived(
    [
      ctx.metricsViewName,
      useMetricsView(ctx),
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      ctx.selectors.pivot.rows,
      ctx.selectors.pivot.columns,
    ],
    ([
      metricsViewName,
      metricsView,
      timeControlState,
      dashboardState,
      rows,
      columns,
    ]) => {
      const metricsViewSpec = metricsView.data ?? {};
      const timeRange = mapTimeRange(timeControlState, metricsViewSpec);
      if (!timeRange) return undefined;

      return getPivotAggregationRequest(
        metricsViewName,
        metricsViewSpec.timeDimension ?? "",
        dashboardState,
        timeRange,
        rows,
        columns,
      );
    },
  );
}

export function getPivotAggregationRequest(
  metricsView: string,
  timeDimension: string,
  dashboardState: MetricsExplorerEntity,
  timeRange: V1TimeRange,
  rows: PivotRows,
  columns: PivotColumns,
): undefined | V1MetricsViewAggregationRequest {
  const measures = columns.measure.map((m) => {
    return {
      name: m.id,
    };
  });

  const allDimensions = [...rows.dimension, ...columns.dimension].map((d) =>
    d.type === PivotChipType.Time
      ? {
          name: timeDimension,
          timeGrain: d.id as V1TimeGrain,
          timeZone: dashboardState.selectedTimezone,
          alias: `Time ${d.title}`,
        }
      : {
          name: d.id,
        },
  );

  const pivotOn = columns.dimension.map((d) =>
    d.type === PivotChipType.Time ? `Time ${d.title}` : d.id,
  );

  const rowDimensions = [...rows.dimension].map((d) =>
    d.type === PivotChipType.Time
      ? {
          name: timeDimension,
          timeGrain: d.id as V1TimeGrain,
          timeZone: dashboardState.selectedTimezone,
          alias: `Time ${d.title}`,
        }
      : {
          name: d.id,
        },
  );

  // Sort by the dimensions in the pivot's rows
  const sort = rowDimensions.map((d) => {
    if (d.alias) {
      return {
        name: d.alias,
        desc: false,
      };
    }
    return {
      name: d.name,
      desc: false,
    };
  });

  return {
    instanceId: get(runtime).instanceId,
    metricsView,
    timeRange,
    measures,
    dimensions: allDimensions,
    where: sanitiseExpression(mergeMeasureFilters(dashboardState), undefined),
    pivotOn,
    sort,
    offset: "0",
  };
}
