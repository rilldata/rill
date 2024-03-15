import { getChartYaml } from "@rilldata/web-common/features/charts/chartYaml";
import { describe, it, expect } from "vitest";

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
      getChartYaml(VegaLiteSpec, "SQL", {
        sql: `select * from AdBids
where publisher is not null`,
      }),
    ).toEqual(`kind: chart
data:
  sql: |-
    select * from AdBids
    where publisher is not null
vega_lite: |2-
    {
      "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
      "description": "A simple bar chart with embedded data.",
      "mark": "bar",
      "encoding": {
        "x": {
          "field": "time",
          "type": "nominal",
          "axis": {
            "labelAngle": 0
          }
        },
        "y": {
          "field": "total_sales",
          "type": "quantitative"
        }
      }
    }
`);
  });

  it("multi line metrics sql", () => {
    expect(
      getChartYaml(VegaLiteSpec, "MetricsSQL", {
        sql: `select * from AdBids
where publisher is not null`,
      }),
    ).toEqual(`kind: chart
data:
  metrics_sql: |-
    select * from AdBids
    where publisher is not null
vega_lite: |2-
    {
      "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
      "description": "A simple bar chart with embedded data.",
      "mark": "bar",
      "encoding": {
        "x": {
          "field": "time",
          "type": "nominal",
          "axis": {
            "labelAngle": 0
          }
        },
        "y": {
          "field": "total_sales",
          "type": "quantitative"
        }
      }
    }
`);
  });
});
