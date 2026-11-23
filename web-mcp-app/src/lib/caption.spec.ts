import { describe, expect, it } from "vitest";
import { buildCaptionParts, formatTimeRange, grainLabel } from "./caption";

const week = { start: "2023-09-09T00:00:00Z", end: "2023-09-16T00:00:00Z" };

describe("formatTimeRange", () => {
  it("shows a day-aligned range inclusively", () => {
    expect(formatTimeRange(week, "UTC", "en")).toBe("Sep 9 – Sep 15, 2023");
  });

  it("shows both years when they differ", () => {
    expect(
      formatTimeRange(
        { start: "2023-12-25T00:00:00Z", end: "2024-01-08T00:00:00Z" },
        "UTC",
        "en",
      ),
    ).toBe("Dec 25, 2023 – Jan 7, 2024");
  });

  it("collapses a single day", () => {
    expect(
      formatTimeRange(
        { start: "2023-09-09T00:00:00Z", end: "2023-09-10T00:00:00Z" },
        "UTC",
        "en",
      ),
    ).toBe("Sep 9, 2023");
  });

  it("keeps times for sub-day grains and unaligned ranges", () => {
    const label = formatTimeRange(
      { start: "2023-09-09T06:00:00Z", end: "2023-09-09T18:00:00Z" },
      "UTC",
      "en",
      "TIME_GRAIN_HOUR",
    );
    expect(label).toMatch(/Sep 9.*6:00.*–.*Sep 9, 2023.*6:00/);
  });

  it("respects the time zone", () => {
    // Midnight UTC is 8 PM the day before in New York, so the range is no longer
    // day-aligned there and keeps its times.
    expect(formatTimeRange(week, "America/New_York", "en")).toMatch(
      /^Sep 8.*8:00.*– Sep 15, 2023.*8:00/,
    );
    expect(
      formatTimeRange(
        { start: "2023-09-09T04:00:00Z", end: "2023-09-16T04:00:00Z" },
        "America/New_York",
        "en",
      ),
    ).toBe("Sep 9 – Sep 15, 2023");
  });

  it("localizes month names", () => {
    expect(formatTimeRange(week, "UTC", "es")).toMatch(/sept/i);
  });
});

describe("buildCaptionParts", () => {
  it("orders range, grain, comparison and metrics view", () => {
    expect(
      buildCaptionParts({
        range: week,
        grain: "TIME_GRAIN_DAY",
        showGrain: true,
        hasComparison: true,
        metricsViewLabel: "Bids",
        timeZone: "UTC",
        locale: "en",
      }),
    ).toEqual([
      "Sep 9 – Sep 15, 2023",
      "by day",
      "vs. previous period",
      "Bids",
    ]);
  });

  it("omits the grain for charts without a time axis", () => {
    expect(
      buildCaptionParts({
        range: week,
        grain: "TIME_GRAIN_DAY",
        showGrain: false,
        hasComparison: false,
        timeZone: "UTC",
        locale: "en",
      }),
    ).toEqual(["Sep 9 – Sep 15, 2023"]);
  });

  it("labels grains", () => {
    expect(grainLabel("TIME_GRAIN_MONTH")).toBe("by month");
  });
});
