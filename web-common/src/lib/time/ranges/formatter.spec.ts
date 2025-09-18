import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { describe, expect, it } from "vitest";

describe("prettyFormatTimeRange", () => {
  describe("one point", () => {
    const singlePointTestCases = [
      {
        test: "Non-zero minute, minute grain",
        time: "2025-09-04T08:10:20.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MINUTE,
        formattedTime: "Sep 4, 2025 (8:10:20AM)",
      },
      {
        test: "Non-zero minute, hour grain",
        time: "2025-09-04T08:10:20.000Z",
        grain: V1TimeGrain.TIME_GRAIN_HOUR,
        formattedTime: "Sep 4, 2025 (8:10:20AM)",
      },

      {
        test: "Non-zero hour, zero minute, minute grain",
        time: "2025-09-04T08:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MINUTE,
        formattedTime: "Sep 4, 2025 (8:00AM)",
      },
      {
        test: "Non-zero hour, zero minute, hour grain",
        time: "2025-09-04T08:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_HOUR,
        formattedTime: "Sep 4, 2025 (8AM)",
      },
      {
        test: "Non-zero hour, zero minute, day grain",
        time: "2025-09-04T08:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 4, 2025 (8AM)",
      },

      {
        test: "Non-1st of month at midnight, hour grain",
        time: "2025-09-05T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_HOUR,
        formattedTime: "Sep 5, 2025 (12AM)",
      },
      {
        test: "Non-1st of month at midnight, day grain",
        time: "2025-09-05T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 5, 2025",
      },
      {
        test: "Non-1st of month at midnight, month grain",
        time: "2025-09-05T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Sep 5, 2025",
      },

      {
        test: "1st of month at midnight, hour grain",
        time: "2025-08-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_HOUR,
        formattedTime: "Aug 1, 2025 (12AM)",
      },
      {
        test: "1st of month at midnight, day grain",
        time: "2025-08-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Aug 1, 2025",
      },
      {
        test: "1st of month at midnight, month grain",
        time: "2025-08-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Aug 2025",
      },
      {
        test: "1st of month at midnight, quarter grain",
        time: "2025-08-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_QUARTER,
        formattedTime: "Aug 2025",
      },

      {
        test: "Jan 1st of month at midnight, day grain",
        time: "2025-01-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Jan 1, 2025",
      },
      {
        test: "Jan 1st of month at midnight, month grain",
        time: "2025-01-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Jan 2025",
      },
      {
        test: "Jan 1st of month at midnight, quarter grain",
        time: "2025-01-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_QUARTER,
        formattedTime: "Jan 2025",
      },
      {
        test: "Jan 1st of month at midnight, year grain",
        time: "2025-01-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_YEAR,
        formattedTime: "2025",
      },
    ];

    for (const { test, time, grain, formattedTime } of singlePointTestCases) {
      it(test, () => {
        const newFormat = prettyFormatTimeRange(
          Interval.fromDateTimes(
            DateTime.fromISO(time).setZone("UTC"),
            DateTime.fromISO(time).setZone("UTC"),
          ),
          grain,
        );
        expect(newFormat).toEqual(formattedTime);
      });
    }
  });

  describe("two points", () => {
    const twoPointsTestCases = [
      {
        test: "Same day, full time difference",
        start: "2025-09-01T08:10:20.000Z",
        end: "2025-09-01T16:15:30.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MINUTE,
        formattedTime: "Sep 1, 2025 (8:10:20AM-4:15:30PM)",
      },
      {
        test: "Different days, full time difference",
        start: "2025-09-01T08:10:20.000Z",
        end: "2025-09-04T16:15:30.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MINUTE,
        formattedTime: "Sep 1 – 4, 2025 (8:10:20AM-4:15:30PM)",
      },
      {
        test: "Same day, hour difference, minute grain",
        start: "2025-09-01T14:00:00.000Z",
        end: "2025-09-01T22:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MINUTE,
        formattedTime: "Sep 1, 2025 (2:00PM-10:00PM)",
      },
      {
        test: "Same day, hour difference, zero minute, hour grain",
        start: "2025-09-01T14:00:00.000Z",
        end: "2025-09-01T22:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_HOUR,
        formattedTime: "Sep 1, 2025 (2PM-10PM)",
      },
      {
        test: "Same day, hour difference, non-zero minute, hour grain",
        start: "2025-09-01T14:15:00.000Z",
        end: "2025-09-01T22:15:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_HOUR,
        formattedTime: "Sep 1, 2025 (2:15PM-10:15PM)",
      },
      {
        test: "Different days, hour difference",
        start: "2025-09-01T16:45:00.000Z",
        end: "2025-09-05T02:45:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 1 – 5, 2025 (4:45PM-2:45AM)",
      },

      {
        test: "Same month, days difference, same time at midnight",
        start: "2025-09-02T00:00:00.000Z",
        end: "2025-09-06T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 2 – 5, 2025",
      },
      {
        test: "Same month, days difference, same time not at midnight",
        start: "2025-09-01T16:00:00.000Z",
        end: "2025-09-04T16:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 1 – 4, 2025 (4PM)",
      },

      {
        test: "Same year different months, day grain",
        start: "2025-09-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 1 – Oct 31, 2025",
      },
      {
        test: "Same year different months, week grain",
        start: "2025-09-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_WEEK,
        formattedTime: "Sep 1 – Oct 31, 2025",
      },
      {
        test: "Same year different months, month grain",
        start: "2025-09-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Sep – Oct 2025",
      },
      {
        test: "Same year different months, non-1st day, month grain",
        start: "2025-09-05T00:00:00.000Z",
        end: "2025-11-05T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Sep 5 – Nov 4, 2025",
      },
      {
        test: "Same year different months, time not at midnight, month grain",
        start: "2025-09-01T14:30:00.000Z",
        end: "2025-11-01T14:30:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Sep – Nov 2025 (2:30PM)",
      },

      {
        test: "Different years and months with same day and time, day grain",
        start: "2024-09-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Sep 1, 2024 – Oct 31, 2025",
      },
      {
        test: "Different years and months with same day and time, month grain",
        start: "2024-09-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Sep 2024 – Oct 2025",
      },

      {
        test: "Different years everything else same, day grain",
        start: "2024-11-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_DAY,
        formattedTime: "Nov 1, 2024 – Oct 31, 2025",
      },
      {
        test: "Different years everything else same, month grain",
        start: "2024-11-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_MONTH,
        formattedTime: "Nov 2024 – Oct 2025",
      },
      {
        test: "Different years everything else same and non-jan month, year grain",
        start: "2024-11-01T00:00:00.000Z",
        end: "2025-11-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_YEAR,
        formattedTime: "Nov 2024 – Oct 2025",
      },
      {
        test: "Different years everything else same and jan month, year grain",
        start: "2024-01-01T00:00:00.000Z",
        end: "2025-01-01T00:00:00.000Z",
        grain: V1TimeGrain.TIME_GRAIN_YEAR,
        formattedTime: "2024",
      },
    ];

    twoPointsTestCases.forEach(({ test, start, end, grain, formattedTime }) => {
      it(test, () => {
        const interval = Interval.fromDateTimes(
          DateTime.fromISO(start).setZone("UTC"),
          DateTime.fromISO(end).setZone("UTC"),
        );
        const actualFormattedTime = prettyFormatTimeRange(interval, grain);
        expect(actualFormattedTime).toEqual(formattedTime);
      });
    });
  });
});
