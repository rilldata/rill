import { derived, writable, type Readable } from "svelte/store";
import {
  StateManagers,
  memoizeMetricsStore,
} from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { createSparkline } from "@rilldata/web-common/components/data-graphic/marks/sparkline";
import { transposeArray } from "./util";
import {
  FormatPreset,
  humanizeDataType,
} from "@rilldata/web-common/features/dashboards/humanize-numbers";
import {
  DEFAULT_TIME_RANGES,
  TIME_COMPARISON,
} from "@rilldata/web-common/lib/time/config";
import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { DimensionDataItem } from "@rilldata/web-common/features/dashboards/time-series/multiple-dimension-queries";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import type { HighlightedCell, TableData } from "./types";

export type TimeDimensionDataState = {
  isFetching: boolean;
  comparing: "dimension" | "time" | "none";
  data?: TableData;
};

export type TimeSeriesDataStore = Readable<TimeDimensionDataState>;

/***
 * Add totals row from time series data
 * Add rest of dimension values from dimension table data
 * Transpose the data to columnar format
 */
function prepareDimensionData(
  totalsData,
  data: DimensionDataItem[],
  // columnCount: number,
  total: number,
  unfilteredTotal: number,
  measure: MetricsViewSpecMeasureV2,
  selectedValues: string[],
  isAllTime: boolean
): TableData {
  if (!data) return;

  const formatPreset =
    (measure?.formatPreset as FormatPreset) ?? FormatPreset.HUMANIZE;
  const measureName = measure?.name;
  const validPercentOfTotal = measure?.validPercentOfTotal;

  const columnHeaderData = (
    isAllTime ? totalsData?.slice(1) : totalsData?.slice(1, -1)
  )?.map((v) => [{ value: v.ts }]);

  const columnCount = columnHeaderData?.length;

  // Add totals row to count
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

  let body = [
    (isAllTime ? totalsData?.slice(1) : totalsData?.slice(1, -1))?.map((v) =>
      v[measureName] ? humanizeDataType(v[measureName], formatPreset) : null
    ) || [],
  ];

  body = body?.concat(
    data?.map((v) => {
      if (v.isFetching) return new Array(columnCount).fill(undefined);
      return (isAllTime ? v?.data?.slice(1) : v?.data?.slice(1, -1))?.map((v) =>
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
  total: number,
  comparisonTotal: number,
  currentLabel: string,
  comparisonLabel: string,
  measure: MetricsViewSpecMeasureV2,
  hasTimeComparison,
  isAllTime: boolean
): TableData {
  if (!data) return;

  const formatPreset =
    (measure?.formatPreset as FormatPreset) ?? FormatPreset.HUMANIZE;
  const measureName = measure?.name;

  const columnHeaderData = (
    isAllTime ? data?.slice(1) : data?.slice(1, -1)
  )?.map((v) => [{ value: v.ts }]);

  const columnCount = columnHeaderData?.length;

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
    (isAllTime ? data?.slice(1) : data?.slice(1, -1))?.map((v) =>
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
      (isAllTime ? data?.slice(1) : data?.slice(1, -1))?.map((v) =>
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
      if (
        !timeControls.ready ||
        timeControls?.isFetching ||
        timeSeries?.isFetching
      )
        return;

      const measureName = dashboardStore?.expandedMeasureName;
      const dimensionName = dashboardStore?.selectedComparisonDimension;
      const total = timeSeries?.total && timeSeries?.total[measureName];
      const unfilteredTotal =
        timeSeries?.unfilteredTotal && timeSeries?.unfilteredTotal[measureName];
      const comparisonTotal =
        timeSeries?.comparisonTotal && timeSeries?.comparisonTotal[measureName];
      const isAllTime =
        timeControls?.selectedTimeRange?.name === TimeRangePreset.ALL_TIME;

      const measure = metricsView?.data?.measures?.find(
        (m) => m.name === measureName
      );

      let comparing;
      let data: TableData;
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
          total,
          unfilteredTotal,
          measure,
          selectedValues,
          isAllTime
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

        data = prepareTimeData(
          timeSeries?.timeSeriesData,
          total,
          comparisonTotal,
          currentLabel,
          comparisonLabel,
          measure,
          comparing === "time",
          isAllTime
        );
      }

      return { isFetching: false, comparing, data };
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

/**
 * Stores for handling interactions between chart and table
 * Two separate stores created to avoid looped updates and renders
 */
export const tableInteractionStore = writable<HighlightedCell>({
  dimensionValue: undefined,
  time: undefined,
});

export const chartInteractionColumn = writable<number>(undefined);
