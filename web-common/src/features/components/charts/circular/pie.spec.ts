import { describe, expect, it } from "vitest";
import { chartData, vegaFieldReferences } from "../cartesian/test-fixtures";
import type { CircularChartSpec } from "./CircularChartProvider";
import { TOTAL_DOMAIN_KEY } from "./constants";
import { generateVLPieChartSpec } from "./pie";

describe("generateVLPieChartSpec with dotted field names", () => {
  it("escapes the measure in the slice label text encoding", () => {
    // Nested Druid columns surface as fields whose names contain dots, which
    // Vega would otherwise read as nested field paths.
    const measure = "stats.count";
    const dimension = "query.context.metaData.type";
    const config: CircularChartSpec = {
      metrics_view: "druid_requests",
      measure: { field: measure, type: "quantitative" },
      color: { field: dimension, type: "nominal" },
      labels: { show: true, format: "value" },
    };
    const data = chartData();
    data.fields = {
      [measure]: { name: measure, displayName: "Count" },
      [dimension]: { name: dimension, displayName: "Query Type" },
    };
    data.domainValues = {
      [TOTAL_DOMAIN_KEY]: [100],
      [dimension]: ["Scheduled Report", "Other"],
    };

    const spec = generateVLPieChartSpec(config, data);
    const refs = vegaFieldReferences(spec);
    expect(refs).toContain("stats\\.count");
    expect(refs).not.toContain(measure);
    expect(refs).not.toContain(dimension);
  });
});
