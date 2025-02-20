import { formatRefreshSchedule } from "@rilldata/web-admin/features/scheduled-reports/metadata/utils";
import {
  convertFormValuesToCronExpression,
  getFrequencyFromCronExpression,
  ReportFrequency,
} from "@rilldata/web-common/features/scheduled-reports/time-utils";
import { describe, expect, it } from "vitest";

describe("Reports metadata", () => {
  describe("formatRefreshSchedule", () => {
    const TestCase: [string, string][] = [
      [
        "* * * 21 * *",
        "Every second, every minute, every hour, on the 21st of each month",
      ],
      // TODO: we need a follow up to better match what is expected.
      //       perhaps we should replace cronstrue
      ["0 0 * * 1", "At 12:00 AM, only on Monday"],
    ];
    for (const [cron, formattedSchedule] of TestCase) {
      it(`format(${cron})=${formattedSchedule}`, () => {
        expect(formatRefreshSchedule(cron)).toEqual(formattedSchedule);
      });
    }
  });

  describe("frequency extraction and conversion", () => {
    const TestCase: [
      frequency: ReportFrequency,
      dayOfWeek: string,
      timeOfDay: string,
      dayOfMonth: number,
      expectedCron: string,
    ][] = [
      [ReportFrequency.Daily, "Monday", "11:20", 1, "20 11 * * *"],
      [ReportFrequency.Weekly, "Tuesday", "11:20", 1, "20 11 * * 2"],
      [ReportFrequency.Weekdays, "Tuesday", "11:20", 1, "20 11 * * 1-5"],
      [ReportFrequency.Monthly, "Tuesday", "11:20", 11, "20 11 11 * *"],
    ];

    for (const [
      frequency,
      dayOfWeek,
      timeOfDay,
      dayOfMonth,
      expectedCron,
    ] of TestCase) {
      it(expectedCron, () => {
        const actualCron = convertFormValuesToCronExpression(
          frequency,
          dayOfWeek,
          timeOfDay,
          dayOfMonth,
        );
        expect(actualCron).toEqual(expectedCron);
        expect(getFrequencyFromCronExpression(actualCron)).toEqual(frequency);
      });
    }
  });
});
