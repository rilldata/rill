import type { VisualizationSpec } from "svelte-vega";
import * as vega from "vega";
import { compile, type Config, type TopLevelSpec } from "vega-lite";
import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { stubCanvasContext } from "../../features/components/charts/cartesian/test-fixtures";
import { fitLegends, setContainerSize } from "./legend-layout";
import { getRillTheme } from "./vega-config";
import { refitOnPlotResize } from "./view-refit";

const regions = ["North America", "Latin America", "Western Europe"];
const tiers = [
  "Enterprise Accounts",
  "Mid-Market Accounts",
  "Small Business Accounts",
  "Public Sector Accounts",
];
// Each list is in alphabetical order, as Vega-Lite lists a dimension's values in a legend.
/** A dozen names too long to stand side by side. */
const segments = regions
  .flatMap((region) => tiers.map((tier) => `${region} - ${tier}`))
  .sort();
const channels = ["Audio", "Display", "Email", "Search", "Social", "Video"];

function rowsOf(series: string[]) {
  return series.flatMap((name, i) =>
    ["a", "b", "c"].map((category, c) => ({
      series: name,
      category,
      value: 100 + i * 7 + c,
      rate: (i + c) / 100,
    })),
  );
}

function barChart(
  series: string[],
  legend?: Record<string, unknown> | null,
): VisualizationSpec {
  return {
    width: 560,
    height: 300,
    autosize: { type: "fit", contains: "padding" },
    data: { values: rowsOf(series) },
    mark: "bar",
    encoding: {
      x: { field: "category", type: "nominal" },
      y: { field: "value", type: "quantitative" },
      color: {
        field: "series",
        type: "nominal",
        ...(legend !== undefined && { legend }),
      },
    },
  } as VisualizationSpec;
}

const theme: Config = {
  legend: { orient: "top", labelFontSize: 11 },
};

type SceneItem = {
  role?: string;
  items?: SceneItem[];
  bounds?: { x1: number; y1: number; x2: number; y2: number };
  text?: unknown;
  limit?: number;
  x?: number;
  y?: number;
};

function sceneItems(view: vega.View, role: string): SceneItem[] {
  const found: SceneItem[] = [];
  const visit = (node: SceneItem) => {
    if (node.role === role) found.push(...(node.items ?? []));
    for (const item of node.items ?? []) item.items?.forEach(visit);
  };
  visit((view.scenegraph() as unknown as { root: SceneItem }).root);
  return found;
}

/**
 * A chart's legends as Vega drew them in a container of a size, laid out by `fitLegends` under a
 * config. The chart is told the container's size as `VegaLiteRenderer` tells it, unless `told`
 * is false.
 */
async function drawLegends(
  spec: VisualizationSpec,
  size: { width: number; height: number },
  config: Config = theme,
  datasets: Record<string, unknown> = {},
  told = true,
) {
  stubCanvasContext();
  const fitted = fitLegends(spec, datasets, config) as TopLevelSpec;
  const view = new vega.View(vega.parse(compile(fitted, { config }).spec), {
    renderer: "none",
  });
  for (const [name, rows] of Object.entries(datasets)) {
    view.change(name, vega.changeset().insert(rows));
  }
  refitOnPlotResize(view);
  if (told) setContainerSize(view, size.width, size.height);
  view.width(size.width).height(size.height);
  await view.runAsync();
  return { view, ...readLegends(view) };
}

function readLegends(view: vega.View) {
  const legends = sceneItems(view, "legend");
  const labels = sceneItems(view, "legend-label");
  // Each entry is a group placed on the legend's grid.
  const entries = sceneItems(view, "scope");
  return {
    legends,
    labels: labels.map((label) => String(label.text)),
    limits: [...new Set(labels.map((label) => label.limit))],
    rows: new Set(entries.map((entry) => entry.y)).size,
    columns: new Set(entries.map((entry) => entry.x)).size,
    /** How far the legends run past the plot's right edge, and above the view's top. */
    overhang: {
      right: Math.max(...legends.map((l) => l.bounds?.x2 ?? 0)) - view.width(),
      top:
        -Math.min(...legends.map((l) => l.bounds?.y1 ?? 0)) - view.origin()[1],
    },
  };
}

// jsdom has no canvas, so Vega estimates text at 0.8em a character where this module estimates
// 0.6em. Both are given the same measure here, so that the layout is checked against the text
// Vega lays out. A label is no wider than its limit: Vega cuts its text to it.
const textMetrics = (
  vega as unknown as {
    textMetrics: { width: (item: SceneItem, text: string) => number };
  }
).textMetrics;
const estimateWidth = textMetrics.width;
beforeAll(() => {
  textMetrics.width = (item, text) =>
    Math.min(
      text.length * 11 * 0.6,
      item.limit && item.limit > 0 ? item.limit : Infinity,
    );
});
afterAll(() => {
  textMetrics.width = estimateWidth;
});

describe("fitLegends", () => {
  it("wraps a legend above the plot to the plot's width", async () => {
    const legend = await drawLegends(barChart(segments), {
      width: 560,
      height: 300,
    });

    // Left to Vega the twelve names run in one row, some 2,000px long.
    expect(legend.overhang.right).toBeLessThanOrEqual(0);
    expect(legend.overhang.top).toBeLessThanOrEqual(0);
    expect(legend.columns).toBeGreaterThan(1);
    expect(legend.labels).toHaveLength(segments.length);
  });

  it("takes no more than a third of the chart's height", async () => {
    const legend = await drawLegends(barChart(segments), {
      width: 560,
      height: 200,
    });

    // Three rows fit under the title in a third of 200px. Every place on the grid is used, the
    // last by the count of what is left out.
    expect(legend.rows).toBe(3);
    expect(legend.labels).toHaveLength(3 * legend.columns);
    expect(legend.labels.at(-1)).toBe(
      `…${segments.length - legend.labels.length + 1} entries`,
    );
    const { y1, y2 } = legend.legends[0].bounds ?? { y1: 0, y2: 0 };
    expect(y2 - y1).toBeLessThanOrEqual(200 / 3);
    expect(legend.overhang.top).toBeLessThanOrEqual(0);

    // A taller chart has the room to name them all.
    const tall = await drawLegends(barChart(segments), {
      width: 560,
      height: 400,
    });
    expect(tall.labels).toHaveLength(segments.length);
  });

  it("keeps the plot in a small widget", async () => {
    const legend = await drawLegends(barChart(segments), {
      width: 260,
      height: 250,
    });

    // One column, which lists six entries: the last says how many more there are.
    expect(legend.columns).toBe(1);
    expect(legend.labels).toEqual([...segments.slice(0, 5), "…7 entries"]);
    expect(legend.overhang.right).toBeLessThanOrEqual(0);
    // Listing all twelve took the chart's whole height before, and no plot was drawn.
    expect(legend.view.height()).toBeGreaterThan(80);
  });

  it("names every one of a handful of series, however narrow the plot", async () => {
    const legend = await drawLegends(barChart(channels), {
      width: 150,
      height: 250,
    });

    expect(legend.columns).toBe(1);
    expect(legend.labels).toEqual(channels);
  });

  it("keeps a few short labels on one row", async () => {
    const legend = await drawLegends(barChart(channels), {
      width: 560,
      height: 300,
    });

    expect(legend.labels).toEqual(channels);
    expect(legend.rows).toBe(1);
    expect(legend.overhang.right).toBeLessThanOrEqual(0);
  });

  it("wraps short labels where the plot is narrow", async () => {
    const legend = await drawLegends(barChart(channels), {
      width: 260,
      height: 250,
    });

    expect(legend.labels).toEqual(channels);
    expect(legend.rows).toBeGreaterThan(1);
    expect(legend.overhang.right).toBeLessThanOrEqual(0);
  });

  it("shows labels in full where every entry fits, and cuts them where the legend is crowded", async () => {
    const few = await drawLegends(barChart(segments.slice(0, 3)), {
      width: 560,
      height: 300,
    });
    expect(few.labels).toEqual(segments.slice(0, 3));
    expect(few.limits).toEqual([240]);
    expect(few.overhang.right).toBeLessThanOrEqual(0);

    const many = await drawLegends(barChart(segments), {
      width: 560,
      height: 300,
    });
    expect(many.limits).toEqual([140]);
  });

  it("follows a resize without the spec changing", async () => {
    const legend = await drawLegends(barChart(channels), {
      width: 560,
      height: 300,
    });
    setContainerSize(legend.view, 260, 250);
    legend.view.width(260).height(250);
    await legend.view.runAsync();

    const narrow = readLegends(legend.view);
    expect(narrow.rows).toBeGreaterThan(1);
    expect(narrow.overhang.right).toBeLessThanOrEqual(0);
    expect(narrow.overhang.top).toBeLessThanOrEqual(0);
  });

  it("lists beside the plot what its height holds", async () => {
    const legend = await drawLegends(barChart(segments, { orient: "right" }), {
      width: 560,
      height: 150,
    });

    expect(legend.columns).toBe(1);
    expect(legend.limits).toEqual([140]);
    expect(legend.labels.length).toBeLessThan(segments.length);
    expect(legend.labels.at(-1)).toMatch(/^…\d+ entries$/);
  });

  it("cuts the labels of a legend beside the plot to its share of a narrow chart, once told its size", async () => {
    const chart = barChart(segments, { orient: "right" });
    const size = { width: 260, height: 250 };
    // A chart drawn outside `VegaLiteRenderer` is not told, and keeps Vega's fixed limit: at
    // 140px a label, the legend leaves the plot a sliver of a chart this narrow.
    const untold = await drawLegends(chart, size, theme, {}, false);
    expect(untold.limits).toEqual([140]);

    const told = await drawLegends(chart, size);
    expect(told.limits).toEqual([75]);
    expect(told.view.width()).toBeGreaterThan(untold.view.width() + 60);
  });

  it("takes four rows above the plot until it is told its container's size", async () => {
    const legend = await drawLegends(
      barChart(segments),
      { width: 560, height: 200 },
      theme,
      {},
      false,
    );
    expect(legend.rows).toBe(4);
  });

  it("reads a legend's labels from a named dataset", async () => {
    const spec = {
      ...barChart(channels),
      data: { name: "query1" },
    } as VisualizationSpec;
    const legend = await drawLegends(spec, { width: 560, height: 300 }, theme, {
      query1: rowsOf(channels),
    });

    // Without the labels, each column is taken to be 140px wide and six entries wrap.
    expect(legend.labels).toEqual(channels);
    expect(legend.rows).toBe(1);
  });

  it("lays out a chart whose rows are not objects", async () => {
    // Vega-Lite draws a list of numbers or strings as rows of one field, `data`.
    const spec = {
      width: 560,
      height: 300,
      autosize: { type: "fit", contains: "padding" },
      data: { values: channels },
      mark: "point",
      encoding: {
        x: { field: "data", type: "nominal" },
        color: { field: "data", type: "nominal" },
      },
    } as VisualizationSpec;
    const legend = await drawLegends(spec, { width: 560, height: 300 });

    expect(legend.labels).toEqual(channels);
    expect(legend.overhang.right).toBeLessThanOrEqual(0);
  });

  it("stands several legends on one side one above the other, each wrapped", async () => {
    const spec = {
      ...barChart(channels),
      mark: "point",
      encoding: {
        x: { field: "category", type: "nominal" },
        y: { field: "value", type: "quantitative" },
        color: { field: "series", type: "nominal" },
        size: { field: "rate", type: "quantitative" },
      },
    } as VisualizationSpec;
    const { legends, overhang } = await drawLegends(
      spec,
      { width: 260, height: 300 },
      getRillTheme(),
    );

    expect(legends).toHaveLength(2);
    const [colors, sizes] = legends.map((legend) => legend.bounds);
    // Side by side, the sizes started where the colors ended and ran off the chart.
    expect(sizes?.y1).toBeGreaterThanOrEqual(colors?.y2 ?? Infinity);
    expect(overhang.right).toBeLessThanOrEqual(0);
    expect(overhang.top).toBeLessThanOrEqual(0);
  });

  it("leaves a legend its author laid out", () => {
    const columns = barChart(segments, { columns: 2 });
    expect(fitLegends(columns, {}, theme)).toBe(columns);

    const configured = {
      ...barChart(segments),
      config: { legend: { columns: 3 } },
    } as VisualizationSpec;
    expect(fitLegends(configured, {}, theme)).toBe(configured);
  });

  it("keeps what a legend states for itself", () => {
    const fitted = fitLegends(
      barChart(segments, { labelLimit: 80, title: "Segment" }),
      {},
      theme,
    ) as unknown as {
      encoding: { color: { legend: Record<string, unknown> } };
    };

    expect(fitted.encoding.color.legend).toMatchObject({
      labelLimit: 80,
      title: "Segment",
    });
    expect(fitted.encoding.color.legend.columns).toBeDefined();
  });

  it("does not bring back a legend the chart turned off", () => {
    // Vega-Lite draws a legend that states anything of its own, whatever the config says.
    const hidden = barChart(segments, null);
    expect(fitLegends(hidden, {}, theme)).toBe(hidden);

    const disabled = {
      ...barChart(segments),
      config: { legend: { disable: true } },
    } as VisualizationSpec;
    expect(fitLegends(disabled, {}, theme)).toBe(disabled);

    const plain = barChart(segments);
    expect(fitLegends(plain, {}, { legend: { disable: true } })).toBe(plain);
  });

  it("leaves a chart whose plot is not fitted to its container", () => {
    // Without a width the bars are drawn at Vega-Lite's 20px step: three of them make a plot
    // 60px wide in a widget of any width, which is no measure of the room the legend has.
    const stepped = { ...barChart(segments) } as Record<string, unknown>;
    delete stepped.width;
    expect(fitLegends(stepped as VisualizationSpec, {}, theme)).toBe(stepped);

    // Fitted across only, a chart grows downwards to hold a legend beside it.
    const acrossOnly = {
      ...barChart(segments, { orient: "right" }),
      autosize: { type: "fit-x" },
    } as VisualizationSpec;
    expect(fitLegends(acrossOnly, {}, theme)).toBe(acrossOnly);
  });

  it("passes on a spec that is not an object, for Vega to report", () => {
    // A `vega_spec` holding `42` or `"line"` parses, and reaches the renderer as that value.
    for (const malformed of [42, "line", null]) {
      const spec = malformed as unknown as VisualizationSpec;
      expect(fitLegends(spec, {}, theme)).toBe(spec);
    }
    // So does a layer that is no view.
    const layered = { layer: [null] } as unknown as VisualizationSpec;
    expect(fitLegends(layered, {}, theme)).toBe(layered);
  });

  it("leaves composed charts and gradients alone", () => {
    const composed = {
      hconcat: [barChart(segments), barChart(channels)],
    } as VisualizationSpec;
    expect(fitLegends(composed, {}, theme)).toBe(composed);

    const faceted = barChart(segments) as unknown as {
      encoding: Record<string, unknown>;
    };
    faceted.encoding.column = { field: "category", type: "nominal" };
    expect(fitLegends(faceted as unknown as VisualizationSpec, {}, theme)).toBe(
      faceted,
    );

    const heatmap = {
      width: 560,
      autosize: { type: "fit" },
      data: { values: rowsOf(channels) },
      mark: "rect",
      encoding: {
        x: { field: "category", type: "nominal" },
        y: { field: "series", type: "nominal" },
        color: { field: "value", type: "quantitative" },
      },
    } as VisualizationSpec;
    expect(fitLegends(heatmap, {}, theme)).toBe(heatmap);
  });

  it("returns the same spec for as long as its legends are laid out the same", () => {
    const spec = {
      ...barChart(channels),
      data: { name: "query1" },
    } as VisualizationSpec;
    const first = fitLegends(spec, { query1: rowsOf(channels) }, theme);
    // New rows with the same labels, as a refetch gives.
    expect(fitLegends(spec, { query1: rowsOf(channels) }, theme)).toBe(first);
    expect(fitLegends(spec, { query1: rowsOf(segments) }, theme)).not.toBe(
      first,
    );
  });
});
