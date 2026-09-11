import { mapAIResolverPropsToMetricsResolverQuery } from "@rilldata/web-common/features/explore-mappers/map-ai-resolver-props-to-metrics-resolver-query.ts";
import type { Expression } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { describe, expect, it } from "vitest";

describe("mapAIResolverPropsToMetricsResolverQuery", () => {
  it("maps the full scope of an AI report", () => {
    const where: Expression = {
      // The generated type declares `val` as an object, but scalar values are what the resolver accepts.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      cond: { op: "eq", exprs: [{ name: "country" }, { val: "US" as any }] },
    };
    expect(
      mapAIResolverPropsToMetricsResolverQuery(
        {
          explore: "sales",
          dimensions: ["country", "product"],
          measures: ["revenue"],
          time_range: { expression: "7D as of latest/D+1D" },
          comparison_time_range: {
            expression: "7D as of latest/D+1D offset -7D",
          },
          time_zone: "America/New_York",
          where,
        },
        "sales_metrics",
      ),
    ).toEqual({
      metrics_view: "sales_metrics",
      dimensions: [{ name: "country" }, { name: "product" }],
      measures: [{ name: "revenue" }],
      time_range: { expression: "7D as of latest/D+1D" },
      comparison_time_range: { expression: "7D as of latest/D+1D offset -7D" },
      time_zone: "America/New_York",
      where,
    });
  });

  it("only sets the metrics view when the report has no scope", () => {
    expect(
      mapAIResolverPropsToMetricsResolverQuery(
        { explore: "sales" },
        "sales_metrics",
      ),
    ).toEqual({ metrics_view: "sales_metrics" });
  });

  it("ignores empty dimension and measure lists", () => {
    expect(
      mapAIResolverPropsToMetricsResolverQuery(
        { dimensions: [], measures: [] },
        "sales_metrics",
      ),
    ).toEqual({ metrics_view: "sales_metrics" });
  });
});
