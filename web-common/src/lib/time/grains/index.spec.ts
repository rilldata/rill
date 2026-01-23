import { V1TimeGrain } from "../../../runtime-client";
import { TIME_GRAIN } from "../config";
import {
  durationToMillis,
  findValidTimeGrain,
  getAllowedTimeGrains,
  getDefaultTimeGrain,
  getValidatedTimeGrain,
} from "../grains";
import { Interval, DateTime } from "luxon";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { describe, it, expect } from "vitest";
import { Period, type TimeGrain } from "../types";

const allowedGrainTests = [
  {
    test: "should return TIME_GRAIN_MINUTE for < 1 hour",
    start: new Date(0),
    end: new Date(durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) - 1),
    expected: [TIME_GRAIN.TIME_GRAIN_MINUTE],
  },
  {
    test: "should return TIME_GRAIN_MINUTE and TIME_GRAIN_HOUR if otherwise < 6 hours",
    start: new Date(0),
    end: new Date(
      6 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) - 1,
    ),
    expected: [TIME_GRAIN.TIME_GRAIN_MINUTE, TIME_GRAIN.TIME_GRAIN_HOUR],
  },
  {
    test: "should return TIME_GRAIN_MINUTE and TIME_GRAIN_HOUR if otherwise < 1 day",
    start: new Date(0),
    end: new Date(durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) - 1),
    expected: [TIME_GRAIN.TIME_GRAIN_MINUTE, TIME_GRAIN.TIME_GRAIN_HOUR],
  },

  {
    test: "should return TIME_GRAIN_HOUR and TIME_GRAIN_DAY if otherwise < 7 days",
    start: new Date(0),
    end: new Date(durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 7 - 1),
    expected: [TIME_GRAIN.TIME_GRAIN_HOUR, TIME_GRAIN.TIME_GRAIN_DAY],
  },
  {
    test: "should return TIME_GRAIN_HOUR, TIME_GRAIN_DAY, and TIME_GRAIN_WEEK if otherwise < 30 days",
    start: new Date(0),
    end: new Date(
      durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30 - 1,
    ),
    expected: [
      TIME_GRAIN.TIME_GRAIN_HOUR,
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
    ],
  },
  {
    test: "should return TIME_GRAIN_WEEK, TIME_GRAIN_MONTH, and TIME_GRAIN_YEAR if otherwise < 90 days",
    start: new Date(0),
    end: new Date(
      durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 3 * 30 - 1,
    ),
    expected: [
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
    ],
  },
  {
    test: "should return TIME_GRAIN_WEEK, TIME_GRAIN_MONTH, TIME_GRAIN_QUARTER and TIME_GRAIN_YEAR if otherwise < 1 year",
    start: new Date(0),
    end: new Date(durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) - 1),
    expected: [
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
      TIME_GRAIN.TIME_GRAIN_QUARTER,
    ],
  },
  {
    test: "should return TIME_GRAIN_WEEK, TIME_GRAIN_MONTH, TIME_GRAIN_QUARTER and TIME_GRAIN_YEAR if otherwise < 10 years",
    start: new Date(0),
    end: new Date(
      10 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) - 1,
    ),
    expected: [
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
      TIME_GRAIN.TIME_GRAIN_QUARTER,
      TIME_GRAIN.TIME_GRAIN_YEAR,
    ],
  },
];

const defaultTimeGrainTests = [
  {
    test: "should return TIME_GRAIN_MINUTE for < 2 hours",
    start: new Date(0),
    end: new Date(
      2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) - 1,
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
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30 - 1,
    ),
    expected: TIME_GRAIN.TIME_GRAIN_DAY,
  },
  {
    test: "should return TIME_GRAIN_WEEK if otherwise < 3 years",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) - 1,
    ),
    expected: TIME_GRAIN.TIME_GRAIN_WEEK,
  },
  {
    test: "should return TIME_GRAIN_MONTH if otherwise >= 3 years",
    start: new Date(0),
    end: new Date(
      3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration) + 1,
    ),
    expected: TIME_GRAIN.TIME_GRAIN_MONTH,
  },
];

const timeGrainOptions: TimeGrain[] = [
  {
    grain: V1TimeGrain.TIME_GRAIN_DAY,
    label: "day",
    duration: Period.DAY,
    d3format: "",
    formatDate: {},
  },
  {
    grain: V1TimeGrain.TIME_GRAIN_WEEK,
    label: "week",
    duration: Period.WEEK,
    d3format: "",
    formatDate: {},
  },
  {
    grain: V1TimeGrain.TIME_GRAIN_MONTH,
    label: "month",
    duration: Period.MONTH,
    d3format: "",
    formatDate: {},
  },
];

const findValidTimeGrainTests = [
  {
    test: "findValidTimeGrain returns a valid time grain",
    timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
    minTimeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
    expected: V1TimeGrain.TIME_GRAIN_WEEK,
  },
  {
    test: "findValidTimeGrain returns a valid time grain when there is no minTimeGrain",
    timeGrain: V1TimeGrain.TIME_GRAIN_HOUR,
    minTimeGrain: undefined,
    expected: V1TimeGrain.TIME_GRAIN_DAY,
  },
  {
    test: "findValidTimeGrain returns the default time grain as fallback",
    timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
    minTimeGrain: V1TimeGrain.TIME_GRAIN_HOUR,
    expected: V1TimeGrain.TIME_GRAIN_WEEK,
  },
  {
    test: "findValidTimeGrain finds and returns a valid time grain",
    timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
    minTimeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
    expected: V1TimeGrain.TIME_GRAIN_WEEK,
  },
];

describe("getAllowedTimeGrains", () => {
  allowedGrainTests.forEach((testCase) => {
    it(testCase.test, () => {
      const allowedTimeGrains = getAllowedTimeGrains(
        testCase.start,
        testCase.end,
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
        testCase.end,
      );
      expect(defaultTimeGrain).toEqual(testCase.expected);
    });
  });
});

describe("findValidTimeGrain", () => {
  findValidTimeGrainTests.forEach((testCase) => {
    it(testCase.test, () => {
      const defaultTimeGrain = findValidTimeGrain(
        testCase.timeGrain,
        timeGrainOptions,

        testCase.minTimeGrain,
      );
      expect(defaultTimeGrain).toEqual(testCase.expected);
    });
  });
});

describe("getValidatedTimeGrain", () => {
  // Helper to create a valid interval
  function createInterval(days: number): Interval<true> {
    const start = DateTime.fromISO("2024-01-01T00:00:00Z");
    const end = start.plus({ days });
    return Interval.fromDateTimes(start, end) as Interval<true>;
  }

  function createIntervalHours(hours: number): Interval<true> {
    const start = DateTime.fromISO("2024-01-01T00:00:00Z");
    const end = start.plus({ hours });
    return Interval.fromDateTimes(start, end) as Interval<true>;
  }

  describe("returns undefined for invalid inputs", () => {
    it("returns undefined when interval is undefined", () => {
      const result = getValidatedTimeGrain(
        undefined,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_DAY,
        undefined,
      );
      expect(result).toBeUndefined();
    });

    it("returns undefined when interval is invalid", () => {
      const invalidInterval = Interval.invalid("test invalid interval");
      const result = getValidatedTimeGrain(
        invalidInterval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_DAY,
        undefined,
      );
      expect(result).toBeUndefined();
    });
  });

  describe("uses requestedPrecision when allowed", () => {
    it("returns requestedPrecision when it is in allowed grains", () => {
      const interval = createInterval(30); // 30 days: allows hour, day, week
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_DAY,
        undefined,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("returns requestedPrecision for hour grain in short interval", () => {
      const interval = createIntervalHours(6); // 6 hours: allows minute, hour
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_HOUR,
        undefined,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
    });

    it("ignores requestedPrecision when not in allowed grains", () => {
      const interval = createInterval(30); // 30 days: allows hour, day, week
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE, // minute not allowed for 30 days
        undefined,
      );
      // Should fall back to first allowed grain (hour)
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
    });
  });

  describe("uses rangePrecision from parsed RillTime as fallback", () => {
    it("uses rangePrecision when requestedPrecision is not provided", () => {
      const interval = createInterval(7); // 7 days: allows hour, day
      const parsed = parseRillTime("7d as of latest/d"); // snap to day
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        parsed,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("uses rangePrecision when requestedPrecision is not allowed", () => {
      const interval = createInterval(30); // 30 days: allows hour, day, week
      const parsed = parseRillTime("30d as of latest/d"); // snap to day
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE, // not allowed for 30 days
        parsed,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("ignores rangePrecision when not in allowed grains", () => {
      const interval = createInterval(365); // ~1 year: allows day, week, month, quarter
      const parsed = parseRillTime("365d as of latest/h"); // snap to hour, not allowed for 365 days
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        parsed,
      );
      // Should fall back to first allowed grain (day for ~1 year)
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });
  });

  describe("falls back to first allowed grain", () => {
    it("uses first allowed grain when no precision is specified", () => {
      const interval = createInterval(30); // 30 days: allows hour, day, week
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        undefined,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
    });

    it("uses first allowed grain when both precisions are invalid", () => {
      const interval = createInterval(365); // ~1 year
      const parsed = parseRillTime("365d as of latest/m"); // minute precision, not allowed
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_MINUTE, // not allowed for 365 days
        parsed,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });
  });

  describe("respects minTimeGrain constraint", () => {
    it("filters out grains smaller than minTimeGrain", () => {
      const interval = createInterval(7); // 7 days: normally allows hour, day
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_DAY, // min is day, so hour should be excluded
        V1TimeGrain.TIME_GRAIN_HOUR, // requesting hour, but it's below min
        undefined,
      );
      // Hour is below minTimeGrain, so should get day
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("returns minTimeGrain when no grains are available", () => {
      // Very short interval where only minute is valid, but minTimeGrain is year
      const interval = createIntervalHours(1); // 1 hour
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_YEAR,
        undefined,
        undefined,
      );
      // Falls back to minTimeGrain since no grains are appropriate
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_YEAR);
    });
  });

  describe("integration with real Rill time strings", () => {
    it("derives day grain for 365d as of latest/h", () => {
      const interval = createInterval(365);
      const parsed = parseRillTime("365d as of latest/h");
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        parsed,
      );
      // 365 days at hour grain = 8760 buckets (exceeds 1500), so use day
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("derives hour grain for 24h as of latest/h", () => {
      const interval = createIntervalHours(24);
      const parsed = parseRillTime("24h as of latest/h");
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        parsed,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_HOUR);
    });

    it("derives week grain for 52w as of latest/w", () => {
      const interval = createInterval(52 * 7); // 52 weeks
      const parsed = parseRillTime("52w as of latest/w");
      const result = getValidatedTimeGrain(
        interval,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        parsed,
      );
      expect(result).toBe(V1TimeGrain.TIME_GRAIN_WEEK);
    });

    it("derives minute grain for 24h as of latest/m (1440 buckets)", () => {
      const interval = createIntervalHours(24);
      const parsed = parseRillTime("24h as of latest/m");
      const result = getValidatedTimeGrain(
        interval,

        V1TimeGrain.TIME_GRAIN_MINUTE,
        undefined,
        parsed,
      );

      expect(result).toBe(V1TimeGrain.TIME_GRAIN_MINUTE);
    });
  });
});
