import { describe, expect, it } from "vitest";
import { bucketYamlRanges, isoDurationToDays } from "./new-time-controls";

describe("isoDurationToDays", () => {
  it("parses days", () => {
    expect(isoDurationToDays("P7D")).toBe(7);
    expect(isoDurationToDays("P90D")).toBe(90);
  });

  it("approximates months and years to match executor.EstimateNative()", () => {
    expect(isoDurationToDays("P1M")).toBe(30);
    expect(isoDurationToDays("P3M")).toBe(90);
    expect(isoDurationToDays("P1Y")).toBe(365);
  });

  it("handles weeks and combinations", () => {
    expect(isoDurationToDays("P2W")).toBe(14);
    expect(isoDurationToDays("P1M7D")).toBe(37);
  });

  it("converts sub-day units to fractional days", () => {
    expect(isoDurationToDays("PT12H")).toBe(0.5);
    expect(isoDurationToDays("PT24H")).toBe(1);
  });

  it("returns NaN for invalid input", () => {
    expect(Number.isNaN(isoDurationToDays("garbage"))).toBe(true);
    expect(Number.isNaN(isoDurationToDays(""))).toBe(true);
  });
});

describe("bucketYamlRanges with maxQueryTimeRange", () => {
  const yaml = [
    { range: "P7D" },
    { range: "P30D" },
    { range: "P12M" },
    { range: "inf" },
  ];

  it("includes everything when no cap is set", () => {
    const buckets = bucketYamlRanges(yaml, undefined, false);
    expect(buckets.latest.map((p) => p.toString())).toEqual([
      "P7D",
      "P30D",
      "P12M",
    ]);
    expect(buckets.allTime).toBe(true);
  });

  it("drops presets that exceed the cap and disables All Time", () => {
    const buckets = bucketYamlRanges(yaml, undefined, false, "P30D");
    expect(buckets.latest.map((p) => p.toString())).toEqual(["P7D", "P30D"]);
    expect(buckets.allTime).toBe(false);
  });

  it("filters the default RILL_LATEST set when no yamlRanges are provided", () => {
    const buckets = bucketYamlRanges([], undefined, false, "P14D");
    const labels = buckets.latest.map((p) => p.toString());
    // P12M and P4W exceed 14 days; PT24H, P7D, PT6H, P14D fit.
    expect(labels).not.toContain("P12M");
    expect(labels).not.toContain("P4W");
    expect(buckets.allTime).toBe(false);
  });

  it("ignores an unparseable cap and behaves as if uncapped", () => {
    const buckets = bucketYamlRanges(yaml, undefined, false, "garbage");
    expect(buckets.allTime).toBe(true);
    expect(buckets.latest.length).toBe(3);
  });
});
