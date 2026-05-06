import { Duration } from "luxon";
import { describe, expect, it } from "vitest";
import { bucketYamlRanges } from "./new-time-controls";

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
    const cap = Duration.fromISO("P30D");
    const buckets = bucketYamlRanges(yaml, undefined, false, cap);
    expect(buckets.latest.map((p) => p.toString())).toEqual(["P7D", "P30D"]);
    expect(buckets.allTime).toBe(false);
  });

  it("filters the default RILL_LATEST set when no yamlRanges are provided", () => {
    const cap = Duration.fromISO("P14D");
    const buckets = bucketYamlRanges([], undefined, false, cap);
    const labels = buckets.latest.map((p) => p.toString());
    // P12M and P4W exceed 14 days; PT24H, P7D, PT6H, P14D fit.
    expect(labels).not.toContain("P12M");
    expect(labels).not.toContain("P4W");
    expect(buckets.allTime).toBe(false);
  });

  it("treats a zero-duration cap as uncapped", () => {
    const cap = Duration.fromMillis(0);
    const buckets = bucketYamlRanges(yaml, undefined, false, cap);
    expect(buckets.allTime).toBe(true);
    expect(buckets.latest.length).toBe(3);
  });
});
