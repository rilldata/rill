import * as vega from "vega";
import { compile } from "vega-lite";
import { describe, expect, it } from "vitest";
import {
  V1TimeGrain,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { compileFlintSpec } from "./compile";
import { deriveFlintFields } from "./semantic-types";

/**
 * End-to-end check of the render path: metrics view spec + query rows → Flint semantics →
 * compiled Vega-Lite. This is the wiring FlintChartRenderer performs, minus the Svelte shell.
 */

const metricsView: V1MetricsViewSpec = {
  timeDimension: "timestamp",
  measures: [
    {
      name: "bid_price_sum",
      displayName: "Sum of Bid Price",
      formatPreset: "currency_usd",
    },
    { name: "total_records", displayName: "Total records" },
  ],
  dimensions: [{ name: "publisher", displayName: "Publisher" }],
};

const rows = [
  { publisher: "Yahoo", bid_price_sum: 1200, total_records: 40 },
  { publisher: "Google", bid_price_sum: 1450, total_records: 55 },
  { publisher: "Microsoft", bid_price_sum: 980, total_records: 31 },
];

function compileBar() {
  const columns = ["publisher", "bid_price_sum"];
  return compileFlintSpec({
    spec: {
      chartType: "Bar Chart",
      encodings: {
        x: { field: "publisher", sortBy: "y", sortOrder: "descending" },
        y: { field: "bid_price_sum" },
      },
    },
    rows,
    fields: deriveFlintFields(columns, metricsView),
    size: { width: 600, height: 400 },
    measureFields: ["bid_price_sum"],
  });
}

describe("compileFlintSpec", () => {
  it("compiles a Flint spec into a renderable Vega-Lite spec", () => {
    const { spec, error } = compileBar();

    expect(error).toBeNull();
    const compiled = spec as unknown as Record<string, unknown>;
    expect(compiled.mark).toBeTruthy();
    expect(compiled.encoding).toBeTruthy();
    // Flint inlines the rows it was given.
    expect((compiled.data as { values: unknown[] }).values).toHaveLength(3);
  });

  it("fits axes and legends inside the measured chart container", () => {
    const { spec } = compileBar();
    const compiled = spec as unknown as Record<string, unknown>;

    // Without fit autosizing, Vega treats the measured height as the plot height and adds axes
    // and bottom legends below it, leaving a persistent scrollbar of exactly that guide height.
    expect(compiled.autosize).toEqual({
      type: "fit",
      contains: "padding",
    });
    // Pinned to the measured box; Flint's own step and view sizes are dropped.
    expect(compiled.width).toBe(600);
    expect(compiled.height).toBe(400);

    const view = (compiled.config as { view?: Record<string, unknown> })?.view;
    expect(view?.continuousWidth).toBeUndefined();
    expect(view?.continuousHeight).toBeUndefined();
  });

  it("stretches a discrete axis across the container", async () => {
    // Left unsized, a discrete axis falls back to a 20px step and Vega-Lite drops fit on it, so
    // these three bars would render 60px wide.
    const { spec } = compileBar();
    const view = new vega.View(vega.parse(compile(spec as never).spec), {
      renderer: "none",
    });
    await view.runAsync();
    expect(view.width()).toBeGreaterThan(400);
  });

  it("scales a donut's hole with the chart rather than in fixed pixels", async () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Donut Chart",
        encodings: {
          color: { field: "publisher" },
          size: { field: "bid_price_sum" },
        },
        chartProperties: { innerRadius: 50 },
      },
      rows,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 800, height: 800 },
    });
    const mark = (spec as unknown as { mark: Record<string, unknown> }).mark;
    expect(mark.innerRadius).toEqual({ expr: "min(width, height) / 2 * 0.5" });

    const view = new vega.View(vega.parse(compile(spec as never).spec), {
      renderer: "none",
    });
    await view.runAsync();
    // Half of the outer radius, which is roughly half the view.
    type SceneGroup = {
      items: { marktype?: string; items: { innerRadius?: number }[] }[];
    };
    const scene = view.scenegraph() as unknown as {
      root: { items: SceneGroup[] };
    };
    const arcs = scene.root.items[0].items.find(
      (item) => item.marktype === "arc",
    );
    expect(arcs?.items[0].innerRadius).toBeGreaterThan(150);
  });

  it("keeps Flint's sizing for faceted charts, which fit autosizing cannot lay out", () => {
    const sites = Array.from({ length: 12 }, (_, i) => `site_${i}`);
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Violin Plot",
        encodings: { x: { field: "publisher" }, y: { field: "bid_price_sum" } },
      },
      rows: sites.flatMap((publisher, i) =>
        Array.from({ length: 20 }, (_, j) => ({
          publisher,
          bid_price_sum: (i * 31 + j * 7) % 100,
        })),
      ),
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 1200, height: 600 },
    });
    const compiled = spec as unknown as Record<string, unknown>;
    expect(compiled.encoding).toHaveProperty("facet");
    expect(compiled.autosize).toBeUndefined();
    // Each cell takes its share of the canvas, rather than Vega-Lite's 200px default.
    expect(Number(compiled.width) * 12).toBeLessThanOrEqual(1200);
  });

  it("strips Flint's non-standard bookkeeping keys", () => {
    const { spec } = compileBar();
    const keys = Object.keys(spec as unknown as Record<string, unknown>);

    // _warnings/_width/_height/_options/_transform/_pivot are not valid Vega-Lite
    // and make vega-embed reject the spec.
    expect(keys.filter((key) => key.startsWith("_"))).toEqual([]);
  });

  it("labels axes with the metrics view display names", () => {
    const { spec } = compileBar();
    const encoding = (
      spec as unknown as { encoding: Record<string, { title?: string }> }
    ).encoding;

    expect(encoding.x.title).toBe("Publisher");
    expect(encoding.y.title).toBe("Sum of Bid Price");
  });

  it("routes measures through Rill's registered formatter", () => {
    const { spec } = compileBar();
    const y = (
      spec as unknown as {
        encoding: {
          y: {
            formatType?: string;
            format?: string;
            axis?: Record<string, unknown>;
          };
        };
      }
    ).encoding.y;

    // Without this, a currency measure would render with Flint's generic number format.
    expect(y.formatType).toBe("rill_bid_price_sum");
    expect(y.format).toBeUndefined();
    expect(y.axis?.formatType).toBe("rill_bid_price_sum");
    expect(y.axis?.format).toBeUndefined();
  });

  it("does not touch dimension encodings", () => {
    const { spec } = compileBar();
    const x = (spec as unknown as { encoding: { x: Record<string, unknown> } })
      .encoding.x;
    expect(x.formatType).toBeUndefined();
  });

  it("surfaces overflow warnings rather than dropping rows silently", () => {
    const many = Array.from({ length: 300 }, (_, i) => ({
      publisher: `p${i}`,
      bid_price_sum: ((i * 13) % 500) + 1,
    }));

    const { spec, warnings } = compileFlintSpec({
      spec: {
        chartType: "Bar Chart",
        encodings: { x: { field: "publisher" }, y: { field: "bid_price_sum" } },
      },
      rows: many,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 600, height: 400 },
    });

    const rendered = (spec as unknown as { data: { values: unknown[] } }).data
      .values;
    expect(rendered.length).toBeLessThan(many.length);
    expect(warnings.some((warning) => warning.code === "overflow")).toBe(true);
  });

  it("treats the time dimension at the dashboard's grain", () => {
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Line Chart",
        encodings: {
          x: { field: "timestamp" },
          y: { field: "bid_price_sum" },
        },
      },
      rows: [
        { timestamp: "2024-01", bid_price_sum: 10 },
        { timestamp: "2024-02", bid_price_sum: 20 },
      ],
      fields: deriveFlintFields(
        ["timestamp", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_MONTH,
      ),
    });

    expect(error).toBeNull();
    const x = (spec as unknown as { encoding: { x: { type: string } } })
      .encoding.x;
    expect(x.type).toBe("temporal");
  });

  it("reports an unknown chart type instead of throwing", () => {
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Not A Chart",
        encodings: { x: { field: "publisher" } },
      },
      rows,
      fields: deriveFlintFields(["publisher"], metricsView),
    });

    expect(spec).toBeNull();
    expect(error).toBeTruthy();
  });

  it("reports a missing chart type", () => {
    const { error } = compileFlintSpec({
      spec: {},
      rows,
      fields: deriveFlintFields(["publisher"], metricsView),
    });
    expect(error).toBe("No chart type provided");
  });
});
