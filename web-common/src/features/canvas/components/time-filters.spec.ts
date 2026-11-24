import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  normalizeTimeFilters,
  resolveTimeFilters,
  timeFiltersToParams,
} from "./time-filters";

describe("resolveTimeFilters", () => {
  it.each([
    [undefined, false, { mode: "inherit" }, undefined],
    ["", false, { mode: "inherit" }, undefined],
    ["grain=day", false, { mode: "inherit" }, V1TimeGrain.TIME_GRAIN_DAY],
    ["tr=inherit", false, { mode: "none" }, undefined],
    [
      "tr=inherit&compare_tr=rill-PP",
      false,
      { mode: "local", range: "rill-PP" },
      undefined,
    ],
    [
      "compare_tr=rill-PP",
      false,
      { mode: "local", range: "rill-PP" },
      undefined,
    ],
    ["tr=P7D", true, { mode: "none" }, undefined],
    ["tr=P7D&compare_tr=inherit", true, { mode: "inherit" }, undefined],
    [
      "tr=P7D&compare_tr=rill-PW",
      true,
      { mode: "local", range: "rill-PW" },
      undefined,
    ],
    ["tr=P7D&grain=hour", true, { mode: "none" }, V1TimeGrain.TIME_GRAIN_HOUR],
    ["grain=fortnight", false, { mode: "inherit" }, undefined],
  ])("%s", (timeFilters, hasLocalTimeRange, comparison, grain) => {
    expect(resolveTimeFilters(timeFilters)).toEqual({
      hasLocalTimeRange,
      comparison,
      grain,
    });
  });
});

describe("normalizeTimeFilters", () => {
  it.each([
    ["", undefined],
    ["tr=inherit&compare_tr=inherit", undefined],
    ["compare_tr=inherit", undefined],
    ["compare_tr=rill-PP", "tr=inherit&compare_tr=rill-PP"],
    ["tr=inherit&grain=day&tz=UTC", "tr=inherit&grain=day"],
    ["grain=week", "grain=week"],
    ["tr=inherit&compare_tr=inherit&grain=week", "grain=week"],
    ["grain=week&tr=P7D", "tr=P7D&grain=week"],
    ["tr=P7D&grain=day", "tr=P7D&grain=day"],
    ["tr=P7D&compare_tr=inherit", "tr=P7D&compare_tr=inherit"],
  ])("%s", (input, expected) => {
    expect(normalizeTimeFilters(new URLSearchParams(input))).toBe(expected);
  });
});

describe("timeFiltersToParams", () => {
  it.each([
    [undefined, "compare_tr=inherit&tr=inherit"],
    ["", "compare_tr=inherit&tr=inherit"],
    ["grain=week", "grain=week&compare_tr=inherit&tr=inherit"],
    ["compare_tr=rill-PW", "compare_tr=rill-PW&tr=inherit"],
    ["tr=inherit", "tr=inherit"],
    ["tr=P7D&compare_tr=inherit", "tr=P7D&compare_tr=inherit"],
  ])("%s", (input, expected) => {
    expect(timeFiltersToParams(input).toString()).toBe(expected);
  });
});
