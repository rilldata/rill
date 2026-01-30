import { getAdjustedFetchTime } from "../ranges";
import { V1TimeGrain } from "../../../runtime-client/gen/index.schemas";
import { describe, it, expect } from "vitest";

const getAdjustedFetchTimeTestCases = [
  {
    test: "should return adjusted dates for a complete period",
    start: new Date("2020-01-04T00:00:00.000Z"),
    end: new Date("2020-01-06T00:00:00.000Z"),
    zone: "UTC",
    interval: V1TimeGrain.TIME_GRAIN_DAY,
    expected: {
      start: "2020-01-03T00:00:00.000Z",
      end: "2020-01-07T00:00:00.000Z",
    },
  },
  {
    test: "should return adjusted dates for an incomplete period",
    start: new Date("2020-01-10T00:00:00.000Z"),
    end: new Date("2020-02-08T00:00:00.000Z"),
    zone: "UTC",
    interval: V1TimeGrain.TIME_GRAIN_WEEK,
    expected: {
      start: "2019-12-30T00:00:00.000Z",
      end: "2020-02-10T00:00:00.000Z",
    },
  },
  {
    test: "should return same dates for invalid grain",
    start: new Date("2020-01-12T00:00:00.000Z"),
    end: new Date("2025-02-08T12:00:33.000Z"),
    zone: "UTC",
    interval: "invalid_grain" as V1TimeGrain,
    expected: {
      start: "2020-01-12T00:00:00.000Z",
      end: "2025-02-08T12:00:33.000Z",
    },
  },
  {
    test: "should return same dates for undefined grain",
    start: new Date("2020-01-12T00:00:00.000Z"),
    end: new Date("2025-02-08T12:00:33.000Z"),
    zone: "UTC",
    interval: undefined,
    expected: {
      start: "2020-01-12T00:00:00.000Z",
      end: "2025-02-08T12:00:33.000Z",
    },
  },
];

describe("getAdjustedFetchTime", () => {
  getAdjustedFetchTimeTestCases.forEach((testCase) => {
    it(testCase.test, () => {
      const defaultTimeGrain = getAdjustedFetchTime(
        testCase.start,
        testCase.end,
        testCase.zone,
        testCase.interval,
      );
      expect(defaultTimeGrain).toEqual(testCase.expected);
    });
  });
});
