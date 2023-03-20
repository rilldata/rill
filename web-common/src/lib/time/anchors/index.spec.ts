import { Period, TimeOffsetType } from "../types";
import { getEndOfPeriod, getOffset, getStartOfPeriod, getTimeWidth } from "./";
import { durationToMillis } from "./grains";

describe("getStartOfPeriod", () => {
  it("should return the start of the week for given date", () => {
    const timeGrain = getStartOfPeriod(Period.WEEK, new Date("2020-03-15"));
    expect(timeGrain).toEqual(new Date("2020-03-09"));
  });
  it("should return the start of month for given date", () => {
    const timeGrain = getStartOfPeriod(Period.MONTH, new Date("2020-03-15"));
    expect(timeGrain).toEqual(new Date("2020-03-01"));
  });
});

describe("getEndOfPeriod", () => {
  it("should return the end of the week for given date", () => {
    const timeGrain = getEndOfPeriod(Period.WEEK, new Date("2020-03-15"));
    expect(timeGrain).toEqual(new Date("2020-03-15T23:59:59.999Z"));
  });
  it("should return the end of month for given date", () => {
    const timeGrain = getEndOfPeriod(Period.MONTH, new Date("2020-02-15"));
    // leap year!
    expect(timeGrain).toEqual(new Date("2020-02-29T23:59:59.999Z"));
  });
});

describe("getOffset", () => {
  it("should add correct amount of time for given date", () => {
    const timeGrain = getOffset(
      new Date("2020-02-15"),
      "P2W",
      TimeOffsetType.ADD
    );
    expect(timeGrain).toEqual(new Date("2020-02-29"));
  });
  it("should subtract correct amount of time for given date", () => {
    const timeGrain = getOffset(
      new Date("2020-02-15"),
      "P2M",
      TimeOffsetType.SUBTRACT
    );
    expect(timeGrain).toEqual(new Date("2019-12-15"));
  });
});

describe("getTimeWidth", () => {
  it("should give correct amount of time width in milliseconds for given dates", () => {
    const timeGrain = getTimeWidth(
      new Date("2020-03-15"),
      new Date("2020-04-01")
    );
    expect(timeGrain).toEqual(durationToMillis("P1D") * 17);
  });
});
