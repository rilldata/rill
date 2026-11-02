import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { describe, expect, it } from "vitest";
import {
  createComparisonTransforms,
  MeasureKeyField,
} from "./comparison-builder";

describe("createComparisonTransforms", () => {
  it("escapes the folded fields and keeps the raw measure name as the output column", () => {
    const [fold] = createComparisonTransforms(
      "__time",
      "stats.count",
      "query.context.metaData.type",
    );
    expect(fold).toEqual({
      fold: ["stats\\.count", `stats\\.count${ComparisonDeltaPreviousSuffix}`],
      as: [MeasureKeyField, "stats.count"],
    });
  });
});
