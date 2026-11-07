import { calculateTimeRangePartial } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { MIN_TIMESTAMP_MS } from "@rilldata/web-common/lib/time/timestamp-bounds";
import {
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

// Data with placeholder dates such as 0001-01-01 has bounds in year 1.
// A legacy preset resolved locally against those bounds, or a scrub reaching before them,
// would otherwise land in year 0, which the request encoder rejects.
describe("calculateTimeRangePartial", () => {
  const allTimeRange: DashboardTimeControls = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date("0001-01-01T01:00:00.000Z"),
    end: new Date("0001-01-02T00:00:00.000Z"),
  };

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
