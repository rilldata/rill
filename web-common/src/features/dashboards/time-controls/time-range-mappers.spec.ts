import {
  type DashboardTimeControls,
  TimeComparisonOption,
} from "@rilldata/web-common/lib/time/types";
import { describe, expect, it } from "vitest";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapV1TimeRangeToSelectedComparisonTimeRange,
} from "./time-range-mappers";

const timeRangeSummary = {
  min: "2024-01-01T00:00:00Z",
  max: "2024-03-01T00:00:00Z",
};

describe("mapV1TimeRangeToSelectedComparisonTimeRange", () => {
  const options = [
    TimeComparisonOption.CONTIGUOUS,
    TimeComparisonOption.DAY,
    TimeComparisonOption.WEEK,
    TimeComparisonOption.MONTH,
    TimeComparisonOption.QUARTER,
    TimeComparisonOption.YEAR,
  ];

  it.each(options)(
    "reads back the %s comparison of a rill time expression as it is saved",
    (option) => {
      const saved = mapSelectedComparisonTimeRangeToV1TimeRange(
        { name: option } as DashboardTimeControls,
        true,
        { expression: "MTD as of latest/D" },
      );
      const selected = mapV1TimeRangeToSelectedComparisonTimeRange(
        saved!,
        timeRangeSummary,
        timeRangeSummary.max,
      );
      expect(selected?.name).toBe(option);
    },
  );

  it("reads an expression with another offset as the previous period, as before", () => {
    const selected = mapV1TimeRangeToSelectedComparisonTimeRange(
      { expression: "7D as of latest/D offset -7D" },
      timeRangeSummary,
      timeRangeSummary.max,
    );
    expect(selected?.name).toBe(TimeComparisonOption.CONTIGUOUS);
  });
});
