import {
  calculateComparisonTimeRangePartial,
  calculateTimeRangePartial,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { MIN_TIMESTAMP_MS } from "@rilldata/web-common/lib/time/timestamp-bounds";
import {
  TimeComparisonOption,
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

// Data with placeholder dates such as 0001-01-01 has bounds in year 1.
// A legacy preset resolved locally against those bounds, a custom range from the URL,
// or a scrub reaching before them would otherwise land in year 0, which the request encoder rejects.
// Real placeholder data sits exactly at 0001-01-01T00:00:00Z.
const allTimeRange: DashboardTimeControls = {
  name: TimeRangePreset.ALL_TIME,
  start: new Date("0001-01-01T00:00:00.000Z"),
  end: new Date("0001-01-03T00:00:00.000Z"),
};

describe("calculateTimeRangePartial", () => {
  it("keeps a locally resolved preset within the timestamp bounds", () => {
    const state = calculateTimeRangePartial(
      allTimeRange,
      { name: "P30D" } as DashboardTimeControls,
      undefined,
      "UTC",
      allTimeRange,
      V1TimeGrain.TIME_GRAIN_HOUR,
      true,
    );

    expect(state).toBeDefined();
    expect(state?.selectedTimeRange?.start.getTime()).toBe(MIN_TIMESTAMP_MS);
    expect(state?.timeStart).toBe("0001-01-01T00:00:00.000Z");
    expect(state?.interval?.isValid).toBe(true);
    for (const value of [
      state?.timeStart,
      state?.timeEnd,
      state?.adjustedStart,
      state?.adjustedEnd,
    ]) {
      expect(Date.parse(value ?? "")).toBeGreaterThanOrEqual(MIN_TIMESTAMP_MS);
    }
  });

  it("keeps a custom range from the URL within the timestamp bounds", () => {
    const state = calculateTimeRangePartial(
      allTimeRange,
      {
        name: TimeRangePreset.CUSTOM,
        start: new Date("0000-12-31T00:00:00.000Z"),
        end: new Date("0001-01-02T00:00:00.000Z"),
      },
      undefined,
      "UTC",
      allTimeRange,
      V1TimeGrain.TIME_GRAIN_HOUR,
      true,
    );

    expect(state?.timeStart).toBe("0001-01-01T00:00:00.000Z");
    expect(state?.timeEnd).toBe("0001-01-02T00:00:00.000Z");
    expect(state?.adjustedStart).toBe("0001-01-01T00:00:00.000Z");
  });

  it("keeps a scrub range within the timestamp bounds", () => {
    const state = calculateTimeRangePartial(
      allTimeRange,
      { name: "P30D" } as DashboardTimeControls,
      {
        start: new Date("0000-12-31T00:00:00.000Z"),
        end: new Date("0001-01-02T00:00:00.000Z"),
        isScrubbing: false,
      },
      "UTC",
      allTimeRange,
      V1TimeGrain.TIME_GRAIN_HOUR,
      true,
    );

    expect(state?.timeStart).toBe("0001-01-01T00:00:00.000Z");
    expect(state?.timeEnd).toBe("0001-01-02T00:00:00.000Z");
  });
});

describe("calculateComparisonTimeRangePartial", () => {
  const selectedTimeRange: DashboardTimeControls = {
    name: TimeRangePreset.CUSTOM,
    start: new Date("0001-01-02T00:00:00.000Z"),
    end: new Date("0001-01-03T00:00:00.000Z"),
  };

  function getTimeRangeState() {
    const state = calculateTimeRangePartial(
      allTimeRange,
      selectedTimeRange,
      undefined,
      "UTC",
      allTimeRange,
      V1TimeGrain.TIME_GRAIN_HOUR,
      true,
    );
    if (!state) throw new Error("Expected a time range state");
    return state;
  }

  it("keeps a custom comparison range from the URL within the timestamp bounds", () => {
    const state = calculateComparisonTimeRangePartial(
      undefined,
      allTimeRange,
      {
        name: TimeComparisonOption.CUSTOM,
        start: new Date("0000-12-31T00:00:00.000Z"),
        end: new Date("0001-01-01T00:00:00.000Z"),
      },
      "UTC",
      undefined,
      true,
      getTimeRangeState(),
    );

    expect(state.comparisonTimeStart).toBe("0001-01-01T00:00:00.000Z");
    expect(state.comparisonTimeEnd).toBe("0001-01-01T00:00:00.000Z");
  });

  it("keeps a comparison range shifted by a scrub within the timestamp bounds", () => {
    // The previous day is 0001-01-01 to 0001-01-02. A scrub starting 12 hours before the selected range
    // shifts the comparison start by the same amount, into year 0.
    const state = calculateComparisonTimeRangePartial(
      undefined,
      allTimeRange,
      { name: TimeComparisonOption.DAY } as DashboardTimeControls,
      "UTC",
      {
        start: new Date("0001-01-01T12:00:00.000Z"),
        end: new Date("0001-01-02T12:00:00.000Z"),
        isScrubbing: false,
      },
      true,
      getTimeRangeState(),
    );

    expect(state.comparisonTimeStart).toBe("0001-01-01T00:00:00.000Z");
    expect(state.comparisonTimeEnd).toBe("0001-01-01T12:00:00.000Z");
  });
});
