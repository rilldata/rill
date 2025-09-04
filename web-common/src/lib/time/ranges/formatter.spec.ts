import {
  prettyFormatTimeRange,
  prettyFormatTimeRangeV2,
} from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("prettyFormatTimeRange", () => {
  const twoPointsTestCases = [
    {
      test: "Same day, full time difference",
      start: "2025-09-01T08:10:20.000Z",
      end: "2025-09-01T16:15:30.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MINUTE,
      formattedTime: "Sep 1, 2025 (1:55PM-10:00PM)",
    },
    {
      test: "Different days, full time difference",
      start: "2025-09-01T08:10:20.000Z",
      end: "2025-09-04T16:15:30.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MINUTE,
      formattedTime: "Sep 1 - 4, 2025 (1:55PM-10:00PM)",
    },
    {
      test: "Same day, hour difference, minute grain",
      start: "2025-09-01T08:15:00.000Z",
      end: "2025-09-01T16:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MINUTE,
      formattedTime: "Sep 1, 2025 (2:00PM-10:00PM)",
    },
    {
      test: "Same day, hour difference, zero minute, hour grain",
      start: "2025-09-01T08:15:00.000Z",
      end: "2025-09-01T16:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_HOUR,
      formattedTime: "Sep 1, 2025 (2PM-10PM)",
    },
    {
      test: "Same day, hour difference, non-zero minute, hour grain",
      start: "2025-09-01T08:30:00.000Z",
      end: "2025-09-01T16:30:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_HOUR,
      formattedTime: "Sep 1, 2025 (2:15PM-10:15PM)",
    },
    {
      test: "Different days, hour difference",
      start: "2025-09-01T11:00:00.000Z",
      end: "2025-09-04T21:00:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_DAY,
      formattedTime: "Sep 1 - 5, 2025 (4:45PM-2:45AM)",
    },

    {
      test: "Same month, days difference, same time at midnight",
      start: "2025-09-01T18:15:00.000Z",
      end: "2025-09-04T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_DAY,
      formattedTime: "Sep 2 - 5, 2025",
    },
    {
      test: "Same month, days difference, same time not at midnight",
      start: "2025-09-01T10:15:00.000Z",
      end: "2025-09-04T10:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_DAY,
      formattedTime: "Sep 1 - 4, 2025 (4PM)",
    },

    {
      test: "Same year different months, day grain",
      start: "2025-08-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_DAY,
      formattedTime: "Sep 1 - Nov 1, 2025",
    },
    {
      test: "Same year different months, week grain",
      start: "2025-08-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_WEEK,
      formattedTime: "Sep 1 - Nov 1, 2025",
    },
    {
      test: "Same year different months, month grain",
      start: "2025-08-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MONTH,
      formattedTime: "Sep - Nov 2025",
    },
    {
      test: "Same year different months, non-1st day, month grain",
      start: "2025-09-04T18:15:00.000Z",
      end: "2025-11-04T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MONTH,
      formattedTime: "Sep 5 - Nov 5, 2025",
    },
    {
      test: "Same year different months, time not on day boundary, month grain",
      start: "2025-09-01T08:45:00.000Z",
      end: "2025-11-01T08:45:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MONTH,
      formattedTime: "Sep - Nov 2025 (2:30PM)",
    },

    {
      test: "Different years and months with same day and time, day grain",
      start: "2024-08-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_DAY,
      formattedTime: "Sep 1, 2024 - Nov 1, 2025",
    },
    {
      test: "Different years and months with same day and time, month grain",
      start: "2024-08-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MONTH,
      formattedTime: "Sep 2024 - Nov 2025",
    },

    {
      test: "Different years everything else same, day grain",
      start: "2024-10-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_DAY,
      formattedTime: "Nov 1, 2024 - 2025",
    },
    {
      test: "Different years everything else same, month grain",
      start: "2024-10-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_MONTH,
      formattedTime: "Nov 2024 - 2025",
    },
    {
      test: "Different years everything else same and non-jan month, year grain",
      start: "2024-10-31T18:15:00.000Z",
      end: "2025-10-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_YEAR,
      formattedTime: "Nov 2024 - 2025",
    },
    {
      test: "Different years everything else same and jan month, year grain",
      start: "2023-12-31T18:15:00.000Z",
      end: "2024-12-31T18:15:00.000Z",
      grain: V1TimeGrain.TIME_GRAIN_YEAR,
      formattedTime: "2024 - 2025",
    },
  ];

  describe("two points", () => {
    twoPointsTestCases.forEach(({ test, start, end, grain, formattedTime }) => {
      it(test, () => {
        const oldFormat = prettyFormatTimeRange(
          new Date(start),
          new Date(end),
          undefined,
          "Asia/Kathmandu",
        );
        const newFormat = prettyFormatTimeRangeV2(
          new Date(start),
          new Date(end),
          grain,
          "Asia/Kathmandu",
        );
        console.log(`${test}\nOld: ${oldFormat}\nNew: ${newFormat}\n`);
        expect(newFormat).toEqual(formattedTime);
      });
    });
  });
});
