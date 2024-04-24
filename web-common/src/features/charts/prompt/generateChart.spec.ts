import { getChartYaml } from "@rilldata/web-common/features/charts/chartYaml";
import { describe, expect, it } from "vitest";

const VegaLiteSpec = `{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "description": "A simple bar chart with embedded data.",
  "mark": "bar",
  "encoding": {
    "x": {"field": "time", "type": "nominal", "axis": {"labelAngle": 0}},
    "y": {"field": "total_sales", "type": "quantitative"}
  }
}`;

describe("getChartYaml", () => {
  it("multi line sql", () => {
    expect(
      getChartYaml(VegaLiteSpec, "sql", {
        sql: `select * from AdBids
where publisher is not null`,
      }),
    ).toEqual(`# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

kind: chart
data:
  sql: |-
    select * from AdBids
    where publisher is not null
vega_lite: |-
${VegaLiteSpec.replace(/^/gm, "  ")}
`);
  });

  it("multi line metrics sql", () => {
    expect(
      getChartYaml(VegaLiteSpec, "metrics_sql", {
        sql: `select * from AdBids
where publisher is not null`,
      }),
    ).toEqual(`# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

kind: chart
data:
  metrics_sql: |-
    select * from AdBids
    where publisher is not null
vega_lite: |-
${VegaLiteSpec.replace(/^/gm, "  ")}
`);
  });
});
