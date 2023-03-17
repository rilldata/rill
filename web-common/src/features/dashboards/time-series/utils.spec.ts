import { niceMeasureExtents } from "./utils";

describe("niceMeasureExtents", () => {
  it("should return [0, 1] if both values are 0", () => {
    expect(niceMeasureExtents([0, 0], 1)).toEqual([0, 1]);
  });
  it("should ensure that if the minimum is positive, it returns 0", () => {
    expect(niceMeasureExtents([1, 2], 1)).toEqual([0, 2]);
  });
  it("should ensure that if the maximum is negative, it returns 0", () => {
    expect(niceMeasureExtents([-5, -2], 1)).toEqual([-5, 0]);
  });
  it("should inflate the minimum if it is negative", () => {
    expect(niceMeasureExtents([-5, 0], 2)).toEqual([-10, 0]);
  });
  it("should inflate the maximum if it is positive", () => {
    expect(niceMeasureExtents([0, 5], 2)).toEqual([0, 10]);
  });
});
