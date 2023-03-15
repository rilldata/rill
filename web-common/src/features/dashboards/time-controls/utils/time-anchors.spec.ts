import {
  getEndOfPeriod,
  getOffset,
  getStartOfPeriod,
  getTimeWidth,
  Period,
} from "./time-anchors";

describe("getStartOfPeriod", () => {
  it("should return the start of the week for given date", () => {
    const timeGrain = getStartOfPeriod(Period.WEEK, new Date("2020-03-15"));
    expect(timeGrain).toEqual(new Date("2020-03-12"));
  });
  it("should return the start of month for given date", () => {
    const timeGrain = getStartOfPeriod(Period.MONTH, new Date("2020-03-15"));
    expect(timeGrain).toEqual(new Date("2020-03-01"));
  });
});

describe("getEndOfPeriod", () => {
  it("should return the end of the week for given date", () => {
    const timeGrain = getEndOfPeriod(Period.WEEK, new Date("2020-03-15"));
    expect(timeGrain).toEqual(new Date("2020-03-18"));
  });
  it("should return the end of month for given date", () => {
    const timeGrain = getEndOfPeriod(Period.MONTH, new Date("2020-02-15"));
    expect(timeGrain).toEqual(new Date("2020-02-28"));
  });
});

describe("getOffset", () => {
  it("should add correct amount of time for given date", () => {
    const timeGrain = getOffset(new Date("2020-02-15"), Period.WEEK, 2);
    expect(timeGrain).toEqual(new Date("2020-03-01"));
  });
  it("should subtract correct amount of time for given date", () => {
    const timeGrain = getOffset(new Date("2020-02-15"), Period.MONTH, -2);
    expect(timeGrain).toEqual(new Date("2019-12-15"));
  });
});

describe("getTimeWidth", () => {
  it("should give correct amount of time width in milliseconds for given dates", () => {
    const timeGrain = getTimeWidth(
      new Date("2020-03-15"),
      new Date("2020-04-01")
    );
    expect(timeGrain).toEqual(1000 * 60 * 60 * 24 * 17);
  });
});
