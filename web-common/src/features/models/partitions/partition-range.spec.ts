import { describe, it, expect } from "vitest";
import type { V1ModelPartition } from "../../../runtime-client";
import {
  filterPartitionsInRange,
  searchPartitions,
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
  it("uses the observed executor limit when available", () => {
    expect(
      getPartitionExecutionSettings(
        { partitionsConcurrencyLimit: 1 },
        { partitionsConcurrency: 4 },
      ),
    ).toMatchObject({ concurrency: 4, concurrencySource: "observed" });
  });

  it("falls back to configuration before the model has run", () => {
    expect(
      getPartitionExecutionSettings({ partitionsConcurrencyLimit: 2 }, {}),
    ).toMatchObject({ concurrency: 2, concurrencySource: "configured" });
  });

  it("uses observed concurrency and leaves unavailable limits unknown", () => {
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

describe("searchPartitions", () => {
  const selected = [
    { key: "first-key", data: { hour: "0", region: "US" } },
    { key: "second-key", data: { hour: "1", region: "EU" }, error: "Timeout" },
    { key: "third-key" },
  ];

  it("searches every partition attribute without changing the selection", () => {
    expect(keys(searchPartitions(selected, " eu "))).toEqual(["second-key"]);
    expect(keys(selected)).toEqual(["first-key", "second-key", "third-key"]);
    expect(keys(searchPartitions(selected, ""))).toEqual(keys(selected));
  });

  it("matches keys and errors without case sensitivity", () => {
    expect(keys(searchPartitions(selected, "FIRST-KEY"))).toEqual([
      "first-key",
    ]);
    expect(keys(searchPartitions(selected, "timeout"))).toEqual(["second-key"]);
  });

  it("handles numeric values and empty data", () => {
    expect(
      keys(searchPartitions([p("numeric", { hour: 12 }), {}], "12")),
    ).toEqual(["numeric"]);
    expect(searchPartitions(selected, "no match")).toEqual([]);
    expect(keys(searchPartitions(selected, "   "))).toEqual(keys(selected));
  });

  it("does not reintroduce partitions removed from the selection", () => {
    const remaining = selected.filter(
      (partition) => partition.key !== "second-key",
    );
    expect(searchPartitions(remaining, "EU")).toEqual([]);
    expect(keys(searchPartitions(remaining, ""))).toEqual([
      "first-key",
      "third-key",
    ]);
  });
});

describe("partition range edge cases", () => {
  it("does not include text beyond a string end bound", () => {
    const rows = [
      p("a", { label: "a" }),
      p("b", { label: "b" }),
      p("b-child", { label: "b-child" }),
    ];
    expect(keys(filterPartitionsInRange(rows, "label", "a", "b"))).toEqual([
      "a",
      "b",
    ]);
    expect(validatePartitionRange("label", "b-child", "b")).toBe(
      "start_after_end",
    );
  });

  it("only extends a date end bound to timestamps on that date", () => {
    const rows = [
      p("timestamp", { day: "2026-03-08T23:59:59Z" }),
      p("text", { day: "2026-03-08-extra" }),
      p("invalid-timestamp", { day: "2026-03-08T-not-a-timestamp" }),
    ];
    expect(
      keys(filterPartitionsInRange(rows, "day", "2026-03-08", "2026-03-08")),
    ).toEqual(["timestamp"]);
  });

  it("offers only fields with range-comparable values", () => {
    expect(
      getPartitionFields([
        p("a", {
          day: "2026-03-08",
          count: 1,
          missing: null,
          nested: { day: "x" },
          list: [1],
        }),
      ]),
    ).toEqual(["count", "day"]);
  });

  it("shows the executor's resolved concurrency over the configured limit", () => {
    expect(
      getPartitionExecutionSettings(
        { partitionsConcurrencyLimit: 1 },
        { partitionsConcurrency: 8 },
      ),
    ).toMatchObject({ concurrency: 8, concurrencySource: "observed" });
  });
});
