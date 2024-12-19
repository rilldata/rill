import { describe, expect, it } from "vitest";
import { getChartYaml } from "../chartYaml";

const VegaLiteSpec = `{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "description": "A simple bar chart with embedded data.",
  "mark": "bar",
  "encoding": {
    "x": {"field": "time", "type": "nominal", "axis": {"labelAngle": 0}},
    "y": {"field": "total_sales", "type": "quantitative"}
  }
}`;

describe("getComponentYaml", () => {
  it("multi line sql", () => {
    expect(
      getChartYaml(VegaLiteSpec, "sql", {
        sql: `select * from AdBids
where publisher is not null`,
      }),
    ).toEqual(`# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components

type: component
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
    ).toEqual(`# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components

type: component
data:
  metrics_sql: |-
    select * from AdBids
    where publisher is not null
vega_lite: |-
${VegaLiteSpec.replace(/^/gm, "  ")}
`);
  });
});
