import { ordinal } from "@rilldata/web-common/lib/ordinal";
import { describe, it, expect } from "vitest";

describe("ordinal", () => {
  const TestCases: [number, string][] = [
    [0, "0th"],
    [1, "1st"],
    [2, "2nd"],
    [3, "3rd"],
    [4, "4th"],
    [10, "10th"],
    [11, "11th"],
    [12, "12th"],
    [13, "13th"],
    [20, "20th"],
    [21, "21st"],
    [22, "22nd"],
    [23, "23rd"],
    [110, "110th"],
    [111, "111th"],
    [112, "112th"],
    [113, "113th"],
    [120, "120th"],
    [121, "121st"],
    [122, "122nd"],
    [123, "123rd"],
  ];

  for (const [num, expected] of TestCases) {
    it(`ordinal(${num})=${expected}`, () => {
      expect(ordinal(num)).toEqual(expected);
    });
  }
});
