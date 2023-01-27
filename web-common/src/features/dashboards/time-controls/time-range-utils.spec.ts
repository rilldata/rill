import { TimeGrain, TimeRangeName } from "./time-control-types";
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
      TimeRangeName.LastHour,
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
        enabled: false,
        timeGrain: "minute",
      },
      {
        enabled: true,
        timeGrain: "hour",
      },
      {
        enabled: true,
        timeGrain: "day",
      },
      {
        enabled: true,
        timeGrain: "week",
      },
      {
        enabled: false,
        timeGrain: "month",
      },
      {
        enabled: false,
        timeGrain: "year",
      },
    ]);
  });
});

describe("getDefaultTimeGrain", () => {
  it("should return the default time grain (for a LastX time range)", () => {
    const timeGrain = getDefaultTimeGrain(TimeRangeName.Last30Days, {
      start: "2020-03-01",
      end: "2020-03-31",
    });
    expect(timeGrain).toEqual(TimeGrain.OneDay);
  });
  it("should return the default time grain (for an AllTime time range", () => {
    const timeGrain = getDefaultTimeGrain(TimeRangeName.AllTime, {
      start: "2010-03-01",
      end: "2020-03-31",
    });
    expect(timeGrain).toEqual(TimeGrain.OneMonth);
  });
  it("should return the default time grain (for an AllTime time range", () => {
    const timeGrain = getDefaultTimeGrain(TimeRangeName.AllTime, {
      start: "2010-03-01",
      end: "2030-03-31",
    });
    expect(timeGrain).toEqual(TimeGrain.OneYear);
  });
});

describe("makeTimeRange", () => {
  it("should create a TimeRange object representing the Last Two Weeks", () => {
    expect(
      makeTimeRange(TimeRangeName.Last2Weeks, TimeGrain.OneDay, {
        start: "2022-01-01T11:00:01",
        end: "2022-03-31T20:00:01",
      })
    ).toEqual({
      name: TimeRangeName.Last2Weeks,
      start: "2022-03-17T00:00:00.000Z",
      end: "2022-04-01T00:00:00.000Z",
      interval: "day",
    });
  });
});
