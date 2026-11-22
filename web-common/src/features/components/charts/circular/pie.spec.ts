import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import * as vega from "vega";
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

describe("generateVLPieChartSpec with a text scale", () => {
  const measure = "order_count";
  const dimension = "region";
  const rows = [
    { [dimension]: "North", [measure]: 60 },
    { [dimension]: "South", [measure]: 40 },
  ];

  /**
   * Draws a donut of the given size at a text scale and returns the font
   * sizes Vega draws its slice labels and its center total with.
   */
  async function drawnFontSizes(textScale: number, size: number) {
    stubCanvasContext();
    // The total is formatted with the measure's own formatter, which the
    // chart renderer registers under the measure's sanitized name.
    vega.expressionFunction(sanitizeFieldName(measure), (value: number) =>
      String(value),
    );
    const config: CircularChartSpec = {
      metrics_view: "orders",
      innerRadius: 50,
      measure: { field: measure, type: "quantitative", showTotal: true },
      color: { field: dimension, type: "nominal" },
      labels: { show: true, format: "percent" },
    };
    const data = { ...chartData(), textScale };
    data.data = rows;
    data.fields = {
      [measure]: { name: measure, displayName: "Orders" },
      [dimension]: { name: dimension, displayName: "Region" },
    };
    data.domainValues = {
      [TOTAL_DOMAIN_KEY]: [100],
      [dimension]: ["North", "South"],
    };

    const spec = {
      ...(generateVLPieChartSpec(config, data) as TopLevelSpec),
      width: size,
      height: size,
    } as TopLevelSpec;
    const view = new vega.View(vega.parse(compile(spec).spec), {
      renderer: "none",
    });
    view.data("metrics-view", rows);
    await view.runAsync();

    type Node = { items?: Node[]; text?: string; fontSize?: number };
    const sizes: Record<string, number> = {};
    const visit = (node: Node) => {
      if (node.text !== undefined && node.fontSize !== undefined) {
        sizes[node.text] = node.fontSize;
      }
      node.items?.forEach(visit);
    };
    visit((view.scenegraph() as unknown as { root: Node }).root);
    return { label: sizes["60%"], total: sizes["100"] };
  }

  it("scales the slice labels with the text", async () => {
    const standard = await drawnFontSizes(1, 300);
    const xlarge = await drawnFontSizes(1.5, 300);

    expect(xlarge.label).toBeCloseTo(standard.label * 1.5, 5);
  });

  it("scales the center total where the donut's hole has room for it", async () => {
    const standard = await drawnFontSizes(1, 800);
    const xlarge = await drawnFontSizes(1.5, 800);

    // 32px is the cap at the standard size.
    expect(standard.total).toBe(32);
    expect(xlarge.total).toBe(48);
  });

  it("keeps the center total inside a hole it already fills", async () => {
    const standard = await drawnFontSizes(1, 300);
    const xlarge = await drawnFontSizes(1.5, 300);

    expect(standard.total).toBeLessThan(32);
    expect(xlarge.total).toBe(standard.total);
  });
});
