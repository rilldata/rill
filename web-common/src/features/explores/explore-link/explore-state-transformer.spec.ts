import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { transformTimeAndFiltersToExploreState } from "@rilldata/web-common/features/explores/explore-link/explore-state-transformer";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const START = new Date("2024-01-01T00:00:00Z");
const END = new Date("2024-01-08T00:00:00Z");

function store(partial: Partial<TimeAndFilterStore>): TimeAndFilterStore {
  return {
    timeRange: {},
    comparisonTimeRange: undefined,
    where: undefined,
    timeGrain: undefined,
    showTimeComparison: false,
    timeRangeState: undefined,
    comparisonTimeRangeState: undefined,
    hasTimeSeries: true,
    ...partial,
  };
}

describe("transformTimeAndFiltersToExploreState", () => {
  it("carries the grain over as a fixed grain", () => {
    const exploreState = transformTimeAndFiltersToExploreState(
      store({
        timeRange: { timeZone: "UTC" },
        timeRangeState: {
          selectedTimeRange: {
            name: "P7D",
            start: START,
            end: END,
            interval: V1TimeGrain.TIME_GRAIN_HOUR,
          },
          timeStart: START.toISOString(),
          timeEnd: END.toISOString(),
        },
      }),
    );

    expect(exploreState.selectedTimeRange?.interval).toBe(
      V1TimeGrain.TIME_GRAIN_HOUR,
    );
    expect(exploreState.adaptiveTimeGrain).toBe(false);
  });

  it("carries a custom range's grain over as a fixed grain", () => {
    const exploreState = transformTimeAndFiltersToExploreState(
      store({
        timeRange: {
          start: START.toISOString(),
          end: END.toISOString(),
          timeZone: "UTC",
        },
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      }),
    );

    expect(exploreState.selectedTimeRange?.name).toBe(TimeRangePreset.CUSTOM);
    expect(exploreState.adaptiveTimeGrain).toBe(false);
  });

  it("leaves the grain mode alone when no grain is carried", () => {
    const exploreState = transformTimeAndFiltersToExploreState(store({}));

    expect(exploreState.selectedTimeRange).toBeUndefined();
    expect(exploreState.adaptiveTimeGrain).toBeUndefined();
  });
});
