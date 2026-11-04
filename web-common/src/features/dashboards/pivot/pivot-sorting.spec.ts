import {
  canSortPivotBy,
  pivotHasSortField,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-sorting";
import {
  type PivotChipData,
  PivotChipType,
  type PivotTableMode,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const chip = (type: PivotChipType, id: string): PivotChipData => ({
  id,
  title: id,
  type,
});

describe("pivotHasSortField", () => {
  const pivot = (tableMode: PivotTableMode) => ({
    tableMode,
    rows: [chip(PivotChipType.Dimension, "publisher")],
    columns: [
      chip(PivotChipType.Dimension, "domain"),
      chip(PivotChipType.Time, V1TimeGrain.TIME_GRAIN_DAY),
      chip(PivotChipType.Measure, "impressions"),
    ],
  });

  it("accepts a dimension or measure chip in rows or columns", () => {
    expect(pivotHasSortField("publisher", pivot("nest"))).toBe(true);
    expect(pivotHasSortField("domain", pivot("flat"))).toBe(true);
    expect(pivotHasSortField("impressions", pivot("nest"))).toBe(true);
  });

  it("accepts a measure's comparison columns", () => {
    expect(pivotHasSortField("impressions__delta_abs", pivot("nest"))).toBe(
      true,
    );
    expect(pivotHasSortField("impressions__delta_rel", pivot("flat"))).toBe(
      true,
    );
  });

  it("accepts a time column by its grain", () => {
    expect(
      pivotHasSortField("timestamp_rill_TIME_GRAIN_DAY", pivot("nest")),
    ).toBe(true);
    expect(pivotHasSortField(V1TimeGrain.TIME_GRAIN_DAY, pivot("nest"))).toBe(
      true,
    );
    expect(
      pivotHasSortField("timestamp_rill_TIME_GRAIN_WEEK", pivot("nest")),
    ).toBe(false);
  });

  it("accepts a measure under one column value only in a nested table", () => {
    expect(pivotHasSortField("c0v2m0", pivot("nest"))).toBe(true);
    expect(pivotHasSortField("c0v2_c1v0m1", pivot("nest"))).toBe(true);
    expect(pivotHasSortField("m0", pivot("nest"))).toBe(true);
    expect(pivotHasSortField("c0v2m0", pivot("flat"))).toBe(false);
  });

  it("rejects a field that is not in the pivot", () => {
    expect(pivotHasSortField("bid_price", pivot("nest"))).toBe(false);
    expect(pivotHasSortField("timestamp", pivot("flat"))).toBe(false);
    expect(pivotHasSortField("impressions__previous", pivot("nest"))).toBe(
      false,
    );
    expect(pivotHasSortField("c0v2m0x", pivot("nest"))).toBe(false);
  });
});

describe("canSortPivotBy", () => {
  it("accepts any dimension or measure column of a flat table", () => {
    const config = {
      isFlat: true,
      rowDimensionNames: ["domain", "timestamp_rill_TIME_GRAIN_DAY"],
      colDimensionNames: [],
      measureNames: ["impressions", "impressions__delta_abs"],
    };
    expect(canSortPivotBy("domain", config)).toBe(true);
    expect(canSortPivotBy("timestamp_rill_TIME_GRAIN_DAY", config)).toBe(true);
    expect(canSortPivotBy("impressions__delta_abs", config)).toBe(true);
    expect(canSortPivotBy("timestamp", config)).toBe(false);
    expect(canSortPivotBy("bid_price", config)).toBe(false);
  });

  it("accepts the row axis, a measure or a measure under one column value of a nested table", () => {
    const config = {
      isFlat: false,
      rowDimensionNames: ["publisher", "domain"],
      colDimensionNames: ["country", "timestamp_rill_TIME_GRAIN_DAY"],
      measureNames: ["impressions", "bid_price"],
    };
    expect(canSortPivotBy("publisher", config)).toBe(true);
    expect(canSortPivotBy("bid_price", config)).toBe(true);
    expect(canSortPivotBy("c0v2_c1v0m1", config)).toBe(true);
    expect(canSortPivotBy("c1v3m0", config)).toBe(true);
    // Only the first row dimension keys the row axis column.
    expect(canSortPivotBy("domain", config)).toBe(false);
    // Column dimensions are group headers, which cannot be sorted.
    expect(canSortPivotBy("country", config)).toBe(false);
    // Beyond the pivot's column dimensions or measures.
    expect(canSortPivotBy("c2v0m0", config)).toBe(false);
    expect(canSortPivotBy("c0v0m2", config)).toBe(false);
    // Comparison columns are in measureNames only while comparing.
    expect(canSortPivotBy("impressions__delta_abs", config)).toBe(false);
  });

  it("accepts the measure columns of a nested table without column dimensions", () => {
    const config = {
      isFlat: false,
      rowDimensionNames: ["publisher"],
      colDimensionNames: [],
      measureNames: ["impressions"],
    };
    expect(canSortPivotBy("m0", config)).toBe(true);
    expect(canSortPivotBy("m1", config)).toBe(false);
  });
});
