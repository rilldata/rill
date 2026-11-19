import { describe, expect, it } from "vitest";
import { at, chartData, vegaFieldReferences } from "../cartesian/test-fixtures";
import type { FunnelChartSpec } from "./FunnelChartProvider";
import { generateVLFunnelChartSpec } from "./spec";

// Nested Druid columns surface as fields whose names contain dots, which
// Vega would otherwise read as nested field paths.
const stage = "query.context.metaData.type";
const measure = "stats.count";

function data() {
  const result = chartData();
  result.fields = {
    [stage]: { name: stage, displayName: "Query Type" },
    [measure]: { name: measure, displayName: "Count" },
    "other.count": { name: "other.count", displayName: "Other Count" },
  };
  result.domainValues = { [stage]: ["Scheduled Report", "Other"] };
  return result;
}

describe("generateVLFunnelChartSpec with dotted field names", () => {
  it("escapes the stage and measure everywhere Vega resolves them as fields", () => {
    const config: FunnelChartSpec = {
      metrics_view: "druid_requests",
      stage: { field: stage, type: "nominal" },
      measure: { field: measure, type: "quantitative" },
      color: "stage",
    };
    const spec = generateVLFunnelChartSpec(config, data());

    expect(at(spec, "layer.0.encoding.color.field")).toBe(
      "query\\.context\\.metaData\\.type",
    );
    const refs = vegaFieldReferences(spec);
    expect(refs).toContain("stats\\.count");
    expect(refs).not.toContain(stage);
    expect(refs).not.toContain(measure);
  });

  it("escapes the folded measures of a multi-measure funnel", () => {
    const config: FunnelChartSpec = {
      metrics_view: "druid_requests",
      breakdownMode: "measures",
      measure: {
        field: measure,
        fields: [measure, "other.count"],
        type: "quantitative",
      },
      color: "name",
    };
    const spec = generateVLFunnelChartSpec(config, data());

    expect(at(spec, "transform.0.fold")).toEqual([
      "stats\\.count",
      "other\\.count",
    ]);
    expect(at(spec, "transform.0.as")).toEqual(["Measure", "value"]);
  });
});
