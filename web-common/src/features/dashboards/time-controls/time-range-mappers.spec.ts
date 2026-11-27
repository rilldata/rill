import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import {
  convertTimeManagerComparisonRangeToV1TimeRange,
  convertV1TimeRangesToTimeFilterUrlParams,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("convertTimeManagerComparisonRangeToV1TimeRange", () => {
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
    {
      // The runtime rejects a comparison on a different time dimension than the primary range
      title: "keeps the time dimension of the primary range",
      timeRange: {
        expression: "7D",
        timeZone: "UTC",
        timeDimension: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
      },
      comparison: TimeComparisonOption.WEEK,
      expected: {
        expression: "7D offset -1W",
        timeDimension: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
      },
    },
  ];

  for (const { title, timeRange, comparison, expected } of TestCases) {
    it(title, () => {
      expect(
        convertTimeManagerComparisonRangeToV1TimeRange(
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

describe("convertV1TimeRangeToComparisonTimeOption", () => {
  const TestCases: {
    title: string;
    timeRange: V1TimeRange;
    comparisonTimeRange: V1TimeRange;
    expected: TimeComparisonOption;
  }[] = [
    {
      title: "reads the comparison from the offset of an expression",
      timeRange: { expression: "P7D" },
      comparisonTimeRange: { expression: "P7D offset -1W" },
      expected: TimeComparisonOption.WEEK,
    },
    {
      title: "reads the previous period from the offset of an expression",
      timeRange: { expression: "P1D" },
      comparisonTimeRange: { expression: "P1D offset -1P" },
      expected: TimeComparisonOption.CONTIGUOUS,
    },
    {
      title: "reads the comparison from a legacy offset",
      timeRange: { isoDuration: "P1D" },
      comparisonTimeRange: {
        isoDuration: "P1D",
        isoOffset: TimeComparisonOption.DAY,
      },
      expected: TimeComparisonOption.DAY,
    },
    {
      title: "reads the previous period from a legacy offset",
      timeRange: { isoDuration: "P7D" },
      comparisonTimeRange: { isoDuration: "P7D", isoOffset: "P7D" },
      expected: TimeComparisonOption.CONTIGUOUS,
    },
    {
      title:
        "reads the previous period from a legacy duration without an offset",
      timeRange: { isoDuration: "P7D" },
      comparisonTimeRange: { isoDuration: "P7D" },
      expected: TimeComparisonOption.CONTIGUOUS,
    },
  ];

  for (const { title, timeRange, comparisonTimeRange, expected } of TestCases) {
    it(title, () => {
      const urlParams = convertV1TimeRangesToTimeFilterUrlParams(
        timeRange,
        comparisonTimeRange,
      );
      expect(urlParams.get(ExploreStateURLParams.ComparisonTimeRange)).toBe(
        expected,
      );

      // The dashboard opens with the comparison on.
      const { partialExploreState, errors } =
        convertURLSearchParamsToExploreState(
          urlParams,
          AD_BIDS_METRICS_INIT_WITH_TIME,
          AD_BIDS_EXPLORE_INIT,
          {},
        );
      expect(errors).toEqual([]);
      expect(partialExploreState.showTimeComparison).toBe(true);
      expect(partialExploreState.selectedComparisonTimeRange?.name).toBe(
        expected,
      );
    });
  }

  it("keeps an unparsable absolute comparison as is instead of throwing", () => {
    const urlParams = convertV1TimeRangesToTimeFilterUrlParams(
      { expression: "P7D" },
      { start: "garbage", end: "2024-02-01T00:00:00Z" },
    );
    expect(urlParams.get(ExploreStateURLParams.ComparisonTimeRange)).toBe(
      "garbage,2024-02-01T00:00:00Z",
    );
  });
});

describe("convertV1TimeRangesToTimeFilterUrlParams", () => {
  const TestCases: {
    title: string;
    timeRange: V1TimeRange;
    expectedTimeRange: string;
    expectedTimeZone: string | null;
    expectedTimeDimension?: string;
  }[] = [
    {
      title: "keeps the saved zone of an absolute range",
      timeRange: {
        start: "2026-03-01T05:00:00Z",
        end: "2026-03-02T05:00:00Z",
        timeZone: "America/New_York",
      },
      expectedTimeRange: "2026-03-01T00:00:00.000Z,2026-03-02T00:00:00.000Z",
      expectedTimeZone: "America/New_York",
    },
    {
      // Otherwise the dashboard's default zone would shift the UTC endpoints
      title: "pins an absolute range without a saved zone to UTC",
      timeRange: {
        start: "2026-03-01T05:00:00Z",
        end: "2026-03-02T05:00:00Z",
      },
      expectedTimeRange: "2026-03-01T05:00:00.000Z,2026-03-02T05:00:00.000Z",
      expectedTimeZone: "UTC",
    },
    {
      title: "leaves the zone unset for a relative range without a saved zone",
      timeRange: { expression: "7D as of watermark/D+1D" },
      expectedTimeRange: "7D as of watermark/D+1D",
      expectedTimeZone: null,
    },
    {
      // Otherwise saving the form would drop the time dimension from the query
      title: "keeps the time dimension of the range",
      timeRange: {
        expression: "7D",
        timeDimension: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
      },
      expectedTimeRange: "7D",
      expectedTimeZone: null,
      expectedTimeDimension: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
    },
  ];

  for (const {
    title,
    timeRange,
    expectedTimeRange,
    expectedTimeZone,
    expectedTimeDimension,
  } of TestCases) {
    it(title, () => {
      const params = convertV1TimeRangesToTimeFilterUrlParams(
        timeRange,
        undefined,
      );
      expect(params.get(ExploreStateURLParams.TimeRange)).toBe(
        expectedTimeRange,
      );
      expect(params.get(ExploreStateURLParams.TimeZone)).toBe(expectedTimeZone);
      expect(params.get(ExploreStateURLParams.TimeDimension)).toBe(
        expectedTimeDimension ?? null,
      );
    });
  }
});
