import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import {
  StateManagers,
  memoizeMetricsStore,
} from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import {
  createQueryServiceMetricsViewAggregation,
  V1MetricsViewAggregationResponse,
} from "@rilldata/web-common/runtime-client";
import { createSparkline } from "./sparkline";
import { transposeArray } from "./util";
import {
  FormatPreset,
  humanizeDataType,
} from "@rilldata/web-common/features/dashboards/humanize-numbers";
import { createTimeFormat } from "@rilldata/web-common/components/data-graphic/utils";

export type TimeDimensionDataState = {
  isFetching: boolean;
  comparing: "dimension" | "time" | "none";
  data?: unknown[];
  timeFormatter: (v: Date) => string;
};

export type TimeSeriesDataStore = Readable<TimeDimensionDataState>;

function createTimeDimensionAggregation(
  ctx: StateManagers
): CreateQueryResult<V1MetricsViewAggregationResponse> {
  return derived(
    [
      ctx.runtime,
      ctx.metricsViewName,
      useTimeControlStore(ctx),
      ctx.dashboardStore,
    ],
    ([runtime, metricsViewName, timeControls, dashboard], set) =>
      createQueryServiceMetricsViewAggregation(
        runtime.instanceId,
        metricsViewName,
        {
          dimensions: [{ name: dashboard?.selectedComparisonDimension }],
          measures: [{ name: dashboard?.expandedMeasureName }],
          filter: dashboard?.filters,
          timeStart: timeControls.adjustedStart,
          timeEnd: timeControls.adjustedEnd,
          limit: "1000", //TODO Use blocks later
        },
        {
          query: {
            enabled: !!timeControls.ready && !!ctx.dashboardStore,
            queryClient: ctx.queryClient,
          },
        }
      ).subscribe(set)
  );
}

function prepareDimensionData(data, measureName) {
  if (!data) return;

  const rowCount = data?.length;
  // When using row headers, be careful not to accidentally merge cells
  const rowHeaderData = data?.map((row) => [
    // Dim
    {
      value: row?.value,
    },
    // Measure total
    {
      value: row?.total,
      spark: createSparkline(row?.data, (v) => v[measureName]),
    },
    // Measure percent of total
    {
      value: 44 + "%",
    },
  ]);

  const columnCount = data?.[0]?.data?.length;
  const columnHeaderData = data?.[0]?.data?.map((v) => [{ value: v.ts }]);

  /* 
    Important: regular-table expects body data in columnar format,
    aka an array of arrays where outer array is the columns,
    inner array is the row values for a specific column
  */
  const body = data?.map((v) => v?.data?.map((v) => v[measureName]));
  const columnarBody = transposeArray(body, rowCount, columnCount);

  return {
    rowCount,
    rowHeaderData,
    columnCount,
    columnHeaderData,
    body: columnarBody,
  };
}

function prepareTimeData(data, measureName, hasTimeComparison) {
  if (!data) return;

  let rowHeaderData = [];
  rowHeaderData.push([
    { value: "Total" },
    {
      value: 228.4,
      spark: createSparkline(data, (v) => v[measureName]),
    },
    { value: 44 + "%" },
  ]);

  const body = [];
  body.push(data?.map((v) => v[measureName]));

  const columnCount = data?.length;
  const columnHeaderData = data?.map((v) => [{ value: v.ts }]);

  if (hasTimeComparison) {
    rowHeaderData = rowHeaderData.concat([
      [
        { value: "Previous" },
        {
          value: 128.4,
          spark: createSparkline(data, (v) => v[`comparison.${measureName}`]),
        },
        { value: 24 + "%" },
      ],
      [{ value: "Percentage Change" }],
      [{ value: "Absolute Change" }],
    ]);

    console.log(rowHeaderData);
    body.push(data?.map((v) => v[`comparison.${measureName}`]));

    body.push(
      data?.map((v) => {
        const comparisonValue = v[`comparison.${measureName}`];
        const currentValue = v[measureName];
        const comparisonPercChange =
          comparisonValue && currentValue !== undefined && currentValue !== null
            ? (currentValue - comparisonValue) / comparisonValue
            : undefined;
        return humanizeDataType(comparisonPercChange, FormatPreset.PERCENTAGE);
      })
    );

    body.push(
      data?.map((v) => {
        const comparisonValue = v[`comparison.${measureName}`];
        const currentValue = v[measureName];
        return comparisonValue &&
          currentValue !== undefined &&
          currentValue !== null
          ? currentValue - comparisonValue
          : undefined;
      })
    );
  }

  const rowCount = rowHeaderData.length;
  const columnarBody = transposeArray(body, rowCount, columnCount);

  return {
    rowCount,
    rowHeaderData,
    columnCount,
    columnHeaderData,
    body: columnarBody,
  };
}

export function createTimeDimensionDataStore(ctx: StateManagers) {
  return derived(
    [ctx.dashboardStore, useTimeControlStore(ctx), useTimeSeriesDataStore(ctx)],
    ([dashboardStore, timeControls, timeSeries]) => {
      const measureName = dashboardStore?.expandedMeasureName;
      const timeFormatter = createTimeFormat([
        new Date(timeControls?.adjustedStart),
        new Date(timeControls?.adjustedEnd),
      ])[0];

      let comparing;
      let data;
      if (dashboardStore?.selectedComparisonDimension) {
        comparing = "dimension";

        // TODO: Fix types
        const allFetched = timeSeries?.dimensionTableData?.every(
          (v) => !v?.isFetching
        );

        if (allFetched) {
          data = prepareDimensionData(
            timeSeries?.dimensionTableData,
            measureName
          );
        }
      } else {
        comparing = timeControls.showComparison ? "time" : "none";
        data = prepareTimeData(
          timeSeries?.timeSeriesData,
          measureName,
          comparing === "time"
        );
      }

      return { comparing, data, timeFormatter };
    }
  ) as TimeSeriesDataStore;
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeDimensionDataStore =
  memoizeMetricsStore<TimeSeriesDataStore>((ctx: StateManagers) =>
    createTimeDimensionDataStore(ctx)
  );
