import {
  applyExactMeasures,
  exactTimeSeriesMeasureNames,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
import { prepareMeasuresForRequest } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type { V1MetricsViewAggregationMeasure } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("applyExactMeasures", () => {
  const measures: V1MetricsViewAggregationMeasure[] = [
    { name: "users" },
    { name: "revenue" },
    { name: "users__previous", comparisonValue: { measure: "users" } },
    { name: "users__delta_abs", comparisonDelta: { measure: "users" } },
    { name: "users__delta_rel", comparisonRatio: { measure: "users" } },
    { name: "users__percent_of_total", percentOfTotal: { measure: "users" } },
    { name: "revenue__delta_abs", comparisonDelta: { measure: "revenue" } },
  ];

  it("flags the measure and the computes derived from it", () => {
    expect(applyExactMeasures(measures, new Set(["users"]))).toEqual([
      { name: "users", exact: true },
      { name: "revenue" },
      {
        name: "users__previous",
        comparisonValue: { measure: "users" },
        exact: true,
      },
      {
        name: "users__delta_abs",
        comparisonDelta: { measure: "users" },
        exact: true,
      },
      {
        name: "users__delta_rel",
        comparisonRatio: { measure: "users" },
        exact: true,
      },
      {
        name: "users__percent_of_total",
        percentOfTotal: { measure: "users" },
        exact: true,
      },
      { name: "revenue__delta_abs", comparisonDelta: { measure: "revenue" } },
    ]);
  });

  it("is a no-op without exact measures", () => {
    expect(applyExactMeasures(measures, undefined)).toBe(measures);
    expect(applyExactMeasures(measures, new Set())).toBe(measures);
  });

  it("does not mutate the input", () => {
    const input = [{ name: "users" }];
    applyExactMeasures(input, new Set(["users"]));
    expect(input).toEqual([{ name: "users" }]);
  });
});

describe("exactTimeSeriesMeasureNames", () => {
  it("keeps the requested names that are exact", () => {
    expect(
      exactTimeSeriesMeasureNames(
        ["users", "revenue"],
        new Set(["users", "other"]),
      ),
    ).toEqual(["users"]);
  });

  it("is undefined when none are exact", () => {
    expect(
      exactTimeSeriesMeasureNames(["revenue"], new Set(["users"])),
    ).toBeUndefined();
    expect(exactTimeSeriesMeasureNames(["revenue"], undefined)).toBeUndefined();
  });
});

describe("prepareMeasuresForRequest with exact measures", () => {
  it("flags pivot measures and their comparison columns", () => {
    expect(
      prepareMeasuresForRequest(
        [
          { name: "users" },
          { name: "users__delta_abs" },
          { name: "users__delta_rel" },
          { name: "revenue" },
        ],
        undefined,
        ["users"],
      ),
    ).toEqual([
      { name: "users", exact: true },
      {
        name: "users__delta_abs",
        comparisonDelta: { measure: "users" },
        exact: true,
      },
      {
        name: "users__delta_rel",
        comparisonRatio: { measure: "users" },
        exact: true,
      },
      { name: "revenue" },
    ]);
  });
});
