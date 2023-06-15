import { roundToNearestTimeUnit } from "./round-to-nearest-time-unit";
import { describe, it, expect } from "vitest";

describe("roundToNearestTimeUnit", () => {
  it("rounds to nearest minute", () => {
    const date = new Date("2023-03-29T12:34:56");
    const expectedResult = new Date("2023-03-29T12:35:00");
    expect(roundToNearestTimeUnit(date, "minute")).toEqual(expectedResult);
  });

  it("rounds to nearest hour", () => {
    const date = new Date("2023-03-29T12:34:56");
    const expectedResult = new Date("2023-03-29T13:00:00");
    expect(roundToNearestTimeUnit(date, "hour")).toEqual(expectedResult);
  });

  it("rounds to nearest day", () => {
    const date = new Date("2023-03-29T12:34:56");
    const expectedResult = new Date("2023-03-30T00:00:00");
    expect(roundToNearestTimeUnit(date, "day")).toEqual(expectedResult);
  });

  it("rounds to nearest month", () => {
    const date = new Date("2023-03-14T12:34:56");
    const expectedResult = new Date("2023-03-01T00:00:00");
    expect(roundToNearestTimeUnit(date, "month")).toEqual(expectedResult);
  });
});
