import { TIME_GRAIN } from "../config";
import {
  durationToMillis,
  getAllowedTimeGrains,
  getDefaultTimeGrain,
} from "../grains";
import { describe, it, expect } from "vitest";

const allowedGrainTests = [
  {
    test: "should return TIME_GRAIN_MINUTE for < 2 hours",
    start: new Date(0),
    end: new Date(
      2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) - 1
    ),
    expected: [TIME_GRAIN.TIME_GRAIN_MINUTE],
  },
  {
    test: "should return TIME_GRAIN_MINUTE and TIME_GRAIN_HOUR if otherwise < 6 hours",
    start: new Date(0),
    end: new Date(
      6 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) - 1
    ),
    expected: [TIME_GRAIN.TIME_GRAIN_MINUTE, TIME_GRAIN.TIME_GRAIN_HOUR],
  },
  {
    test: "should return TIME_GRAIN_HOUR if otherwise < 1 day",
    start: new Date(0),
    end: new Date(durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) - 1),
    expected: [TIME_GRAIN.TIME_GRAIN_HOUR],
  },
  {
    test: "should return TIME_GRAIN_HOUR for 24 hours",
    start: new Date(0),
    end: new Date(durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)),
    expected: [TIME_GRAIN.TIME_GRAIN_HOUR],
  },
  {
    test: "should return TIME_GRAIN_HOUR and TIME_GRAIN_DAY if otherwise < 14 days",
    start: new Date(0),
    end: new Date(
      durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 14 - 1
    ),
    expected: [TIME_GRAIN.TIME_GRAIN_HOUR, TIME_GRAIN.TIME_GRAIN_DAY],
  },
  {
    test: "should return TIME_GRAIN_HOUR, TIME_GRAIN_DAY, and TIME_GRAIN_WEEK if otherwise < 30 days",
    start: new Date(0),
    end: new Date(
      durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30 - 1
    ),
    expected: [
      TIME_GRAIN.TIME_GRAIN_HOUR,
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
    ],
  },
  {
    test: "should return TIME_GRAIN_DAY, TIME_GRAIN_WEEK, and TIME_GRAIN_MONTH if otherwise < 90 days",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30 - 1
    ),
    expected: [TIME_GRAIN.TIME_GRAIN_DAY, TIME_GRAIN.TIME_GRAIN_WEEK],
  },
  {
    test: "should return TIME_GRAIN_WEEK, TIME_GRAIN_MONTH, and TIME_GRAIN_YEAR if otherwise < 3 years",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) - 1
    ),
    expected: [
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
    ],
  },
  {
    test: "should return TIME_GRAIN_MONTH, TIME_GRAIN_YEAR, and TIME_GRAIN_QUARTER if otherwise < 10 years",
    start: new Date(0),
    end: new Date(
      10 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) - 1
    ),
    expected: [
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
      TIME_GRAIN.TIME_GRAIN_YEAR,
    ],
  },
];

const defaultTimeGrainTests = [
  {
    test: "should return TIME_GRAIN_MINUTE for < 2 hours",
    start: new Date(0),
    end: new Date(
      2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) - 1
    ),
    expected: TIME_GRAIN.TIME_GRAIN_MINUTE,
  },
  {
    test: "should return TIME_GRAIN_HOUR if otherwise < 7 days",
    start: new Date(0),
    end: new Date(7 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) - 1),
    expected: TIME_GRAIN.TIME_GRAIN_HOUR,
  },
  {
    test: "should return TIME_GRAIN_DAY if otherwise < 7 days",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30 - 1
    ),
    expected: TIME_GRAIN.TIME_GRAIN_DAY,
  },
  {
    test: "should return TIME_GRAIN_WEEK if otherwise < 3 years",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) - 1
    ),
    expected: TIME_GRAIN.TIME_GRAIN_WEEK,
  },
  {
    test: "should return TIME_GRAIN_MONTH if otherwise >= 3 years",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) + 1
    ),
    expected: TIME_GRAIN.TIME_GRAIN_MONTH,
  },
];

describe("getAllowedTimeGrains", () => {
  allowedGrainTests.forEach((testCase) => {
    it(testCase.test, () => {
      const allowedTimeGrains = getAllowedTimeGrains(
        testCase.start,
        testCase.end
      );
      expect(allowedTimeGrains).toEqual(testCase.expected);
    });
  });
});

describe("getDefaultTimeGrain", () => {
  defaultTimeGrainTests.forEach((testCase) => {
    it(testCase.test, () => {
      const defaultTimeGrain = getDefaultTimeGrain(
        testCase.start,
        testCase.end
      );
      expect(defaultTimeGrain).toEqual(testCase.expected);
    });
  });
});
