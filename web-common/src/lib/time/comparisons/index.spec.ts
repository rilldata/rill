import {
  getAvailableComparisonsForTimeRange,
  getComparisonRange,
  isComparisonInsideBounds,
} from ".";
import { TimeComparisonOption } from "../types";
import { describe, it, expect } from "vitest";

const contiguousAndCustomComparisonRanges = [
  // contiguous cases
  {
    description:
      "should should return a contiguous time range for a single day when comparison is CONTIGUOUS",
    input: {
      start: new Date("2020-03-05T00:00:00.000Z"),
      end: new Date("2020-03-06T00:00:00.000Z"),
      comparison: TimeComparisonOption.CONTIGUOUS,
    },
    output: {
      start: new Date("2020-03-04T00:00:00.000Z"),
      end: new Date("2020-03-05T00:00:00.000Z"),
    },
  },
  {
    description:
      "should should return a contiguous, un-quantized time range for six hours when comparison is CONTIGUOUS",
    input: {
      start: new Date("2020-03-05T06:05:00.000Z"),
      end: new Date("2020-03-06T12:05:00.000Z"),
      comparison: TimeComparisonOption.CONTIGUOUS,
    },
    output: {
      start: new Date("2020-03-04T00:05:00.000Z"),
      end: new Date("2020-03-05T06:05:00.000Z"),
    },
  },
];

const periodStart = new Date("2020-03-05T00:00:00.000Z");
const periodEnd = new Date("2020-03-10T00:00:00.000Z");
const periodicComparisonTests = [
  {
    description: "should return a 1 day period when comparison is P1D",
    input: {
      start: periodStart,
      end: periodEnd,
      comparison: TimeComparisonOption.DAY,
    },
    output: {
      start: new Date("2020-03-04T00:00:00.000Z"),
      end: new Date("2020-03-09T00:00:00.000Z"),
    },
  },
  {
    description: "should return a 1 week period when comparison is P1W",
    input: {
      start: periodStart,
      end: periodEnd,
      comparison: TimeComparisonOption.WEEK,
    },
    output: {
      start: new Date("2020-02-27T00:00:00.000Z"),
      end: new Date("2020-03-03T00:00:00.000Z"),
    },
  },
  {
    description: "should return a 1 month period when comparison is P1M",
    input: {
      start: periodStart,
      end: periodEnd,
      comparison: TimeComparisonOption.MONTH,
    },
    output: {
      start: new Date("2020-02-05T00:00:00.000Z"),
      end: new Date("2020-02-10T00:00:00.000Z"),
    },
  },
  {
    description: "should return a 1 year period when comparison is P1Y",
    input: {
      start: periodStart,
      end: periodEnd,
      comparison: TimeComparisonOption.YEAR,
    },
    output: {
      start: new Date("2019-03-05T00:00:00.000Z"),
      end: new Date("2019-03-10T00:00:00.000Z"),
    },
  },
];

const getComparisonRangeTests = [
  ...contiguousAndCustomComparisonRanges,
  ...periodicComparisonTests,
];

describe("getComparisonRange", () => {
  getComparisonRangeTests.forEach((test) => {
    it(test.description, () => {
      const { start, end, comparison } = test.input;
      const { start: expectedStart, end: expectedEnd } = test.output;
      const { start: actualStart, end: actualEnd } = getComparisonRange(
        start,
        end,
        comparison
      );
      expect(actualStart).toEqual(expectedStart);
      expect(actualEnd).toEqual(expectedEnd);
    });
  });
});

const boundsTestStart = new Date("2020-03-05T00:00:00.000Z");
const boundsTestEnd = new Date("2020-03-12T00:00:00.000Z");
const rangeStart = new Date("2020-03-06T00:00:00.000Z");
const rangeEnd = new Date("2020-03-07T00:00:00.000Z");

const isComparisonInsideBoundsTests = [
  {
    description:
      "should return true when a day offset comparison is inside the bounds",
    input: TimeComparisonOption.DAY,
    output: true,
  },
  {
    description:
      "should return false when a week offset comparison is inside the bounds",
    input: TimeComparisonOption.WEEK,
    output: false,
  },
];
describe("isComparisonInsideBounds", () => {
  isComparisonInsideBoundsTests.forEach((test) => {
    it(test.description, () => {
      const { input, output } = test;
      const actual = isComparisonInsideBounds(
        boundsTestStart,
        boundsTestEnd,
        rangeStart,
        rangeEnd,
        input
      );
      expect(actual).toEqual(output);
    });
  });
});

const getAvailableComparisonsForTimeRangeTests = [
  {
    description:
      "should return all comparison points for a 1 day range over years",
    input: {
      start: new Date("2023-03-04T00:00:00.000Z"),
      end: new Date("2023-03-05T00:00:00.000Z"),
      boundStart: new Date("2020-03-05T00:00:00.000Z"),
      boundEnd: new Date("2023-03-05T00:00:00.000Z"),
    },
    output: [
      TimeComparisonOption.CONTIGUOUS,
      TimeComparisonOption.DAY,
      TimeComparisonOption.WEEK,
      TimeComparisonOption.MONTH,
      TimeComparisonOption.QUARTER,
      TimeComparisonOption.YEAR,
    ],
  },
  {
    description:
      "should return all comparison points for a 1 week range over years",
    input: {
      start: new Date("2023-02-01T00:00:00.000Z"),
      end: new Date("2023-02-07T00:00:00.000Z"),
      boundStart: new Date("2020-03-05T00:00:00.000Z"),
      boundEnd: new Date("2023-03-05T00:00:00.000Z"),
    },
    output: [
      TimeComparisonOption.CONTIGUOUS,
      TimeComparisonOption.WEEK,
      TimeComparisonOption.MONTH,
      TimeComparisonOption.QUARTER,
      TimeComparisonOption.YEAR,
    ],
  },
  {
    description:
      "should return all comparison points for larger than week range over years",
    input: {
      start: new Date("2023-01-20T00:00:00.000Z"),
      end: new Date("2023-02-18T00:00:00.000Z"),
      boundStart: new Date("2020-03-05T00:00:00.000Z"),
      boundEnd: new Date("2023-03-05T00:00:00.000Z"),
    },
    output: [
      TimeComparisonOption.CONTIGUOUS,
      TimeComparisonOption.MONTH,
      TimeComparisonOption.QUARTER,
      TimeComparisonOption.YEAR,
    ],
  },
  {
    description:
      "should return all comparison points for larger than month range over years",
    input: {
      start: new Date("2023-01-20T00:00:00.000Z"),
      end: new Date("2023-03-05T00:00:00.000Z"),
      boundStart: new Date("2020-03-05T00:00:00.000Z"),
      boundEnd: new Date("2023-03-05T00:00:00.000Z"),
    },
    output: [
      TimeComparisonOption.CONTIGUOUS,
      TimeComparisonOption.QUARTER,
      TimeComparisonOption.YEAR,
    ],
  },
  {
    description:
      "should return no options if range is too big to have a comparison window",
    input: {
      start: new Date("2021-01-01T00:00:00.000Z"),
      end: new Date("2022-01-01T00:00:00.000Z"),
      boundStart: new Date("2020-06-01T00:00:00.000Z"),
      boundEnd: new Date("2023-01-01T00:00:00.000Z"),
    },
    output: [],
  },
  {
    description:
      "should return no options if the range is too close to the end to have a comparison window",
    input: {
      start: new Date("2020-01-01T00:00:00.000Z"),
      end: new Date("2020-02-01T00:00:00.000Z"),
      boundStart: new Date("2020-01-01T00:00:00.000Z"),
      boundEnd: new Date("2023-01-01T00:00:00.000Z"),
    },
    output: [],
  },
];

describe("getAvailableComparisonsForTimeRange", () => {
  getAvailableComparisonsForTimeRangeTests.forEach((test) => {
    it(test.description, () => {
      const { start, end, boundStart, boundEnd } = test.input;
      const actual = getAvailableComparisonsForTimeRange(
        boundStart,
        boundEnd,
        start,
        end,
        [...(Object.values(TimeComparisonOption) as TimeComparisonOption[])]
      );
      expect(actual).toEqual(test.output);
    });
  });
});
