import { describe, expect, it } from "vitest";
import { at, chartData, vegaFieldReferences } from "../cartesian/test-fixtures";
import type { HeatmapChartSpec } from "./HeatmapChartProvider";
import { generateVLHeatmapSpec } from "./spec";

describe("generateVLHeatmapSpec with dotted field names", () => {
  it("escapes the measure in the data label text encoding", () => {
    const measure = "stats.count";
    const config: HeatmapChartSpec = {
      metrics_view: "druid_requests",
      x: { field: "query.dataSource.name", type: "nominal" },
      y: { field: "query.context.metaData.type", type: "nominal" },
      color: { field: measure, type: "quantitative" },
      show_data_labels: true,
    };
    const data = chartData();
    data.fields = {
      "query.dataSource.name": {
        name: "query.dataSource.name",
        displayName: "Data Source",
      },
      "query.context.metaData.type": {
        name: "query.context.metaData.type",
        displayName: "Query Type",
      },
      [measure]: { name: measure, displayName: "Count" },
    };
    data.domainValues = {};

    const spec = generateVLHeatmapSpec(config, data);
    expect(at(spec, "layer.1.encoding.text.field")).toBe("stats\\.count");
    const refs = vegaFieldReferences(spec);
    for (const name of [
      measure,
      "query.dataSource.name",
      "query.context.metaData.type",
    ]) {
      expect(refs).not.toContain(name);
    }
  });
});
