import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { mapTimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import type { TimeRangeString } from "@rilldata/web-common/lib/time/types";
import {
  type V1MetricsViewAggregationRequest,
  V1TimeGrain,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { derived, get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import type { StateManagers } from "../state-managers/state-managers";
import { getPivotConfig } from "./pivot-data-config";
import { prepareMeasureForComparison } from "./pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  PivotChipType,
  type PivotColumns,
  type PivotRows,
  type PivotState,
} from "./types";

export function getPivotExportArgs(ctx: StateManagers) {
  return derived(
    [
      ctx.metricsViewName,
      ctx.validSpecStore,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
      getPivotConfig(ctx),
      ctx.selectors.pivot.rows,
      ctx.selectors.pivot.columns,
    ],
    ([
      metricsViewName,
      validSpecStore,
      timeControlState,
      dashboardState,
      configStore,
      rows,
      columns,
    ]) => {
      if (!validSpecStore.data?.explore || !timeControlState.ready)
        return undefined;

      const enableComparison = configStore.enableComparison;
      const comparisonTime = configStore.comparisonTime;
      const pivotState = configStore.pivot;

      const metricsViewSpec = validSpecStore.data?.metricsView ?? {};
      const exploreSpec = validSpecStore.data?.explore ?? {};
      const timeRange = mapTimeRange(
        timeControlState,
        dashboardState.selectedTimezone,
        exploreSpec,
      );
      if (!timeRange) return undefined;

      return getPivotAggregationRequest(
        metricsViewName,
        metricsViewSpec.timeDimension ?? "",
        dashboardState,
        timeRange,
        rows,
        columns,
        enableComparison,
        comparisonTime,
        pivotState,
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
  enableComparison: boolean,
  comparisonTime: TimeRangeString | undefined,
  pivotState: PivotState,
): undefined | V1MetricsViewAggregationRequest {
  const measures = columns.measure.flatMap((m) => {
    const measureName = m.id;
    const group = [{ name: measureName }];

    if (enableComparison) {
      group.push(
        { name: `${measureName}${COMPARISON_DELTA}` },
        { name: `${measureName}${COMPARISON_PERCENT}` },
      );
    }

    return group;
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
    return {
      name: d.alias ? d.alias : d.name,
      desc: pivotState.sorting.find((s) => s.id === d.name)?.desc ?? false,
    };
  });

  return {
    instanceId: get(runtime).instanceId,
    metricsView,
    timeRange,
    comparisonTimeRange: comparisonTime,
    measures: enableComparison
      ? prepareMeasureForComparison(measures)
      : measures,
    dimensions: allDimensions,
    where: sanitiseExpression(
      mergeDimensionAndMeasureFilters(
        dashboardState.whereFilter,
        dashboardState.dimensionThresholdFilters,
      ),
      undefined,
    ),
    pivotOn,
    sort,
    offset: "0",
  };
}
