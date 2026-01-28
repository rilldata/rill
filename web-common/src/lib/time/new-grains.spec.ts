import { describe, it, expect } from "vitest";
import { DateTime, Interval } from "luxon";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { allowedGrainsForInterval } from "./new-grains";

function createInterval(start: string, end: string): Interval<true> {
  const interval = Interval.fromDateTimes(
    DateTime.fromISO(start),
    DateTime.fromISO(end),
  );
  if (!interval.isValid) {
    throw new Error(`Invalid interval: ${start} - ${end}`);
  }
  return interval as Interval<true>;
}

describe("allowedGrainsForInterval", () => {
  describe("edge cases", () => {
    it("returns empty array when interval is undefined", () => {
      expect(allowedGrainsForInterval(undefined)).toEqual([]);
    });

    it("returns minTimeGrain when interval is too small for any grain", () => {
      // 30 seconds - less than 1 minute bucket
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-01T00:00:30",
      );
      expect(allowedGrainsForInterval(interval)).toEqual([
        V1TimeGrain.TIME_GRAIN_MINUTE,
      ]);
    });
  });

  describe("basic intervals", () => {
    it("returns minute and hour for a 1 hour interval", () => {
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-01T01:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_MINUTE);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_HOUR);
    });

    it("returns minute, hour, and day for a 1 day interval", () => {
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-02T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_MINUTE);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("returns appropriate grains for a 1 week interval", () => {
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-08T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_WEEK);
    });

    it("returns appropriate grains for a 1 month interval", () => {
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-02-01T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_WEEK);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_MONTH);
    });

    it("returns appropriate grains for a 1 year interval", () => {
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2025-01-01T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_WEEK);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_MONTH);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_QUARTER);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_YEAR);
    });
  });

  describe("MAX_BUCKETS constraint (1500 buckets)", () => {
    it("excludes minute grain when interval exceeds 1500 minutes (25 hours)", () => {
      // 26 hours = 1560 minutes, should exceed MAX_BUCKETS for minute
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-02T02:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).not.toContain(V1TimeGrain.TIME_GRAIN_MINUTE);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_HOUR);
    });

    it("excludes hour grain when interval exceeds 1500 hours (~62 days)", () => {
      // 70 days = 1680 hours, should exceed MAX_BUCKETS for hour
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-03-11T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).not.toContain(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("always allows year grain even for very large intervals", () => {
      // 10 years
      const interval = createInterval(
        "2014-01-01T00:00:00",
        "2024-01-01T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_YEAR);
    });
  });

  describe("minTimeGrain parameter", () => {
    it("filters out grains smaller than minTimeGrain", () => {
      // 1 day interval
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-02T00:00:00",
      );
      const grains = allowedGrainsForInterval(
        interval,
        V1TimeGrain.TIME_GRAIN_HOUR,
      );
      expect(grains).not.toContain(V1TimeGrain.TIME_GRAIN_MINUTE);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
    });

    it("respects minTimeGrain of DAY", () => {
      // 1 week interval
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-08T00:00:00",
      );
      const grains = allowedGrainsForInterval(
        interval,
        V1TimeGrain.TIME_GRAIN_DAY,
      );
      expect(grains).not.toContain(V1TimeGrain.TIME_GRAIN_MINUTE);
      expect(grains).not.toContain(V1TimeGrain.TIME_GRAIN_HOUR);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_DAY);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_WEEK);
    });

    it("returns minTimeGrain when no grains are valid for the interval", () => {
      // Very small interval with large minTimeGrain
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-01T01:00:00",
      );
      const grains = allowedGrainsForInterval(
        interval,
        V1TimeGrain.TIME_GRAIN_WEEK,
      );
      // No grains are valid (interval < 1 week), so should return minTimeGrain
      expect(grains).toEqual([V1TimeGrain.TIME_GRAIN_WEEK]);
    });

    it("defaults minTimeGrain to MINUTE when not specified", () => {
      // 2 hour interval
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-01-01T02:00:00",
      );
      const grains = allowedGrainsForInterval(interval);
      expect(grains).toContain(V1TimeGrain.TIME_GRAIN_MINUTE);
    });
  });

  describe("grain ordering", () => {
    it("returns grains in ascending order (smallest to largest)", () => {
      // 1 month interval
      const interval = createInterval(
        "2024-01-01T00:00:00",
        "2024-02-01T00:00:00",
      );
      const grains = allowedGrainsForInterval(interval);

      const expectedOrder = [
        V1TimeGrain.TIME_GRAIN_MINUTE,
        V1TimeGrain.TIME_GRAIN_HOUR,
        V1TimeGrain.TIME_GRAIN_DAY,
        V1TimeGrain.TIME_GRAIN_WEEK,
        V1TimeGrain.TIME_GRAIN_MONTH,
      ];

      // Filter grains to only those in result
      const orderedResult = expectedOrder.filter((g) => grains.includes(g));
      expect(grains).toEqual(orderedResult);
    });
  });
});
