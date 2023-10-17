import { derived, type Readable } from "svelte/store";
import {
  StateManagers,
  memoizeMetricsStore,
} from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { createSparkline } from "./sparkline";
import { transposeArray } from "./util";
import {
  FormatPreset,
  humanizeDataType,
} from "@rilldata/web-common/features/dashboards/humanize-numbers";
import { createTimeFormat } from "@rilldata/web-common/components/data-graphic/utils";
import { getTimeWidth } from "@rilldata/web-common/lib/time/transforms";
import {
  DEFAULT_TIME_RANGES,
  TIME_COMPARISON,
  TIME_GRAIN,
} from "@rilldata/web-common/lib/time/config";
import { durationToMillis } from "@rilldata/web-common/lib/time/grains";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

export type TimeDimensionDataState = {
  isFetching: boolean;
  comparing: "dimension" | "time" | "none";
  data?: unknown[];
  timeFormatter: (v: Date) => string;
};

export type TimeSeriesDataStore = Readable<TimeDimensionDataState>;

/***
 * Add totals row from time series data
 * Add rest of dimension values from dimension table data
 * Transpose the data to columnar format
 */
function prepareDimensionData(
  totalsData,
  data,
  columnCount: number,
  total: number,
  unfilteredTotal: number,
  measure: MetricsViewSpecMeasureV2,
  selectedValues: string[]
) {
  if (!data) return;

  const formatPreset =
    (measure?.format as FormatPreset) ?? FormatPreset.HUMANIZE;
  const measureName = measure?.name;
  const validPercentOfTotal = measure?.validPercentOfTotal;

  const rowCount = data?.length + 1;

  const totalsRow = [
    { value: "Total" },
    {
      value: humanizeDataType(total, formatPreset),
      spark: createSparkline(totalsData, (v) => v[measureName]),
    },
  ];

  let fixedColCount = 2;
  if (validPercentOfTotal) {
    fixedColCount = 3;
    const percOfTotal = total / unfilteredTotal;
    totalsRow.push({
      value: isNaN(percOfTotal)
        ? "...%"
        : humanizeDataType(percOfTotal, FormatPreset.PERCENTAGE),
    });
  }

  let rowHeaderData = [totalsRow];

  rowHeaderData = rowHeaderData.concat(
    data?.map((row) => {
      const dataRow = [
        { value: row?.value },
        {
          value: row?.total ? humanizeDataType(row?.total, formatPreset) : null,
          spark: createSparkline(row?.data, (v) => v[measureName]),
        },
      ];
      if (validPercentOfTotal) {
        const percOfTotal = row?.total / unfilteredTotal;
        dataRow.push({
          value: isNaN(percOfTotal)
            ? "...%"
            : humanizeDataType(percOfTotal, FormatPreset.PERCENTAGE),
        });
      }
      return dataRow;
    })
  );

  let columnHeaderData = new Array(columnCount).fill([{ value: null }]);

  if (data?.[0]?.data) {
    columnHeaderData = data?.[0]?.data
      ?.slice(1, -1)
      .map((v) => [{ value: v.ts }]);
  }

  let body = [
    totalsData
      ?.slice(1, -1)
      ?.map((v) =>
        v[measureName] ? humanizeDataType(v[measureName], formatPreset) : null
      ) || [],
  ];

  body = body?.concat(
    data?.map((v) => {
      if (v.isFetching) return new Array(columnCount).fill(undefined);
      return v?.data
        ?.slice(1, -1)
        ?.map((v) =>
          v[measureName] ? humanizeDataType(v[measureName], formatPreset) : null
        );
    })
  );
  /* 
    Important: regular-table expects body data in columnar format,
    aka an array of arrays where outer array is the columns,
    inner array is the row values for a specific column
  */
  const columnarBody = transposeArray(body, rowCount, columnCount);

  return {
    rowCount,
    fixedColCount,
    rowHeaderData,
    columnCount,
    columnHeaderData,
    body: columnarBody,
    selectedValues: selectedValues,
  };
}

/***
 * Add totals row from time series data
 * Add Current, Previous, Percentage Change, Absolute Change rows for time comparison
 * Transpose the data to columnar format
 */
function prepareTimeData(
  data,
  columnCount: number,
  total: number,
  comparisonTotal: number,
  currentLabel: string,
  comparisonLabel: string,
  measure: MetricsViewSpecMeasureV2,
  hasTimeComparison
) {
  if (!data) return;

  const formatPreset =
    (measure?.format as FormatPreset) ?? FormatPreset.HUMANIZE;
  const measureName = measure?.name;

  const columnHeaderData = data?.slice(1, -1)?.map((v) => [{ value: v.ts }]);
  let rowHeaderData = [];
  rowHeaderData.push([
    { value: "Total" },
    {
      value: humanizeDataType(total, formatPreset),
      spark: createSparkline(data, (v) => v[measureName]),
    },
  ]);

  const body = [];
  body.push(
    data
      ?.slice(1, -1)
      ?.map((v) =>
        v[measureName] ? humanizeDataType(v[measureName], formatPreset) : null
      )
  );

  if (hasTimeComparison) {
    rowHeaderData = rowHeaderData.concat([
      [
        { value: currentLabel },
        {
          value: humanizeDataType(total, formatPreset),
          spark: createSparkline(data, (v) => v[measureName]),
        },
      ],
      [
        { value: comparisonLabel },
        {
          value: humanizeDataType(comparisonTotal, formatPreset),
          spark: createSparkline(data, (v) => v[`comparison.${measureName}`]),
        },
      ],
      [{ value: "Percentage Change" }],
      [{ value: "Absolute Change" }],
    ]);

    // Push current range
    body.push(
      data
        ?.slice(1, -1)
        ?.map((v) =>
          v[measureName] ? humanizeDataType(v[measureName], formatPreset) : null
        )
    );

    body.push(
      data?.map((v) =>
        v[`comparison.${measureName}`]
          ? humanizeDataType(v[`comparison.${measureName}`], formatPreset)
          : null
      )
    );

    // Push percentage change
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

    // Push absolute change
    body.push(
      data?.map((v) => {
        const comparisonValue = v[`comparison.${measureName}`];
        const currentValue = v[measureName];
        const change =
          comparisonValue && currentValue !== undefined && currentValue !== null
            ? currentValue - comparisonValue
            : undefined;

        return humanizeDataType(change, formatPreset);
      })
    );
  }

  const rowCount = rowHeaderData.length;
  const columnarBody = transposeArray(body, rowCount, columnCount);

  return {
    rowCount,
    fixedColCount: 2,
    rowHeaderData,
    columnCount,
    columnHeaderData,
    body: columnarBody,
    selectedValues: [],
  };
}

export function createTimeDimensionDataStore(ctx: StateManagers) {
  return derived(
    [
      ctx.dashboardStore,
      useMetaQuery(ctx),
      useTimeControlStore(ctx),
      useTimeSeriesDataStore(ctx),
    ],
    ([dashboardStore, metricsView, timeControls, timeSeries]) => {
      const measureName = dashboardStore?.expandedMeasureName;
      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const timeFormatter = createTimeFormat([
        new Date(timeControls?.adjustedStart),
        new Date(timeControls?.adjustedEnd),
      ])[0];
      const interval = timeControls?.selectedTimeRange?.interval;
      const intervalWidth = durationToMillis(TIME_GRAIN[interval]?.duration);
      const total = timeSeries?.total && timeSeries?.total[measureName];
      const unfilteredTotal =
        timeSeries?.unfilteredTotal && timeSeries?.unfilteredTotal[measureName];
      const comparisonTotal =
        timeSeries?.comparisonTotal && timeSeries?.comparisonTotal[measureName];

      // Compute columnCount
      const columnCount =
        getTimeWidth(
          new Date(timeControls?.adjustedStart),
          new Date(timeControls?.adjustedEnd)
        ) /
          intervalWidth -
        2;

      const measure = metricsView?.data?.measures?.find(
        (m) => m.name === measureName
      );

      let comparing;
      let data;
      if (dimensionName) {
        comparing = "dimension";

        const excludeMode =
          dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ??
          false;
        const selectedValues =
          ((excludeMode
            ? dashboardStore?.filters.exclude.find(
                (d) => d.name === dimensionName
              )?.in
            : dashboardStore?.filters.include.find(
                (d) => d.name === dimensionName
              )?.in) as string[]) ?? [];

        data = prepareDimensionData(
          timeSeries?.timeSeriesData,
          timeSeries?.dimensionTableData,
          columnCount,
          total,
          unfilteredTotal,
          measure,
          selectedValues
        );
      } else {
        comparing = timeControls.showComparison ? "time" : "none";
        const currentRange = timeControls?.selectedTimeRange?.name;

        let currentLabel = "Custom Range";
        if (currentRange in DEFAULT_TIME_RANGES)
          currentLabel = DEFAULT_TIME_RANGES[currentRange].label;

        const comparisonRange = timeControls?.selectedComparisonTimeRange?.name;
        let comparisonLabel = "Custom Range";

        if (comparisonRange in TIME_COMPARISON)
          comparisonLabel = TIME_COMPARISON[comparisonRange].label;

        console.log(timeControls?.selectedTimeRange);
        data = prepareTimeData(
          timeSeries?.timeSeriesData,
          columnCount,
          total,
          comparisonTotal,
          currentLabel,
          comparisonLabel,
          measure,
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
