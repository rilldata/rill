import { describe, it, expect } from "vitest";
import { DateTime, Interval } from "luxon";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  aggregationGrainOrder,
  allowedAggregationGrains,
  allowedGrainsForInterval,
  DateTimeUnitToV1TimeGrain,
  floorToGrain,
  getAvailableTimeGrains,
  grainBucketCount,
  isGrainAllowed,
  plusGrain,
  TokenToV1TimeGrain,
  translateV1TimeGrain,
  V1TimeGrainToToken,
} from "./new-grains";

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

describe("catalog grains", () => {
  it("orders multiples between their base unit and the next standard grain", () => {
    const order = (grain: V1TimeGrain) => aggregationGrainOrder.indexOf(grain);
    expect(order(V1TimeGrain.TIME_GRAIN_MINUTE)).toBeLessThan(
      order(V1TimeGrain.TIME_GRAIN_10_MINUTE),
    );
    expect(order(V1TimeGrain.TIME_GRAIN_30_MINUTE)).toBeLessThan(
      order(V1TimeGrain.TIME_GRAIN_HOUR),
    );
    expect(order(V1TimeGrain.TIME_GRAIN_12_HOUR)).toBeLessThan(
      order(V1TimeGrain.TIME_GRAIN_DAY),
    );
    expect(isGrainAllowed(V1TimeGrain.TIME_GRAIN_10_MINUTE, "minute")).toBe(
      true,
    );
    expect(
      isGrainAllowed(
        V1TimeGrain.TIME_GRAIN_10_MINUTE,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ),
    ).toBe(false);
    expect(allowedAggregationGrains).not.toContain(
      V1TimeGrain.TIME_GRAIN_10_MINUTE,
    );
  });

  it("maps URL and resolver tokens both ways", () => {
    expect(V1TimeGrainToToken[V1TimeGrain.TIME_GRAIN_10_MINUTE]).toBe("10m");
    expect(V1TimeGrainToToken[V1TimeGrain.TIME_GRAIN_6_HOUR]).toBe("6h");
    expect(V1TimeGrainToToken[V1TimeGrain.TIME_GRAIN_DAY]).toBe("day");
    expect(TokenToV1TimeGrain["10m"]).toBe(V1TimeGrain.TIME_GRAIN_10_MINUTE);
    expect(TokenToV1TimeGrain["minute"]).toBe(V1TimeGrain.TIME_GRAIN_MINUTE);
    expect(TokenToV1TimeGrain["day"]).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    expect(DateTimeUnitToV1TimeGrain["minute"]).toBe(
      V1TimeGrain.TIME_GRAIN_MINUTE,
    );
  });

  it("floors and steps on the wall clock", () => {
    const utc = DateTime.fromISO("2024-03-10T04:27:33Z", { zone: "UTC" });
    expect(floorToGrain(utc, V1TimeGrain.TIME_GRAIN_10_MINUTE).toISO()).toBe(
      "2024-03-10T04:20:00.000Z",
    );
    expect(floorToGrain(utc, V1TimeGrain.TIME_GRAIN_6_HOUR).toISO()).toBe(
      "2024-03-10T00:00:00.000Z",
    );
    // 10:30 in Asia/Kolkata floors to 06:00 local (00:30Z), not to the UTC epoch bucket.
    const kolkata = DateTime.fromISO("2024-03-10T05:00:00Z").setZone(
      "Asia/Kolkata",
    );
    expect(
      floorToGrain(kolkata, V1TimeGrain.TIME_GRAIN_6_HOUR).toUTC().toISO(),
    ).toBe("2024-03-10T00:30:00.000Z");
    expect(plusGrain(utc, V1TimeGrain.TIME_GRAIN_10_MINUTE, 3).toISO()).toBe(
      "2024-03-10T04:57:33.000Z",
    );
    expect(
      grainBucketCount(
        createInterval("2024-03-10T04:00:00", "2024-03-10T05:00:00"),
        V1TimeGrain.TIME_GRAIN_10_MINUTE,
      ),
    ).toBe(6);
  });

  it("offers the declared grains, else the standard grains above the smallest one", () => {
    const declared = [
      V1TimeGrain.TIME_GRAIN_10_MINUTE,
      V1TimeGrain.TIME_GRAIN_HOUR,
      V1TimeGrain.TIME_GRAIN_DAY,
    ];
    expect(getAvailableTimeGrains({ timeGrains: declared })).toEqual(declared);
    expect(
      getAvailableTimeGrains({
        smallestTimeGrain: V1TimeGrain.TIME_GRAIN_HOUR,
      }),
    ).toEqual([
      V1TimeGrain.TIME_GRAIN_HOUR,
      V1TimeGrain.TIME_GRAIN_DAY,
      V1TimeGrain.TIME_GRAIN_WEEK,
      V1TimeGrain.TIME_GRAIN_MONTH,
      V1TimeGrain.TIME_GRAIN_QUARTER,
      V1TimeGrain.TIME_GRAIN_YEAR,
    ]);
    expect(getAvailableTimeGrains(undefined)).toEqual(allowedAggregationGrains);

    // A day offers the declared grains that fit it; a week drops the 10-minute grain (1008 buckets still fit) only past 1500 buckets.
    const day = createInterval("2024-03-10T00:00:00", "2024-03-11T00:00:00");
    expect(
      allowedGrainsForInterval(day, V1TimeGrain.TIME_GRAIN_MINUTE, declared),
    ).toEqual(declared);
    const month = createInterval("2024-03-01T00:00:00", "2024-04-01T00:00:00");
    expect(
      allowedGrainsForInterval(month, V1TimeGrain.TIME_GRAIN_MINUTE, declared),
    ).toEqual([V1TimeGrain.TIME_GRAIN_HOUR, V1TimeGrain.TIME_GRAIN_DAY]);
    // Nothing fits a few seconds: fall back to the finest declared grain rather than the smallest grain.
    const seconds = createInterval(
      "2024-03-10T00:00:00",
      "2024-03-10T00:00:30",
    );
    expect(
      allowedGrainsForInterval(
        seconds,
        V1TimeGrain.TIME_GRAIN_MINUTE,
        declared,
      ),
    ).toEqual([V1TimeGrain.TIME_GRAIN_10_MINUTE]);
  });

  it("labels multiples with their count", () => {
    expect(translateV1TimeGrain(V1TimeGrain.TIME_GRAIN_10_MINUTE)).toBe(
      "10 minutes",
    );
    expect(translateV1TimeGrain(V1TimeGrain.TIME_GRAIN_6_HOUR)).toBe("6 hours");
    expect(translateV1TimeGrain(V1TimeGrain.TIME_GRAIN_DAY)).toBe("day");
  });
});
