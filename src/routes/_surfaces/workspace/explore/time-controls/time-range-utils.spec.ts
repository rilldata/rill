import {
  TimeGrain,
  TimeRangeName,
} from "$common/database-service/DatabaseTimeSeriesActions";
import {
  getDefaultTimeGrain,
  getDefaultTimeRangeName,
  getSelectableTimeGrains,
  getSelectableTimeRangeNames,
  makeTimeRange,
} from "./time-range-utils";

describe("getSelectableTimeRangeNames", () => {
  it("should return an array of available time range names", () => {
    const timeRangeNames = getSelectableTimeRangeNames({
      start: "2020-03-01",
      end: "2020-03-31",
    });
    expect(timeRangeNames).toEqual([
      // TimeRangeName.LastHour,
      TimeRangeName.Last6Hours,
      TimeRangeName.LastDay,
      TimeRangeName.Last2Days,
      TimeRangeName.Last5Days,
      TimeRangeName.LastWeek,
      TimeRangeName.Last2Weeks,
      TimeRangeName.Last30Days,
      TimeRangeName.AllTime,
    ]);
  });
});

describe("getDefaultTimeRangeName", () => {
  it("should return the default time range name", () => {
    const timeRangeName = getDefaultTimeRangeName();
    expect(timeRangeName).toEqual(TimeRangeName.AllTime);
  });
});

describe("getSelectableTimeGrains", () => {
  it("should return an array of available time grains", () => {
    const timeGrains = getSelectableTimeGrains(TimeRangeName.Last30Days, {
      start: "2020-03-01",
      end: "2020-03-31",
    });
    expect(timeGrains).toEqual([
      {
        enabled: true,
        timeGrain: "1 hour",
      },
      {
        enabled: true,
        timeGrain: "1 day",
      },
      {
        enabled: true,
        timeGrain: "7 day",
      },
      {
        enabled: false,
        timeGrain: "1 month",
      },
      {
        enabled: false,
        timeGrain: "1 year",
      },
    ]);
  });
});

describe("getDefaultTimeGrain", () => {
  it("should return the default time grain", () => {
    const timeGrain = getDefaultTimeGrain(TimeRangeName.Last30Days);
    expect(timeGrain).toEqual(TimeGrain.OneDay);
  });
});

describe("makeTimeRange", () => {
  it("should create a TimeRange object representing the Last Two Weeks", () => {
    expect(
      makeTimeRange(TimeRangeName.Last2Weeks, TimeGrain.OneDay, {
        start: "2022-01-01",
        end: "2022-03-31",
      })
    ).toEqual({
      name: TimeRangeName.Last2Weeks,
      start: "2022-03-17T00:00:00.000Z",
      end: "2022-03-30T23:59:59.000Z",
      interval: "1 day",
    });
  });
});
