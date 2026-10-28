import { describe, it, expect } from "vitest";
import type { V1ModelPartition } from "../../../runtime-client";
import {
  filterPartitionsInRange,
  getPartitionExecutionSettings,
  getPartitionFields,
  isPartitionPending,
  validatePartitionRange,
} from "./partition-range";

function p(key: string, data: Record<string, unknown>): V1ModelPartition {
  return { key, data };
}

const keys = (ps: V1ModelPartition[]) => ps.map((x) => x.key);

describe("getPartitionFields", () => {
  it("returns the sorted union of data keys", () => {
    expect(
      getPartitionFields([
        p("a", { day: "2026-03-01" }),
        p("b", { uri: "s3://x", day: "2026-03-02" }),
        { key: "c" },
      ]),
    ).toEqual(["day", "uri"]);
  });
});

describe("filterPartitionsInRange", () => {
  const days = [
    p("d1", { day: "2026-03-01" }),
    p("d2", { day: "2026-03-02" }),
    p("d3", { day: "2026-03-08T00:00:00Z" }),
    p("d4", { day: "2026-03-09T00:00:00Z" }),
  ];

  it("includes both bounds for ISO strings", () => {
    expect(
      keys(filterPartitionsInRange(days, "day", "2026-03-02", "2026-03-08")),
    ).toEqual(["d2", "d3"]);
  });

  it("treats an end bound that prefixes a timestamp as inclusive", () => {
    expect(
      keys(filterPartitionsInRange(days, "day", "2026-03-08", "2026-03-08")),
    ).toEqual(["d3"]);
  });

  it("compares numbers numerically", () => {
    const nums = [
      p("n2", { num: 2 }),
      p("n10", { num: 10 }),
      p("n30", { num: 30 }),
    ];
    expect(keys(filterPartitionsInRange(nums, "num", "2", "10"))).toEqual([
      "n2",
      "n10",
    ]);
  });

  it("does not prefix-match numbers past the end bound", () => {
    const nums = [1, 2, 10, 20, 100, 109, 200].map((n) =>
      p(`n${n}`, { num: n }),
    );
    expect(keys(filterPartitionsInRange(nums, "num", "1", "10"))).toEqual([
      "n1",
      "n2",
      "n10",
    ]);
    expect(keys(filterPartitionsInRange(nums, "num", "2", "2"))).toEqual([
      "n2",
    ]);
  });

  it("compares numeric strings numerically", () => {
    const hours = ["0", "9", "10", "12", "13", "23"].map((h) =>
      p(`h${h}`, { hour: h }),
    );
    expect(keys(filterPartitionsInRange(hours, "hour", "9", "12"))).toEqual([
      "h9",
      "h10",
      "h12",
    ]);
  });

  it("excludes partitions missing the field or holding non-scalar values", () => {
    const mixed = [
      p("m1", { day: "2026-03-02" }),
      p("m2", { other: "2026-03-02" }),
      p("m3", { day: null }),
      p("m4", { day: { nested: true } }),
    ];
    expect(
      keys(filterPartitionsInRange(mixed, "day", "2026-03-01", "2026-03-31")),
    ).toEqual(["m1"]);
  });

  it("returns nothing when no partition matches", () => {
    expect(
      filterPartitionsInRange(days, "day", "2027-01-01", "2027-01-31"),
    ).toEqual([]);
  });
});

describe("validatePartitionRange", () => {
  it("requires field, start, and end", () => {
    expect(validatePartitionRange("", "a", "b")).toBe("missing");
    expect(validatePartitionRange("day", "", "b")).toBe("missing");
    expect(validatePartitionRange("day", "a", "")).toBe("missing");
  });

  it("rejects a start after the end", () => {
    expect(validatePartitionRange("day", "2026-03-09", "2026-03-08")).toBe(
      "start_after_end",
    );
    expect(validatePartitionRange("num", "10", "9")).toBe("start_after_end");
  });

  it("accepts valid ranges", () => {
    expect(
      validatePartitionRange("day", "2026-03-08", "2026-03-08"),
    ).toBeUndefined();
    expect(validatePartitionRange("num", "9", "10")).toBeUndefined();
  });
});

describe("isPartitionPending", () => {
  it("is pending only when not executed and not skipped", () => {
    expect(isPartitionPending({ key: "a" })).toBe(true);
    expect(
      isPartitionPending({ key: "a", executedOn: "2026-03-01T00:00:00Z" }),
    ).toBe(false);
    expect(isPartitionPending({ key: "a", skipped: true })).toBe(false);
  });
});

describe("getPartitionExecutionSettings", () => {
  it("prefers the configured limit over a previously observed value", () => {
    expect(
      getPartitionExecutionSettings(
        { partitionsConcurrencyLimit: 1 },
        { partitionsConcurrency: 4 },
      ),
    ).toMatchObject({ concurrency: 1, concurrencySource: "configured" });
  });

  it("falls back to the observed value, then unknown", () => {
    expect(
      getPartitionExecutionSettings({}, { partitionsConcurrency: 4 }),
    ).toMatchObject({ concurrency: 4, concurrencySource: "observed" });
    expect(getPartitionExecutionSettings({}, {})).toMatchObject({
      concurrency: undefined,
      concurrencySource: "unknown",
    });
  });

  it("applies the runtime retry defaults when unset", () => {
    const s = getPartitionExecutionSettings({}, {});
    expect(s).toMatchObject({
      retryAttempts: 3,
      retryDelaySeconds: 5,
      retryExponentialBackoff: true,
      retryIsDefault: true,
    });
    expect(s.retryIfErrorMatches).toHaveLength(4);
  });

  it("uses configured retry options", () => {
    const s = getPartitionExecutionSettings(
      {
        retryAttempts: 0,
        retryDelaySeconds: 30,
        retryExponentialBackoff: false,
        retryIfErrorMatches: [".*quota.*"],
      },
      {},
    );
    expect(s).toEqual({
      concurrency: undefined,
      concurrencySource: "unknown",
      retryAttempts: 1,
      retryDelaySeconds: 30,
      retryExponentialBackoff: false,
      retryIfErrorMatches: [".*quota.*"],
      retryIsDefault: false,
    });
  });
});
