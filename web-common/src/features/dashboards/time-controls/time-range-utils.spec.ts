import { TimeGrain } from "./time-control-types";
import {
  getDefaultTimeGrain,
  getSelectableTimeGrains,
} from "./time-range-utils";

describe("getSelectableTimeGrains", () => {
  it("should return an array of available time grains", () => {
    const timeGrains = getSelectableTimeGrains(
      new Date("2020-03-01"),
      new Date("2020-03-31")
    );
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
    const timeGrain = getDefaultTimeGrain(
      new Date("2020-03-01"),
      new Date("2020-03-31")
    );
    expect(timeGrain).toEqual(TimeGrain.OneDay);
  });
  it("should return the default time grain (for an AllTime time range", () => {
    const timeGrain = getDefaultTimeGrain(
      new Date("2010-03-01"),
      new Date("2020-03-31")
    );
    expect(timeGrain).toEqual(TimeGrain.OneMonth);
  });
  it("should return the default time grain (for an AllTime time range", () => {
    const timeGrain = getDefaultTimeGrain(
      new Date("2010-03-01"),
      new Date("2030-03-31")
    );
    expect(timeGrain).toEqual(TimeGrain.OneYear);
  });
});
