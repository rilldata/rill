import * as vega from "vega";
import { compile } from "vega-lite";
import { describe, expect, it } from "vitest";
import {
  TypeCode,
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
    {
      name: "impressions",
      displayName: "Impressions",
      dataType: { code: TypeCode.CODE_INT64 },
    },
  ],
  dimensions: [
    { name: "publisher", displayName: "Publisher" },
    { name: "device", displayName: "Device" },
  ],
};

const rows = [
  { publisher: "Yahoo", bid_price_sum: 1200, total_records: 40 },
  { publisher: "Google", bid_price_sum: 1450, total_records: 55 },
  { publisher: "Microsoft", bid_price_sum: 980, total_records: 31 },
];

type SceneNode = {
  role?: string;
  items?: SceneNode[];
  text?: unknown;
  datum?: Record<string, unknown>;
  opacity?: number;
};

/**
 * Renders a compiled spec headless and returns the items Vega drew in the given role, such as
 * "axis-label" or "mark". Items label overlap removal hid are left out.
 */
async function renderRole(spec: unknown, role: string): Promise<SceneNode[]> {
  const view = new vega.View(vega.parse(compile(spec as never).spec), {
    renderer: "none",
  });
  await view.runAsync();
  const found: SceneNode[] = [];
  const visit = (node: SceneNode) => {
    if (node.role === role) found.push(...(node.items ?? []));
    for (const item of node.items ?? []) item.items?.forEach(visit);
  };
  visit((view.scenegraph() as unknown as { root: SceneNode }).root);
  return found.filter((item) => item.opacity !== 0);
}

/** The ticks of a chart's time axis, as Vega placed them. */
async function timeTicks(spec: unknown): Promise<Date[]> {
  return (await renderRole(spec, "axis-label"))
    .map((label) => label.datum?.value)
    .filter((value): value is Date => value instanceof Date);
}

/** The parts of a compiled Vega-Lite view the assertions below read. */
type Encoding = {
  field?: string;
  type?: string;
  title?: unknown;
  format?: string;
  formatType?: string;
  stack?: unknown;
  scale?: Record<string, unknown>;
  axis?: Record<string, unknown> | null;
  header?: Record<string, unknown>;
};
type View = {
  mark?: Record<string, unknown>;
  encoding?: Record<string, Encoding>;
  transform?: Record<string, unknown>[];
  layer?: View[];
  hconcat?: View[];
  title?: Record<string, unknown>;
  width?: number;
  spacing?: number;
  autosize?: unknown;
};
const viewOf = (spec: unknown) => spec as View;

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

  it("defers typography, colors and the color scheme to Rill's theme", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Bar Chart",
        encodings: {
          x: { field: "publisher" },
          y: { field: "bid_price_sum" },
          color: { field: "publisher" },
        },
      },
      rows,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });
    const compiled = spec as unknown as {
      config: Record<string, Record<string, unknown> | undefined>;
      encoding: { color: { scale?: Record<string, unknown> } };
    };

    // Flint's font sizes and grey titles would override the theme's, which the renderer passes
    // as the embed config; so would its tableau10 scheme override the project palette.
    for (const group of ["axisX", "axisY", "legend"]) {
      expect(compiled.config[group]?.labelFontSize).toBeUndefined();
      expect(compiled.config[group]?.titleFontSize).toBeUndefined();
      expect(compiled.config[group]?.titleColor).toBeUndefined();
    }
    expect(compiled.encoding.color.scale?.scheme).toBeUndefined();
  });

  it("keeps a color scheme the author chose", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Bar Chart",
        encodings: {
          x: { field: "publisher" },
          y: { field: "bid_price_sum" },
          color: { field: "publisher", scheme: "category20" },
        },
      },
      rows,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 600, height: 400 },
    });
    const color = (
      spec as unknown as { encoding: { color: { scale: { scheme: string } } } }
    ).encoding.color;
    expect(color.scale.scheme).toBe("category20");
  });

  it("labels the series of a multi-measure fold with the measures' display names", () => {
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Bar Chart",
        encodings: {
          x: { field: "publisher" },
          y: [{ field: "bid_price_sum" }, { field: "total_records" }],
        },
      },
      rows,
      fields: deriveFlintFields(
        ["publisher", "bid_price_sum", "total_records"],
        metricsView,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum", "total_records"],
    });

    expect(error).toBeNull();
    const values = (
      spec as unknown as { data: { values: Record<string, unknown>[] } }
    ).data.values;
    expect(new Set(values.map((row) => row.__flint_series_key))).toEqual(
      new Set(["Sum of Bid Price", "Total records"]),
    );
  });

  it("spans the box with the panels of a row-faceted chart, less its longest row label", () => {
    const days = ["2024-01-01", "2024-01-02", "2024-01-03"];
    const compileRows = (publishers: string[]) =>
      compileFlintSpec({
        spec: {
          chartType: "Line Chart",
          encodings: {
            x: { field: "timestamp" },
            y: { field: "bid_price_sum" },
            row: { field: "publisher" },
          },
        },
        rows: publishers.flatMap((publisher, p) =>
          days.map((day, i) => ({
            publisher,
            timestamp: `${day}T00:00:00Z`,
            bid_price_sum: 1000 * (p + 1) + i * 10,
          })),
        ),
        fields: deriveFlintFields(
          ["timestamp", "publisher", "bid_price_sum"],
          metricsView,
          V1TimeGrain.TIME_GRAIN_DAY,
        ),
        size: { width: 1200, height: 600 },
        measureFields: ["bid_price_sum"],
      });
    type RowConfig = {
      config: {
        view: { continuousWidth: number };
        headerRow: {
          labelAngle: number;
          labelAlign: string;
          labelLimit: number;
        };
      };
    };

    const short = compileRows(["Yahoo", "Google", "Microsoft"]);
    expect(short.error).toBeNull();
    const { config } = short.spec as unknown as RowConfig;
    // The width less the room for the axis and the row title (130px) and for "Microsoft".
    expect(config.view.continuousWidth).toBe(1200 - 130 - 59);
    // Upright labels, lined up against the axis.
    expect(config.headerRow.labelAngle).toBe(0);
    expect(config.headerRow.labelAlign).toBe("left");
    expect(config.headerRow.labelLimit).toBe(110);

    // A label past the limit is cut at it, so it takes no more room than the limit.
    const long = compileRows([
      "Johnson & Johnson Consumer Health",
      "Procter & Gamble International",
    ]);
    expect(
      (long.spec as unknown as RowConfig).config.view.continuousWidth,
    ).toBe(1200 - 130 - 110);
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

  it("labels and formats a waterfall's derived encodings like the measure they sum", () => {
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Waterfall Chart",
        encodings: { x: { field: "timestamp" }, y: { field: "bid_price_sum" } },
      },
      rows: [
        { timestamp: "2024-01-01T00:00:00Z", bid_price_sum: 1200 },
        { timestamp: "2024-01-08T00:00:00Z", bid_price_sum: 1450 },
        { timestamp: "2024-01-15T00:00:00Z", bid_price_sum: 980 },
      ],
      fields: deriveFlintFields(
        ["timestamp", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_WEEK,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    expect(error).toBeNull();
    const layers = (
      spec as unknown as {
        layer: {
          encoding: Record<
            string,
            {
              field?: string;
              type?: string;
              title?: string;
              formatType?: string;
              axis?: Record<string, unknown>;
            }
          >;
        }[];
      }
    ).layer;
    const sums = layers
      .map((layer) => layer.encoding.y)
      .filter((y) => y?.type === "quantitative");
    expect(sums.length).toBeGreaterThan(1);
    // Flint sums the measure under its own field names and titles them with the raw measure name;
    // every encoding on the shared y channel takes the measure's formatter and the display name.
    for (const y of sums) {
      expect(y.formatType).toBe("rill_bid_price_sum");
      expect(y.axis?.format).toBeUndefined();
      expect(y.title).not.toBe("bid_price_sum");
    }
    expect(sums.some((y) => y.title === "Sum of Bid Price")).toBe(true);
    // The step axis gets the time field's label instead of "timestamp, __wf_lead", and its
    // category labels are the weeks rather than raw timestamps.
    const x = layers.map((layer) => layer.encoding.x).find(Boolean);
    expect(x?.title).toBe("Time");
    expect(x?.axis?.labelExpr).toContain("timeFormat");
  });

  it("formats a time field drawn on a discrete axis at its grain", () => {
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Slope Chart",
        encodings: {
          x: { field: "timestamp" },
          y: { field: "bid_price_sum" },
          color: { field: "publisher" },
        },
      },
      rows: [
        {
          timestamp: "2024-01-01T00:00:00Z",
          publisher: "Yahoo",
          bid_price_sum: 10,
        },
        {
          timestamp: "2024-02-01T00:00:00Z",
          publisher: "Yahoo",
          bid_price_sum: 20,
        },
        {
          timestamp: "2024-01-01T00:00:00Z",
          publisher: "Google",
          bid_price_sum: 15,
        },
        {
          timestamp: "2024-02-01T00:00:00Z",
          publisher: "Google",
          bid_price_sum: 12,
        },
      ],
      fields: deriveFlintFields(
        ["timestamp", "publisher", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_MONTH,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    expect(error).toBeNull();
    const encoding = (
      spec as unknown as {
        encoding: { x: { type: string; axis?: { labelExpr?: string } } };
      }
    ).encoding;
    expect(encoding.x.type).toBe("ordinal");
    expect(encoding.x.axis?.labelExpr).toBe(
      "timeFormat(toDate(datum.value), '%b %Y')",
    );
  });

  it("reaches the category axis inside a concatenated chart", () => {
    // A bar table is a concatenation of panels; each one draws the time field on a nominal y axis.
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Bar Table",
        encodings: { y: { field: "timestamp" }, x: { field: "bid_price_sum" } },
      },
      rows: [
        { timestamp: "2024-01-01T00:00:00Z", bid_price_sum: 1200 },
        { timestamp: "2024-02-01T00:00:00Z", bid_price_sum: 1450 },
        { timestamp: "2024-03-01T00:00:00Z", bid_price_sum: 980 },
      ],
      fields: deriveFlintFields(
        ["timestamp", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_MONTH,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    expect(error).toBeNull();
    const panels = (
      spec as unknown as {
        hconcat?: {
          encoding?: {
            y?: { field?: string; axis?: { labelExpr?: string } | null };
          };
        }[];
      }
    ).hconcat;
    const timeAxes = (panels ?? [])
      .map((panel) => panel.encoding?.y)
      .filter((y) => y?.field === "timestamp");
    expect(timeAxes.length).toBeGreaterThan(0);
    for (const y of timeAxes) {
      if (y?.axis === null) continue; // a hidden axis has no labels to format
      expect(y?.axis?.labelExpr).toContain("timeFormat");
    }
  });

  it("ticks a day-grain time axis on its days", async () => {
    // Local midnights, which is how the renderers hand Vega the buckets. Flint pads the banded
    // axis by half a day at each end, so it starts and ends at noon; over a wide box Vega then
    // ticked every 12 hours, and label overlap removal kept only the noons.
    const days = Array.from({ length: 8 }, (_, i) => new Date(2026, 9, 30 + i));
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Bar Chart",
        encodings: { x: "timestamp", y: "bid_price_sum" },
      },
      rows: days.map((timestamp, i) => ({
        timestamp,
        bid_price_sum: 1000 + i,
      })),
      fields: deriveFlintFields(
        ["timestamp", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_DAY,
      ),
      size: { width: 800, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    const ticks = await timeTicks(spec);
    expect(ticks).toHaveLength(days.length);
    for (const tick of ticks) expect(tick.getHours()).toBe(0);
  });

  it("keeps an hourly time axis on hours over two days and steps it by days over eight", async () => {
    // From 9 AM, so that 12-hour ticks would start at noon, as they did on an eight-day line.
    const hourly = (hours: number) =>
      compileFlintSpec({
        spec: {
          chartType: "Line Chart",
          encodings: { x: "timestamp", y: "bid_price_sum" },
        },
        rows: Array.from({ length: hours }, (_, i) => ({
          timestamp: new Date(2026, 9, 30, 9 + i),
          bid_price_sum: 100 + (i % 7),
        })),
        fields: deriveFlintFields(
          ["timestamp", "bid_price_sum"],
          metricsView,
          V1TimeGrain.TIME_GRAIN_HOUR,
        ),
        size: { width: 600, height: 400 },
        measureFields: ["bid_price_sum"],
      }).spec;

    const twoDays = await timeTicks(hourly(48));
    expect(twoDays.some((tick) => tick.getHours() !== 0)).toBe(true);
    const eightDays = await timeTicks(hourly(8 * 24));
    expect(eightDays.length).toBeGreaterThan(1);
    for (const tick of eightDays) expect(tick.getHours()).toBe(0);
  });

  it("labels quarters from Q1", async () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Slope Chart",
        encodings: { x: "timestamp", y: "bid_price_sum", color: "publisher" },
      },
      rows: [new Date(2024, 0, 1), new Date(2024, 3, 1)].flatMap(
        (timestamp, i) =>
          rows.map((row) => ({
            timestamp,
            publisher: row.publisher,
            bid_price_sum: row.bid_price_sum + i * 100,
          })),
      ),
      fields: deriveFlintFields(
        ["timestamp", "publisher", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_QUARTER,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    const labels = (await renderRole(spec, "axis-label")).map(
      (label) => label.text,
    );
    expect(labels).toEqual(expect.arrayContaining(["Q1 2024", "Q2 2024"]));
  });

  it("formats a count measure's axis with its own formatter, on integer ticks", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Bar Chart",
        encodings: { x: "publisher", y: "impressions" },
      },
      rows: rows.map((row, i) => ({
        publisher: row.publisher,
        impressions: 600_000 * (i + 1),
      })),
      fields: deriveFlintFields(["publisher", "impressions"], metricsView),
      size: { width: 600, height: 400 },
      measureFields: ["impressions"],
    });

    const axis = viewOf(spec).encoding?.y.axis;
    // Flint's integer label expression formats with ',d' and would override the formatter.
    expect(axis?.labelExpr).toBeUndefined();
    expect(axis?.formatType).toBe("rill_impressions");
    expect(axis?.tickMinStep).toBe(1);
  });

  it("labels a 100% stacked count axis with percentages", async () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Stacked Bar Chart",
        encodings: { x: "publisher", y: "impressions", color: "device" },
        chartProperties: { stackMode: "normalize" },
      },
      rows: rows.flatMap((row, i) =>
        ["Phone", "TV"].map((device, d) => ({
          publisher: row.publisher,
          device,
          impressions: 100 * (i + 1) + d * 50,
        })),
      ),
      fields: deriveFlintFields(
        ["publisher", "device", "impressions"],
        metricsView,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["impressions"],
    });

    const y = viewOf(spec).encoding?.y;
    expect(y?.stack).toBe("normalize");
    expect(y?.axis).toEqual({ format: ".0%" });
    // The tooltip still shows the count.
    expect(y?.formatType).toBe("rill_impressions");
    const labels = (await renderRole(spec, "axis-label")).map(
      (label) => label.text,
    );
    expect(labels).toEqual(expect.arrayContaining(["0%", "50%", "100%"]));
  });

  it("gives a donut a hole when the spec sets none", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Donut Chart",
        encodings: { color: "publisher", size: "bid_price_sum" },
      },
      rows,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 600, height: 400 },
    });
    // Flint's donut is its pie, whose hole defaults to 0.
    expect(viewOf(spec).mark?.innerRadius).toEqual({
      expr: "min(width, height) / 2 * 0.5",
    });
  });

  it("ranks a bump chart's measure at each step and keeps the value in the tooltip", async () => {
    const days = [new Date(2026, 9, 30), new Date(2026, 9, 31)];
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType: "Bump Chart",
        encodings: { x: "timestamp", y: "bid_price_sum", color: "publisher" },
      },
      // Google leads on the first day, then the order flips.
      rows: days.flatMap((timestamp, i) =>
        rows.map((row, r) => ({
          timestamp,
          publisher: row.publisher,
          bid_price_sum: i === 0 ? row.bid_price_sum : 2000 - r * 100,
        })),
      ),
      fields: deriveFlintFields(
        ["timestamp", "publisher", "bid_price_sum"],
        metricsView,
        V1TimeGrain.TIME_GRAIN_DAY,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    expect(error).toBeNull();
    const encoding = viewOf(spec).encoding!;
    expect(encoding.y.title).toBe("Rank by Sum of Bid Price");
    // Rank 1 on top, with every rank labelled.
    expect(encoding.y.scale?.reverse).toBe(true);
    expect(encoding.y.axis?.values).toEqual([1, 2, 3]);
    expect(encoding.tooltip).toContainEqual({
      field: "bid_price_sum",
      type: "quantitative",
      title: "Sum of Bid Price",
      formatType: "rill_bid_price_sum",
    });

    // The renderers register a formatter per measure; the tooltip calls this one.
    vega.expressionFunction("rill_bid_price_sum", (value: number) =>
      String(value),
    );
    const ranks = new Map<string, unknown>();
    for (const { datum } of await renderRole(spec, "mark")) {
      if (!datum) continue;
      const day = new Date(datum.timestamp as string).getDate();
      ranks.set(`${day} ${String(datum.publisher)}`, datum.__rill_rank);
    }
    expect(Object.fromEntries(ranks)).toEqual({
      "30 Google": 1,
      "30 Yahoo": 2,
      "30 Microsoft": 3,
      "31 Yahoo": 1,
      "31 Google": 2,
      "31 Microsoft": 3,
    });
  });

  it("titles a bar table with display names and formats its values, inside its box", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Bar Table",
        encodings: { y: "publisher", x: "bid_price_sum" },
      },
      rows,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 600, height: 300 },
      measureFields: ["bid_price_sum"],
    });

    const view = viewOf(spec);
    const panels = view.hconcat!;
    const [bars, values] = [panels[0], panels[panels.length - 1]];
    expect(bars.title?.text).toBe("Publisher");
    expect(values.title?.text).toBe("Sum of Bid Price");
    expect(values.encoding?.text).toEqual({
      field: "__bt_val",
      type: "quantitative",
      formatType: "rill_bid_price_sum",
    });
    // The label gutter, the panels and the gaps between them leave room for Vega's padding,
    // and the height is fitted to the box, headers included.
    const gutter = Number(bars.encoding?.y.axis?.labelPadding);
    const used =
      gutter +
      panels.reduce((sum, panel) => sum + Number(panel.width), 0) +
      Number(view.spacing) * (panels.length - 1);
    expect(used).toBeLessThanOrEqual(600 - 10);
    expect(view.autosize).toEqual({ type: "fit-y", contains: "padding" });
  });

  it("centers violins in their panels and keeps them within the data", () => {
    const sites = Array.from({ length: 12 }, (_, i) => `site_${i}`);
    const violinRows = sites.flatMap((publisher, i) =>
      Array.from({ length: 20 }, (_, j) => ({
        publisher,
        bid_price_sum: 5 + ((i * 31 + j * 7) % 100),
      })),
    );
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Violin Plot",
        encodings: { x: { field: "publisher" }, y: { field: "bid_price_sum" } },
      },
      rows: violinRows,
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 1200, height: 600 },
    });

    const view = viewOf(spec);
    // The theme's area outline, drawn in x order, zig-zagged across a horizontal area.
    expect(view.mark?.line).toBe(false);
    // The density axis spans the violins exactly, so each sits in the middle of its panel.
    expect(view.encoding?.x.scale).toMatchObject({ nice: false });
    const values = violinRows.map((row) => row.bid_price_sum);
    expect(view.transform?.[0].extent).toEqual([
      Math.min(...values),
      Math.max(...values),
    ]);
    // Each violin is as wide as its panel allows, at its own widest point.
    expect(view.transform).toContainEqual({
      joinaggregate: [
        { op: "max", field: "density", as: "__rill_violin_peak" },
      ],
      groupby: ["publisher"],
    });
    // The panels, and the room for the value axis, fit the box.
    expect(view.spacing).toBe(0);
    expect(12 * view.width! + 80 + 10).toBeLessThanOrEqual(1200);
    expect(view.encoding?.facet.header?.labelLimit).toBe(view.width);
  });

  it("keeps a rose chart's labels off its short wedges, with room for them past its long ones", async () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Rose Chart",
        encodings: { x: "publisher", y: "bid_price_sum" },
      },
      rows: [...rows, { publisher: "Tiny", bid_price_sum: 20 }],
      fields: deriveFlintFields(["publisher", "bid_price_sum"], metricsView),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum"],
    });

    const [arcs, labels] = viewOf(spec).layer!;
    // Both layers share a radius that stops short of the box by the longest label.
    expect(arcs.encoding?.radius.scale?.range).toEqual(
      labels.encoding?.radius.scale?.range,
    );
    expect(labels.mark).toMatchObject({ radiusOffset: 6 });
    const drawn = (await renderRole(spec, "mark"))
      .map((item) => item.text)
      .filter((text) => typeof text === "string");
    expect(drawn).toEqual(
      expect.arrayContaining(["Yahoo", "Google", "Microsoft"]),
    );
    expect(drawn).not.toContain("Tiny");
  });

  it("keeps a color scheme the author set on a grouped bar chart's group", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Grouped Bar Chart",
        encodings: {
          x: "publisher",
          y: "bid_price_sum",
          group: { field: "device", scheme: "category20" },
        },
      },
      rows: rows.flatMap((row) =>
        ["Phone", "TV"].map((device) => ({ ...row, device })),
      ),
      fields: deriveFlintFields(
        ["publisher", "device", "bid_price_sum"],
        metricsView,
      ),
      size: { width: 600, height: 400 },
    });
    // Flint draws the group on the color channel.
    expect(viewOf(spec).encoding?.color.scale?.scheme).toBe("category20");
  });

  it("formats each measure of a layered chart with its own formatter", () => {
    const { spec } = compileFlintSpec({
      spec: {
        chartType: "Bullet Chart",
        encodings: {
          y: "publisher",
          x: "bid_price_sum",
          goal: "total_records",
        },
      },
      rows,
      fields: deriveFlintFields(
        ["publisher", "bid_price_sum", "total_records"],
        metricsView,
      ),
      size: { width: 600, height: 400 },
      measureFields: ["bid_price_sum", "total_records"],
    });

    const x = (field: string) =>
      viewOf(spec)
        .layer?.map((layer) => layer.encoding?.x)
        .find((encoding) => encoding?.field === field);
    expect(x("bid_price_sum")?.formatType).toBe("rill_bid_price_sum");
    // The goal shares the axis with the bars but keeps its own format.
    expect(x("total_records")?.formatType).toBe("rill_total_records");
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
