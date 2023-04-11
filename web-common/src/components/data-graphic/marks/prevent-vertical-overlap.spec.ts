import { preventVerticalOverlap } from "./prevent-vertical-overlap";

describe("preventVerticalOverlap", () => {
  test("returns an empty array if input is empty", () => {
    const result = preventVerticalOverlap([], 0, 100, 10, 2);
    expect(result).toEqual([]);
  });

  test("returns the input array if only one point is provided", () => {
    const input = [{ key: 1, value: 50 }];
    const expectedOutput = [{ key: 1, value: 50 }];
    const result = preventVerticalOverlap(input, 0, 100, 10, 2);
    expect(result).toEqual(expectedOutput);
  });

  test("prevents overlap for points close together", () => {
    const input = [
      { key: 1, value: 50 },
      { key: 2, value: 55 },
    ];
    const expectedOutput = [
      { key: 1, value: 43 },
      { key: 2, value: 55 },
    ];
    const result = preventVerticalOverlap(input, 0, 100, 10, 2);
    expect(result).toEqual(expectedOutput);
  });

  test("prevents overlap for points and respects boundaries", () => {
    const input = [
      { key: 1, value: 10 },
      { key: 2, value: 25 },
      { key: 3, value: 60 },
      { key: 4, value: 90 },
    ];
    const expectedOutput = [
      { key: 1, value: 12 },
      { key: 2, value: 25 },
      { key: 3, value: 60 },
      { key: 4, value: 88 },
    ];
    const result = preventVerticalOverlap(input, 10, 90, 10, 2);
    expect(result).toEqual(expectedOutput);
  });

  test("handles case when all points are close to the top boundary", () => {
    const input = [
      { key: 1, value: 15 },
      { key: 2, value: 20 },
      { key: 3, value: 25 },
    ];
    const expectedOutput = [
      { key: 1, value: 12 },
      { key: 2, value: 24 },
      { key: 3, value: 36 },
    ];
    const result = preventVerticalOverlap(input, 10, 90, 10, 2);
    expect(result).toEqual(expectedOutput);
  });
  test("handles case when all points are close to the bottom boundary", () => {
    const input = [
      { key: 1, value: 75 },
      { key: 2, value: 80 },
      { key: 3, value: 85 },
    ];
    const expectedOutput = [
      { key: 1, value: 64 },
      { key: 2, value: 76 },
      { key: 3, value: 88 },
    ];
    const result = preventVerticalOverlap(input, 10, 90, 10, 2);
    expect(result).toEqual(expectedOutput);
  });

  test("handles points close together in the middle", () => {
    const input = [
      { key: 1, value: 45 },
      { key: 2, value: 50 },
      { key: 3, value: 55 },
    ];
    const expectedOutput = [
      { key: 1, value: 38 },
      { key: 2, value: 50 },
      { key: 3, value: 62 },
    ];
    const result = preventVerticalOverlap(input, 10, 90, 10, 2);
    expect(result).toEqual(expectedOutput);
  });

  test("handles points near both boundaries", () => {
    const input = [
      { key: 1, value: 15 },
      { key: 2, value: 85 },
    ];
    const expectedOutput = [
      { key: 1, value: 15 },
      { key: 2, value: 85 },
    ];
    const result = preventVerticalOverlap(input, 10, 90, 10, 2);
    expect(result).toEqual(expectedOutput);
  });

  test("handles a large number of points close together", () => {
    const input = Array.from({ length: 10 }, (_, i) => ({
      key: i,
      value: i * 5 + 10,
    }));
    const result = preventVerticalOverlap(input, 0, 100, 10, 2);

    for (let i = 1; i < result.length; i++) {
      const difference = result[i].value - result[i - 1].value;
      expect(difference).toBeGreaterThanOrEqual(12);
    }
  });
});
