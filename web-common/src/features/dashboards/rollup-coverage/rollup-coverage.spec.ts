import { describe, expect, it } from "vitest";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  BASE_TABLE_KEY,
  computeCoverageWarning,
  coverageFromTimeRangeResponse,
  rollupGrainsFromSpec,
  type TableCoverage,
} from "./rollup-coverage";

const ms = (iso: string) => Date.parse(iso);

// Base table (BIG): 4 months of retention. Rollup (SMALL): 2 years.
const coverage = new Map<string, TableCoverage>([
  [
    BASE_TABLE_KEY,
    { minMs: ms("2024-03-01T00:00:00Z"), maxMs: ms("2024-07-01T00:00:00Z") },
  ],
  [
    "rollup_small",
    { minMs: ms("2022-07-01T00:00:00Z"), maxMs: ms("2024-07-01T00:00:00Z") },
  ],
]);

const grains = new Map([["rollup_small", V1TimeGrain.TIME_GRAIN_DAY]]);

describe("computeCoverageWarning", () => {
  it("warns when a base-served widget is narrower than a rollup in use", () => {
    // Requested: 1 year. Timeseries serves from rollup (2y), this toplist from base (4M).
    const warning = computeCoverageWarning(
      BASE_TABLE_KEY,
      coverage,
      [BASE_TABLE_KEY, "rollup_small"],
      grains,
      ms("2023-07-01T00:00:00Z"),
    );
    expect(warning).toEqual({
      servedStartMs: ms("2024-03-01T00:00:00Z"),
      availableStartMs: ms("2023-07-01T00:00:00Z"),
      servingTable: BASE_TABLE_KEY,
    });
  });

  it("does not warn when all widgets use the same table", () => {
    // Uniformly-truncated dashboard (e.g. a filter routed everything to base): consistent, no warning.
    const warning = computeCoverageWarning(
      BASE_TABLE_KEY,
      coverage,
      [BASE_TABLE_KEY],
      grains,
      ms("2023-07-01T00:00:00Z"),
    );
    expect(warning).toBeUndefined();
  });

  it("does not warn for the widget served from the widest table", () => {
    const warning = computeCoverageWarning(
      "rollup_small",
      coverage,
      [BASE_TABLE_KEY, "rollup_small"],
      grains,
      ms("2023-07-01T00:00:00Z"),
    );
    expect(warning).toBeUndefined();
  });

  it("does not warn when the requested range is within the narrow table's coverage", () => {
    // Requested: 1 month, inside base coverage. Tables differ but nothing is missing.
    const warning = computeCoverageWarning(
      BASE_TABLE_KEY,
      coverage,
      [BASE_TABLE_KEY, "rollup_small"],
      grains,
      ms("2024-06-01T00:00:00Z"),
    );
    expect(warning).toBeUndefined();
  });

  it("clamps the available start to the requested range", () => {
    // Requested start predates even the rollup's coverage: report from rollup min, not request start.
    const warning = computeCoverageWarning(
      BASE_TABLE_KEY,
      coverage,
      [BASE_TABLE_KEY, "rollup_small"],
      grains,
      ms("2020-01-01T00:00:00Z"),
    );
    expect(warning?.availableStartMs).toBe(ms("2022-07-01T00:00:00Z"));
  });

  it("suppresses sub-bucket differences as grain-truncation artifacts", () => {
    // Monthly rollup built from base data starting Jan 15 reports min = Jan 1.
    // The 14-day difference is truncation, not extra data.
    const truncated = new Map<string, TableCoverage>([
      [
        BASE_TABLE_KEY,
        {
          minMs: ms("2024-01-15T00:00:00Z"),
          maxMs: ms("2024-07-01T00:00:00Z"),
        },
      ],
      [
        "rollup_month",
        {
          minMs: ms("2024-01-01T00:00:00Z"),
          maxMs: ms("2024-07-01T00:00:00Z"),
        },
      ],
    ]);
    const warning = computeCoverageWarning(
      BASE_TABLE_KEY,
      truncated,
      [BASE_TABLE_KEY, "rollup_month"],
      new Map([["rollup_month", V1TimeGrain.TIME_GRAIN_MONTH]]),
      ms("2024-01-01T00:00:00Z"),
    );
    expect(warning).toBeUndefined();
  });

  it("returns undefined without coverage data", () => {
    const warning = computeCoverageWarning(
      BASE_TABLE_KEY,
      undefined,
      [BASE_TABLE_KEY],
      grains,
      ms("2023-07-01T00:00:00Z"),
    );
    expect(warning).toBeUndefined();
  });
});

describe("coverageFromTimeRangeResponse", () => {
  it("returns undefined when the metrics view has no rollups", () => {
    expect(
      coverageFromTimeRangeResponse({
        timeRangeSummary: {
          min: "2024-01-01T00:00:00Z",
          max: "2024-07-01T00:00:00Z",
        },
      }),
    ).toBeUndefined();
  });

  it("maps base and rollup coverage", () => {
    const result = coverageFromTimeRangeResponse({
      timeRangeSummary: {
        min: "2024-03-01T00:00:00Z",
        max: "2024-07-01T00:00:00Z",
      },
      rollupTimeRanges: {
        rollup_small: {
          min: "2022-07-01T00:00:00Z",
          max: "2024-07-01T00:00:00Z",
        },
      },
    });
    expect(result?.get(BASE_TABLE_KEY)).toEqual({
      minMs: ms("2024-03-01T00:00:00Z"),
      maxMs: ms("2024-07-01T00:00:00Z"),
    });
    expect(result?.get("rollup_small")).toEqual({
      minMs: ms("2022-07-01T00:00:00Z"),
      maxMs: ms("2024-07-01T00:00:00Z"),
    });
  });
});

describe("rollupGrainsFromSpec", () => {
  it("maps rollup tables to grains", () => {
    const result = rollupGrainsFromSpec([
      { table: "rollup_small", timeGrain: V1TimeGrain.TIME_GRAIN_DAY },
      { model: "unresolved" },
    ]);
    expect(result.get("rollup_small")).toBe(V1TimeGrain.TIME_GRAIN_DAY);
    expect(result.size).toBe(1);
  });
});
