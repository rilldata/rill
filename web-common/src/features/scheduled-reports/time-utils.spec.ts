import { DateTime, Settings } from "luxon";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ReportFrequency,
  convertFormValuesToCronExpression,
  getInitialScheduleFormValues,
  getTodaysDayOfWeek,
} from "./time-utils";

describe("getTodaysDayOfWeek", () => {
  const originalLocale = Settings.defaultLocale;

  beforeEach(() => {
    vi.useFakeTimers();
    // A Wednesday.
    vi.setSystemTime(DateTime.fromISO("2026-10-21T10:00:00").toJSDate());
  });

  afterEach(() => {
    vi.useRealTimers();
    Settings.defaultLocale = originalLocale;
  });

  it("returns the English day name whatever the app locale is", () => {
    for (const locale of ["en-US", "es", "es-PE"]) {
      Settings.defaultLocale = locale;
      expect(getTodaysDayOfWeek()).toEqual("Wednesday");
    }
  });

  it("builds a valid weekly cron from the default form values in a non-English locale", () => {
    Settings.defaultLocale = "es";
    const values = getInitialScheduleFormValues();
    expect(
      convertFormValuesToCronExpression(
        ReportFrequency.Weekly,
        values.dayOfWeek,
        "10:30",
        values.dayOfMonth,
      ),
    ).toEqual("30 10 * * 3");
  });
});
