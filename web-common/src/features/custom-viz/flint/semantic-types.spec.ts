import { describe, expect, it } from "vitest";
import { SemanticTypes } from "flint-chart/core";
import {
  MetricsViewSpecDimensionType,
  V1TimeGrain,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { deriveFlintFields, type FlintSemanticType } from "./semantic-types";

/**
 * Every semantic type this module can emit. Kept as a literal list rather than derived from the
 * FlintSemanticType union, so the registry check below sees real runtime values.
 */
const EMITTED_TYPES: FlintSemanticType[] = [
  "Year",
  "YearQuarter",
  "YearMonth",
  "YearWeek",
  "Date",
  "DateTime",
  "Amount",
  "Percentage",
  "Duration",
  "Count",
  "Number",
  "Category",
];

const metricsView: V1MetricsViewSpec = {
  timeDimension: "__time",
  measures: [
    { name: "revenue", displayName: "Revenue", formatPreset: "currency_usd" },
    { name: "ctr", displayName: "CTR", formatPreset: "percentage" },
    { name: "dwell", formatPreset: "interval_ms" },
    { name: "d3_money", formatD3: "$,.2f" },
    { name: "d3_pct", formatD3: ".1%" },
    { name: "impressions", dataType: { code: "CODE_INT64" } },
    { name: "score", dataType: { code: "CODE_FLOAT64" } },
  ],
  dimensions: [
    { name: "advertiser", displayName: "Advertiser" },
    {
      name: "event_time",
      type: MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
    },
  ],
};

describe("deriveFlintFields", () => {
  it("only emits semantic types that exist in Flint's registry", () => {
    // Flint silently degrades unregistered types to nominal/plain, and its docs advertise several
    // (Revenue, Cost, Rating, Index) that are not registered. This is the guard against that.
    const registered = new Set(Object.keys(SemanticTypes));
    for (const type of EMITTED_TYPES) {
      expect(
        registered,
        `"${type}" is not a registered Flint semantic type`,
      ).toContain(type);
    }
  });

  it("maps measures by how they are formatted", () => {
    const { semantic_types } = deriveFlintFields(
      ["revenue", "ctr", "dwell", "d3_money", "d3_pct", "impressions", "score"],
      metricsView,
    );

    expect(semantic_types).toEqual({
      revenue: "Amount",
      ctr: "Percentage",
      dwell: "Duration",
      d3_money: "Amount",
      d3_pct: "Percentage",
      impressions: "Count",
      score: "Number",
    });
  });

  it("maps dimensions to Category and time dimensions by grain", () => {
    expect(
      deriveFlintFields(["advertiser"], metricsView).semantic_types,
    ).toEqual({ advertiser: "Category" });

    expect(
      deriveFlintFields(["__time"], metricsView, V1TimeGrain.TIME_GRAIN_MONTH)
        .semantic_types,
    ).toEqual({ __time: "YearMonth" });

    expect(
      deriveFlintFields(["__time"], metricsView, V1TimeGrain.TIME_GRAIN_DAY)
        .semantic_types,
    ).toEqual({ __time: "Date" });

    // A time-typed dimension that is not the primary time dimension is still temporal.
    expect(
      deriveFlintFields(
        ["event_time"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_YEAR,
      ).semantic_types,
    ).toEqual({ event_time: "Year" });
  });

  it("falls back to DateTime for sub-day grains and when no grain is set", () => {
    expect(
      deriveFlintFields(["__time"], metricsView).semantic_types.__time,
    ).toBe("DateTime");

    expect(
      deriveFlintFields(["__time"], metricsView, V1TimeGrain.TIME_GRAIN_HOUR)
        .semantic_types.__time,
    ).toBe("DateTime");
  });

  it("exposes display names so Flint labels axes and legends with them", () => {
    const { field_display_names } = deriveFlintFields(
      ["revenue", "advertiser", "dwell"],
      metricsView,
    );

    // dwell has no displayName, so it is omitted and Flint falls back to the column name.
    expect(field_display_names).toEqual({
      revenue: "Revenue",
      advertiser: "Advertiser",
    });
  });

  it("leaves unknown columns unannotated for Flint to infer", () => {
    const { semantic_types } = deriveFlintFields(["mystery"], metricsView);
    expect(semantic_types).toEqual({});
  });

  it("returns empty metadata when the metrics view is not yet resolved", () => {
    expect(deriveFlintFields(["revenue"], undefined)).toEqual({
      semantic_types: {},
      field_display_names: {},
    });
  });
});
