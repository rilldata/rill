import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { MetricsViewTimeSeriesResponse } from "$common/rill-developer-service/MetricsViewActions";
import type { PreviewRollupInterval } from "$lib/duckdb-data-types";
import { END_DATE, START_DATE } from "../data/generator/data-constants";
import { isTimestampDiffAccurate } from "./time-series-time-diff";

export type TimeSeriesMeasureRange = Record<string, [min: number, max: number]>;

export function getTimeRange(
  interval: string,
  startDate = START_DATE,
  endDate = END_DATE
) {
  return {
    interval,
    start: new Date(`${startDate} UTC`).toISOString(),
    end: new Date(`${endDate} UTC`).toISOString(),
  } as TimeSeriesTimeRange;
}

export function assertTimeSeries(
  timeSeries: MetricsViewTimeSeriesResponse,
  rollupInterval: PreviewRollupInterval,
  measures: Array<string>
) {
  const mismatchTimestamps = new Array<[string, string]>();
  const mismatchMeasures = new Array<
    [dimension: string, value: number, timestamp: string]
  >();
  const rollupIntervalGrain = rollupInterval.split(" ")[1];

  let prevRow: any;
  for (const row of timeSeries.data) {
    if (prevRow) {
      if (!isTimestampDiffAccurate(prevRow.ts, row.ts, rollupIntervalGrain)) {
        mismatchTimestamps.push([prevRow.ts, row.ts]);
      }
    }
    prevRow = row;
    for (const measure of measures) {
      if (Number.isNaN(Number(row[measure]))) {
        mismatchMeasures.push([measure, row[measure], row.ts]);
      }
    }
  }

  if (mismatchTimestamps.length) {
    console.log("Mismatch timestamps: ", mismatchTimestamps);
  }
  if (mismatchMeasures.length) {
    console.log("Mismatch measures: ", mismatchMeasures);
  }
  expect(mismatchTimestamps.length).toBe(0);
  expect(mismatchMeasures.length).toBe(0);
}

export function assertTimeSeriesMeasureRange(
  timeSeries: MetricsViewTimeSeriesResponse,
  measureRanges: Array<TimeSeriesMeasureRange>
) {
  expect(timeSeries.data.length).toBe(measureRanges.length);

  const mismatchMeasures = new Array<
    [dimension: string, value: number, timestamp: string]
  >();

  timeSeries.data.forEach((row, index) => {
    for (const measureName in measureRanges[index]) {
      const value = row[measureName];
      if (
        value < measureRanges[index][measureName][0] &&
        value > measureRanges[index][measureName][1]
      ) {
        mismatchMeasures.push([measureName, value, row.ts]);
      }
    }
  });

  if (mismatchMeasures.length) {
    console.log("Mismatch measures value ranges: ", mismatchMeasures);
  }
  expect(mismatchMeasures.length).toBe(0);
}

export function assertBigNumber(
  bigNumber: Record<string, number>,
  expectedBigNumber: TimeSeriesMeasureRange
) {
  const mismatchBigNumbers = new Array<[dimension: string, value: number]>();

  for (const measureName in expectedBigNumber) {
    const value = bigNumber[measureName];
    if (
      value < expectedBigNumber[measureName][0] &&
      value > expectedBigNumber[measureName][1]
    ) {
      mismatchBigNumbers.push([measureName, value]);
    }
  }

  if (mismatchBigNumbers.length) {
    console.log("Mismatch big numbers: ", mismatchBigNumbers);
  }
  expect(mismatchBigNumbers.length).toBe(0);
}
