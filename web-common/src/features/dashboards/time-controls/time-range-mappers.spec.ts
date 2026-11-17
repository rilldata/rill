import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import {
  mapTimeManagerComparisonRangeToV1TimeRange,
  mapV1TimeRangeToSelectedComparisonTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("mapTimeManagerComparisonRangeToV1TimeRange", () => {
  const TestCases: {
    title: string;
    timeRange: V1TimeRange;
    comparison: TimeComparisonOption;
    expected: V1TimeRange;
  }[] = [
    {
      // The runtime resolves the legacy expression and its offset the same way, so the ranges line up
      title: "offsets a legacy ISO expression",
      timeRange: { expression: "P1D", timeZone: "UTC" },
      comparison: TimeComparisonOption.DAY,
      expected: { expression: "P1D offset -1D" },
    },
    {
      title: "offsets a legacy ISO expression by its previous period",
      timeRange: { expression: "P7D", timeZone: "UTC" },
      comparison: TimeComparisonOption.CONTIGUOUS,
      expected: { expression: "P7D offset -1P" },
    },
    {
      title: "offsets a rill time expression",
      timeRange: { expression: "7D as of watermark/D+1D", timeZone: "UTC" },
      comparison: TimeComparisonOption.WEEK,
      expected: { expression: "7D as of watermark/D+1D offset -1W" },
    },
    {
      // The runtime can't offset all time
      title: "keeps legacy fields for all time",
      timeRange: { expression: "inf", timeZone: "UTC" },
      comparison: TimeComparisonOption.DAY,
      expected: { isoDuration: "inf", isoOffset: TimeComparisonOption.DAY },
    },
    {
      title: "keeps legacy fields for a legacy ISO duration",
      timeRange: { isoDuration: "P1D", timeZone: "UTC" },
      comparison: TimeComparisonOption.DAY,
      expected: { isoDuration: "P1D", isoOffset: TimeComparisonOption.DAY },
    },
  ];

  for (const { title, timeRange, comparison, expected } of TestCases) {
    it(title, () => {
      expect(
        mapTimeManagerComparisonRangeToV1TimeRange(
          {
            showComparison: true,
            comparisonTimeRange: comparison,
            isCustomComparison: false,
          } as TimeFilterManager,
          timeRange,
        ),
      ).toEqual(expected);
    });
  }
});

describe("mapV1TimeRangeToSelectedComparisonTimeRange", () => {
  const TimeRangeSummary = {
    min: "2024-01-01T00:00:00Z",
    max: "2024-02-01T00:00:00Z",
  };

  const TestCases: {
    title: string;
    comparisonTimeRange: V1TimeRange;
    expected: TimeComparisonOption;
  }[] = [
    {
      title: "reads the comparison from the offset of an expression",
      comparisonTimeRange: { expression: "P7D offset -1W" },
      expected: TimeComparisonOption.WEEK,
    },
    {
      title: "reads the previous period from the offset of an expression",
      comparisonTimeRange: { expression: "P1D offset -1P" },
      expected: TimeComparisonOption.CONTIGUOUS,
    },
    {
      title: "reads the comparison from a legacy offset",
      comparisonTimeRange: {
        isoDuration: "P1D",
        isoOffset: TimeComparisonOption.DAY,
      },
      expected: TimeComparisonOption.DAY,
    },
    {
      title: "reads the previous period from a legacy offset",
      comparisonTimeRange: { isoDuration: "P7D", isoOffset: "P7D" },
      expected: TimeComparisonOption.CONTIGUOUS,
    },
  ];

  for (const { title, comparisonTimeRange, expected } of TestCases) {
    it(title, () => {
      expect(
        mapV1TimeRangeToSelectedComparisonTimeRange(
          comparisonTimeRange,
          TimeRangeSummary,
          TimeRangeSummary.max,
        )?.name,
      ).toBe(expected);
    });
  }
});
