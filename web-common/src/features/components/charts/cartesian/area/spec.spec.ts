import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { at, chartData } from "../test-fixtures";
import { generateVLAreaChartSpec } from "./spec";

describe("generateVLAreaChartSpec with a dotted color dimension", () => {
  it("escapes the dimension in the hover pivot", () => {
    // Nested Druid columns surface as dimensions whose names contain dots,
    // which Vega would otherwise read as a nested field path.
    const dimension = "query.context.metaData.type";
    const data = chartData();
    data.fields[dimension] = { name: dimension, displayName: "Query Type" };
    data.domainValues = {
      ...data.domainValues,
      [dimension]: ["Scheduled Report", "Other"],
    };

    const config: CartesianChartSpec = {
      metrics_view: "druid_requests",
      x: { field: "__time", type: "temporal" },
      y: { field: "count", type: "quantitative", zeroBasedOrigin: true },
      color: { field: dimension, type: "nominal" },
    };
    const spec = generateVLAreaChartSpec(config, data);
    expect(at(spec, "layer.1.transform.0.pivot")).toBe(
      "query\\.context\\.metaData\\.type",
    );
    expect(at(spec, "layer.0.encoding.color.field")).toBe(
      "query\\.context\\.metaData\\.type",
    );
  });
});
