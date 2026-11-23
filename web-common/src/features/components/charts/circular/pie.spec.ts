import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import {
  chartData,
  stubCanvasContext,
  vegaFieldReferences,
} from "../cartesian/test-fixtures";
import type { CircularChartSpec } from "./CircularChartProvider";
import {
  OTHER_VALUE,
  OTHER_VALUE_DOMAIN_KEY,
  TOTAL_DOMAIN_KEY,
} from "./constants";
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

describe('generateVLPieChartSpec with an "Other" slice', () => {
  // The provider appends the "Other" row to the dataset, so the spec has to keep
  // reading the named `metrics-view` source: that is the name Chart.svelte
  // updates the view through. Swapping it for inline values, or leaving it for
  // Vega-Lite to drop as unread, failed every data update with
  // "Unrecognized data set: metrics-view".
  const measure = "post_count";
  const dimension = "post_title";
  const variants: Array<[string, Partial<CircularChartSpec>]> = [
    ["on its own", {}],
    ["with slice labels", { labels: { show: true, format: "value" } }],
    [
      "with a center total",
      { measure: { field: measure, type: "quantitative", showTotal: true } },
    ],
  ];

  for (const [variant, overrides] of variants) {
    it(`keeps the metrics-view dataset ${variant}`, () => {
      stubCanvasContext();
      const config: CircularChartSpec = {
        metrics_view: "reddit_posts",
        measure: { field: measure, type: "quantitative" },
        color: { field: dimension, type: "nominal" },
        show_other: true,
        ...overrides,
      };
      const data = chartData();
      data.data = [
        { [dimension]: "A", [measure]: 60 },
        { [dimension]: "B", [measure]: 30 },
        { [dimension]: OTHER_VALUE, [measure]: 10 },
      ];
      data.domainValues = {
        [TOTAL_DOMAIN_KEY]: [100],
        [OTHER_VALUE_DOMAIN_KEY]: [10],
        [dimension]: ["A", "B", OTHER_VALUE],
      };

      const spec = generateVLPieChartSpec(config, data);
      expect(spec.data).toEqual({ name: "metrics-view" });

      const sources = compile(spec as TopLevelSpec).spec.data?.map(
        (source) => source.name,
      );
      expect(sources).toContain("metrics-view");
    });
  }
});
