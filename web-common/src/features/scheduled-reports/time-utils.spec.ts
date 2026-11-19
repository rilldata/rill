import { DateTime, Settings } from "luxon";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  ReportFrequency,
  convertFormValuesToCronExpression,
  getExistingScheduleFormValues,
  getInitialScheduleFormValues,
  getTodaysDayOfWeek,
} from "./time-utils";

describe("weekly schedule day of week", () => {
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

  it("keeps the day of an existing weekly schedule", () => {
    const values = getExistingScheduleFormValues({ cron: "30 10 * * 1" });
    expect(values.frequency).toEqual(ReportFrequency.Weekly);
    expect(values.dayOfWeek).toEqual("Monday");
  });

  it("repairs an existing weekly schedule saved with an invalid day", () => {
    const values = getExistingScheduleFormValues({
      cron: "30 10 * * undefined",
    });
    expect(values.frequency).toEqual(ReportFrequency.Weekly);
    expect(values.dayOfWeek).toEqual("Wednesday");
    expect(
      convertFormValuesToCronExpression(
        values.frequency,
        values.dayOfWeek,
        values.timeOfDay,
        values.dayOfMonth,
      ),
    ).toEqual("30 10 * * 3");
  });

  it("builds a valid cron when an existing daily schedule is switched to weekly", () => {
    const values = getExistingScheduleFormValues({ cron: "0 9 * * *" });
    expect(values.frequency).toEqual(ReportFrequency.Daily);
    expect(
      convertFormValuesToCronExpression(
        ReportFrequency.Weekly,
        values.dayOfWeek,
        values.timeOfDay,
        values.dayOfMonth,
      ),
    ).toEqual("0 9 * * 3");
  });
});
