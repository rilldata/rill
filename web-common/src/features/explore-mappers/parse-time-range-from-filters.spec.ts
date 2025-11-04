import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { parseTimeRangeFromFilters } from "@rilldata/web-common/features/explore-mappers/parse-time-range-from-filters.ts";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  type V1Expression,
  V1Operation,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

const CONSISTENT_TIME_RANGE_SUMMARY: V1TimeRangeSummary = {
  min: "2025-10-01T00:00:00Z",
  max: "2025-11-01T00:00:00Z",
};

describe("parseTimeRangeFromFilters", () => {
  const TestCases: {
    title: string;
    expression: V1Expression;
    expectedTimeRange: DashboardTimeControls;
  }[] = [
    {
      title: "equality time dimension filter",
      expression: createBinaryExpression(
        AD_BIDS_TIMESTAMP_DIMENSION,
        V1Operation.OPERATION_EQ,
        "2025-10-10T10:00:00Z",
      ),
      expectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-10-10T10:00:00Z"),
        end: new Date("2025-10-10T11:00:00Z"),
      },
    },
    {
      title: "less than time dimension filter",
      expression: createBinaryExpression(
        AD_BIDS_TIMESTAMP_DIMENSION,
        V1Operation.OPERATION_LT,
        "2025-10-10T10:00:00Z",
      ),
      expectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-10-01T00:00:00Z"),
        end: new Date("2025-10-10T10:00:00Z"),
      },
    },
    {
      title: "less than or equals time dimension filter",
      expression: createBinaryExpression(
        AD_BIDS_TIMESTAMP_DIMENSION,
        V1Operation.OPERATION_LTE,
        "2025-10-10T10:00:00Z",
      ),
      expectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-10-01T00:00:00Z"),
        end: new Date("2025-10-10T11:00:00Z"),
      },
    },
    {
      title: "greater than time dimension filter",
      expression: createBinaryExpression(
        AD_BIDS_TIMESTAMP_DIMENSION,
        V1Operation.OPERATION_GT,
        "2025-10-10T10:00:00Z",
      ),
      expectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-10-10T11:00:00Z"),
        end: new Date("2025-11-01T00:00:00Z"),
      },
    },
    {
      title: "greater than or equals time dimension filter",
      expression: createBinaryExpression(
        AD_BIDS_TIMESTAMP_DIMENSION,
        V1Operation.OPERATION_GTE,
        "2025-10-10T10:00:00Z",
      ),
      expectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-10-10T10:00:00Z"),
        end: new Date("2025-11-01T00:00:00Z"),
      },
    },
    {
      title:
        "greater than or equals and less than or equals time with additional filters",
      expression: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
        createBinaryExpression(
          AD_BIDS_TIMESTAMP_DIMENSION,
          V1Operation.OPERATION_GTE,
          "2025-10-04T00:00:00Z",
        ),
        createBinaryExpression(
          AD_BIDS_TIMESTAMP_DIMENSION,
          V1Operation.OPERATION_LTE,
          "2025-10-22T00:00:00Z",
        ),
      ]),
      expectedTimeRange: {
        name: TimeRangePreset.CUSTOM,
        start: new Date("2025-10-04T00:00:00Z"),
        end: new Date("2025-10-23T00:00:00Z"),
      },
    },
  ];

  for (const { title, expression, expectedTimeRange } of TestCases) {
    it(title, () => {
      expect(
        parseTimeRangeFromFilters(
          expression,
          AD_BIDS_TIMESTAMP_DIMENSION,
          "UTC",
          CONSISTENT_TIME_RANGE_SUMMARY,
        ),
      ).toEqual(expectedTimeRange);
    });
  }
});
