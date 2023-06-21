import { V1TimeGrain } from "../../../runtime-client";
import { getDefaultTimeGrain } from "./time-range-utils";

describe("getDefaultTimeGrain", () => {
  it("should return the default time grain (for a 5 day time range)", () => {
    const timeGrain = getDefaultTimeGrain(
      new Date("2020-03-01"),
      new Date("2020-03-05")
    );
    expect(timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_HOUR);
  });
  it("should return the default time grain (for a 30 day time range)", () => {
    const timeGrain = getDefaultTimeGrain(
      new Date("2020-03-01"),
      new Date("2020-03-31")
    );
    expect(timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_DAY);
  });
  it("should return the default time grain (for an AllTime time range", () => {
    const timeGrain = getDefaultTimeGrain(
      new Date("2010-03-01"),
      new Date("2020-03-31")
    );
    expect(timeGrain).toEqual(V1TimeGrain.TIME_GRAIN_MONTH);
  });
});
