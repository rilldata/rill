import { describe, expect, it } from "vitest";
import {
  normalizeTimeFilters,
  resolveTimeFilters,
  stripInheritedTimeFilters,
} from "./time-filters";

describe("resolveTimeFilters", () => {
  it.each([
    [undefined, false, { mode: "inherit" }],
    ["", false, { mode: "inherit" }],
    ["grain=day", false, { mode: "inherit" }],
    ["tr=inherit", false, { mode: "none" }],
    [
      "tr=inherit&compare_tr=rill-PP",
      false,
      { mode: "local", range: "rill-PP" },
    ],
    ["compare_tr=rill-PP", false, { mode: "local", range: "rill-PP" }],
    ["tr=P7D", true, { mode: "none" }],
    ["tr=P7D&compare_tr=inherit", true, { mode: "inherit" }],
    ["tr=P7D&compare_tr=rill-PW", true, { mode: "local", range: "rill-PW" }],
  ])("%s", (timeFilters, hasLocalTimeRange, comparison) => {
    expect(resolveTimeFilters(timeFilters)).toEqual({
      hasLocalTimeRange,
      comparison,
    });
  });
});

describe("normalizeTimeFilters", () => {
  it.each([
    ["", undefined],
    ["tr=inherit&compare_tr=inherit", undefined],
    ["compare_tr=inherit", undefined],
    ["compare_tr=rill-PP", "tr=inherit&compare_tr=rill-PP"],
    ["tr=inherit&grain=day&tz=UTC", "tr=inherit"],
    ["tr=P7D&grain=day", "tr=P7D&grain=day"],
    ["tr=P7D&compare_tr=inherit", "tr=P7D&compare_tr=inherit"],
  ])("%s", (input, expected) => {
    expect(normalizeTimeFilters(new URLSearchParams(input))).toBe(expected);
  });
});

describe("stripInheritedTimeFilters", () => {
  it("drops the inherit sentinels and keeps real values", () => {
    expect(
      stripInheritedTimeFilters(
        new URLSearchParams("tr=inherit&compare_tr=rill-PP"),
      ).toString(),
    ).toBe("compare_tr=rill-PP");
    expect(
      stripInheritedTimeFilters(
        new URLSearchParams("tr=P7D&compare_tr=inherit&grain=day"),
      ).toString(),
    ).toBe("tr=P7D&grain=day");
  });
});
