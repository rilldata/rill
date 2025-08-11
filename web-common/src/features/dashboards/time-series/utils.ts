import type { GraphicScale } from "@rilldata/web-common/components/data-graphic/state/types";
import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";
import { createIndexMap } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  createAndExpression,
  filterExpressions,
  matchExpressionByName,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { chartInteractionColumn } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import type {
  V1Expression,
  V1MetricsViewAggregationResponseDataItem,
  V1TimeSeriesValue,
} from "@rilldata/web-common/runtime-client";
import type { DateTimeUnit } from "luxon";
import { get } from "svelte/store";
import { removeZoneOffset } from "../../../lib/time/timezone";
import { getDurationMultiple, getOffset } from "../../../lib/time/transforms";
import { TimeOffsetType } from "../../../lib/time/types";
import { roundToNearestTimeUnit } from "./round-to-nearest-time-unit";
import type { TimeSeriesDatum } from "./timeseries-data-store";

/** sets extents to 0 if it makes sense; otherwise, inflates each extent component */
export function niceMeasureExtents(
  [smallest, largest]: [number, number],
  inflator: number,
) {
  if (smallest === 0 && largest === 0) {
    return [0, 1];
  }
  return [
    smallest < 0 ? smallest * inflator : 0,
    largest > 0 ? largest * inflator : 0,
  ];
}

export function toComparisonKeys(
  d,
  offsetDuration: string,
  zone: string,
  grainDuration: string,
) {
  return Object.keys(d).reduce((acc, key) => {
    if (key === "records") {
      Object.entries(d.records).forEach(([key, value]) => {
        acc[`comparison.${key}`] = value;
      });
    } else if (`comparison.${key}` === "comparison.ts") {
      acc[`comparison.${key}`] = adjustOffsetForZone(
        d[key],
        zone,
        grainDuration,
      );
      acc["comparison.ts_position"] = getOffset(
        acc["comparison.ts"],
        offsetDuration,
        TimeOffsetType.ADD,
      );
    } else {
      acc[`comparison.${key}`] = d[key];
    }
    return acc;
  }, {});
}

export function updateChartInteractionStore(
  xHoverValue: undefined | number | Date,
  yHoverValue: undefined | string | null,
  isAllTime: boolean,
  formattedData: TimeSeriesDatum[],
) {
  let xHoverColNum: number | undefined = undefined;

  const slicedData = isAllTime
    ? formattedData?.slice(1)
    : formattedData?.slice(1, -1);

  if (xHoverValue && xHoverValue instanceof Date) {
    const { position } = bisectData(
      xHoverValue,
      "center",
      "ts_position",
      slicedData,
    );
    xHoverColNum = position;
  }

  const currentCol = get(chartInteractionColumn);

  if (
    currentCol?.xHover !== xHoverColNum ||
    currentCol?.yHover !== yHoverValue
  ) {
    chartInteractionColumn.update((state) => ({
      ...state,
      yHover: yHoverValue,
      xHover: xHoverColNum,
    }));
  }
}

export function prepareTimeSeries(
  original: V1TimeSeriesValue[],
  comparison: V1TimeSeriesValue[] | undefined,
  timeGrainDuration: string,
  zone: string,
): TimeSeriesDatum[] {
  return original?.map((originalPt, i) => {
    const comparisonPt = comparison?.[i];

    const emptyPt = {
      ts: undefined,
      ts_position: undefined,
      bin: undefined,
      ...originalPt.records,
    };

    if (!originalPt?.ts) {
      return emptyPt;
    }
    const ts = adjustOffsetForZone(originalPt.ts, zone, timeGrainDuration);

    if (!ts || typeof ts === "string") {
      return emptyPt;
    }
    const offsetDuration = getDurationMultiple(timeGrainDuration, 0.5);
    const ts_position = getOffset(ts, offsetDuration, TimeOffsetType.ADD);
    return {
      ts,
      ts_position,
      bin: originalPt.bin,
      ...originalPt.records,
      ...toComparisonKeys(
        comparisonPt || {},
        offsetDuration,
        zone,
        timeGrainDuration,
      ),
    };
  });
}

export function getBisectedTimeFromCordinates(
  value: number,
  scaleStore: GraphicScale,
  accessor: string,
  data: TimeSeriesDatum[],
  grainLabel: DateTimeUnit,
): Date | null {
  const roundedValue = roundToNearestTimeUnit(
    new Date(scaleStore.invert(value)),
    grainLabel,
  );
  const { entry: bisector } = bisectData(
    roundedValue,
    "center",
    accessor,
    data,
  );
  if (!bisector || typeof bisector === "number") return null;
  const bisected = bisector[accessor];
  if (!bisected) return null;

  return new Date(bisected);
}

/**
 *  The dates in the charts are in the local timezone, this util method
 *  removes the selected timezone offset and adds the local offset
 */
export function localToTimeZoneOffset(dt: Date, zone: string) {
  const utcDate = new Date(dt.getTime() - dt.getTimezoneOffset() * 60000);
  return removeZoneOffset(utcDate, zone);
}

// Return start and end of the time range that is ordered.
export function getOrderedStartEnd(start: Date, stop: Date) {
  const startMs = start?.getTime();
  const stopMs = stop?.getTime();

  if (startMs > stopMs) {
    return { start: stop, end: start };
  } else {
    return { start, end: stop };
  }
}

export function getFilterForComparedDimension(
  dimensionName: string,
  filters: V1Expression,
) {
  let updatedFilter = filterExpressions(
    filters,
    (e) => !matchExpressionByName(e, dimensionName),
  );
  if (!updatedFilter) {
    updatedFilter = createAndExpression([]);
  }
  return updatedFilter;
}

/**
 * This function transforms aggregation response data into time series
 * response data. The aggregation API do not null fill missing data.
 * We use the timeseries response data to create headers and null fill
 * missing data. This also converts Aggregation API's cell level data to
 * row level data
 */
export function transformAggregateDimensionData(
  timeDimension: string,
  dimensionName: string,
  measures: string[],
  dimensionValues: (string | null)[],
  timeSeriesData: V1TimeSeriesValue[],
  response: V1MetricsViewAggregationResponseDataItem[],
): V1TimeSeriesValue[][] {
  const emptyData: V1TimeSeriesValue[][] = new Array<V1TimeSeriesValue[]>(
    dimensionValues.length,
  ).fill([]);

  const hasResponse = response && response.length > 0;

  const headers = timeSeriesData.map((d) => d.ts);
  if (!headers.length) return emptyData;

  const emptyMeasuresObj = measures.reduce((acc, measure) => {
    acc[measure] = hasResponse ? null : undefined;
    return acc;
  }, {});

  const emptyRow = headers.map((h) => ({
    ts: h,
    bin: 0,
    records: emptyMeasuresObj,
  }));

  const data: V1TimeSeriesValue[][] = new Array(dimensionValues.length)
    .fill(null)
    // Create a deep copy of each row for each element
    .map(() =>
      emptyRow.map((row) => ({ ...row, records: { ...row.records } })),
    );

  const dimensionValuesMap = createIndexMap(dimensionValues);
  const headersMap = createIndexMap(headers);

  for (const cell of response) {
    const { [dimensionName]: key, [timeDimension]: ts, ...rest } = cell;
    const timeSeriesCell: V1TimeSeriesValue = {
      ts: ts as string | undefined,
      bin: 0,
      records: { ...rest },
    };

    const rowIndex = dimensionValuesMap.get(key as string | null);
    const colIndex = headersMap.get(ts as string | undefined);

    if (rowIndex !== undefined && colIndex !== undefined) {
      data[rowIndex][colIndex] = timeSeriesCell;
    }
  }

  return data;
}

export function adjustTimeInterval(
  interval: { start: Date; end: Date },
  zone: string,
) {
  const { start, end } = getOrderedStartEnd(interval?.start, interval?.end);
  const adjustedStart = start ? localToTimeZoneOffset(start, zone) : start;
  const adjustedEnd = end ? localToTimeZoneOffset(end, zone) : end;
  return { start: adjustedStart, end: adjustedEnd };
}
