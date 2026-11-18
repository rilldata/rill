import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import { stubCanvasContext } from "@rilldata/web-common/features/components/charts/cartesian/test-fixtures";
import type {
  ChartDataResult,
  ChartSpec,
  ChartType,
} from "@rilldata/web-common/features/components/charts/types";
import { generateSpec } from "@rilldata/web-common/features/components/charts/util";
import type {
  V1ThemeChartStyle,
  V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import chroma from "chroma-js";
import * as vega from "vega";
import { compile, type TopLevelSpec } from "vega-lite";
import { beforeAll, describe, expect, it } from "vitest";
import { chartStyleOfVegaConfig, resolveChartTheme } from "./chart-theme";

/**
 * A theme's chart style, as the parser resolves it per mode: the brand font, heavier labels and
 * dashed gridlines in both modes, and each mode's own colors.
 */
const themeSpec: V1ThemeSpec = {
  chart: {
    light: {
      fontFamily: "IBM Plex Sans, sans-serif",
      text: {
        axisLabel: { size: 12, weight: 500, color: "#475569" },
        axisTitle: { size: 13, weight: 600, color: "#0f172a" },
        legendLabel: { size: 12, color: "#334155" },
      },
      grid: { dash: [4, 2], width: 2, color: "#e2e8f0" },
      axis: { line: true, lineWidth: 2, color: "#94a3b8" },
      line: { width: 3 },
      area: { fillOpacity: 0.3, lineWidth: 0 },
      legend: { position: "bottom", symbol: "line" },
    },
    dark: {
      fontFamily: "IBM Plex Sans, sans-serif",
      text: { axisLabel: { size: 12, weight: 500, color: "#cbd5e1" } },
      grid: { style: "dotted", color: "#334155" },
    },
  },
};
const light = resolveChartTheme(themeSpec, false);
const dark = resolveChartTheme(themeSpec, true);

const rows = ["2026-01-01", "2026-01-02", "2026-01-03"].flatMap((day, i) => [
  { day, region: "EMEA", revenue: 10 + i, cost: 5 + i },
  { day, region: "APAC", revenue: 20 + i, cost: 8 + i },
]);

function chartData(chartTheme: V1ThemeChartStyle | undefined): ChartDataResult {
  return {
    data: rows,
    isFetching: false,
    fields: {
      day: { field: "day", displayName: "Day" },
      region: { name: "region", displayName: "Region" },
      revenue: { name: "revenue", displayName: "Revenue" },
      cost: { name: "cost", displayName: "Cost" },
    },
    domainValues: { region: ["EMEA", "APAC"] },
    isDarkMode: false,
    theme: { primary: chroma("#1d4ed8"), secondary: chroma("#7c3aed") },
    chartTheme,
  };
}

const timeSeries = {
  metrics_view: "sales",
  x: { field: "day", type: "temporal" },
  y: { field: "revenue", type: "quantitative", showAxisTitle: true },
  color: { field: "region", type: "nominal" },
} as ChartSpec;

type SceneNode = {
  role?: string;
  marktype?: string;
  items?: SceneNode[];
  [property: string]: unknown;
};

/**
 * Draws a built-in chart the way `Chart.svelte` does, with the spec the chart type generates and
 * Rill's config for the chart style, and returns what Vega drew: the scenegraph items by role
 * (`axis-grid`, `axis-label`, `legend-symbol`, …), and the data marks by mark type.
 */
async function draw(
  chartType: ChartType,
  chartSpec: ChartSpec,
  chartTheme: V1ThemeChartStyle | undefined,
  isDarkMode = false,
) {
  const spec = generateSpec(chartType, chartSpec, chartData(chartTheme));
  return render(
    {
      ...spec,
      data: { values: rows },
      width: 400,
      height: 300,
    } as TopLevelSpec,
    getRillTheme(isDarkMode, undefined, { chart: chartTheme }),
  );
}

async function render(
  spec: TopLevelSpec,
  config: ReturnType<typeof getRillTheme>,
) {
  const view = new vega.View(vega.parse(compile(spec, { config }).spec), {
    renderer: "none",
  });
  await view.runAsync();

  const drawn = new Map<string, SceneNode[]>();
  const visit = (mark: SceneNode) => {
    const key = mark.role === "mark" ? `mark-${mark.marktype}` : mark.role;
    for (const item of mark.items ?? []) {
      if (key) drawn.set(key, [...(drawn.get(key) ?? []), item]);
      item.items?.forEach(visit);
    }
  };
  visit((view.scenegraph() as unknown as { root: SceneNode }).root);
  return (role: string) => drawn.get(role) ?? [];
}

/** The distinct values a property takes across scenegraph items. */
function values(items: SceneNode[], property: string): unknown[] {
  return [...new Set(items.map((item) => JSON.stringify(item[property])))].map(
    (value) => (value === undefined ? undefined : JSON.parse(value)),
  );
}

describe("a theme's chart style on built-in charts", () => {
  beforeAll(() => {
    stubCanvasContext();
    // Built-in charts format measure axes through an expression function per measure.
    for (const measure of ["revenue", "cost"]) {
      vega.expressionFunction(sanitizeFieldName(measure), (value: number) =>
        String(value),
      );
    }
  });

  it("styles the gridlines, axes, text and line of a line chart", async () => {
    const drawn = await draw("line_chart", timeSeries, light);

    const grid = drawn("axis-grid");
    expect(values(grid, "strokeDash")).toEqual([[4, 2]]);
    expect(values(grid, "strokeWidth")).toEqual([2]);
    expect(values(grid, "stroke")).toEqual(["#e2e8f0"]);

    const labels = drawn("axis-label");
    expect(values(labels, "font")).toEqual(["IBM Plex Sans, sans-serif"]);
    expect(values(labels, "fontSize")).toEqual([12]);
    expect(values(labels, "fontWeight")).toEqual([500]);
    expect(values(labels, "fill")).toEqual(["#475569"]);

    const titles = drawn("axis-title");
    expect(values(titles, "fontSize")).toEqual([13]);
    expect(values(titles, "fontWeight")).toEqual([600]);
    expect(values(titles, "fill")).toEqual(["#0f172a"]);

    const domain = drawn("axis-domain");
    expect(domain).toHaveLength(2);
    expect(values(domain, "strokeWidth")).toEqual([2]);
    expect(values(domain, "stroke")).toEqual(["#94a3b8"]);

    expect(values(drawn("mark-line"), "strokeWidth")).toEqual([3]);
  });

  it("keeps Rill's defaults for what the style leaves unset", async () => {
    const themed = await draw("line_chart", timeSeries, {
      line: { width: 3 },
    });
    const plain = await draw("line_chart", timeSeries, undefined);

    expect(values(themed("mark-line"), "strokeWidth")).toEqual([3]);
    expect(values(plain("mark-line"), "strokeWidth")).toEqual([1.5]);
    for (const role of ["axis-grid", "axis-label", "legend-label"]) {
      for (const property of ["strokeDash", "stroke", "fontSize", "fill"]) {
        expect(values(themed(role), property)).toEqual(
          values(plain(role), property),
        );
      }
    }
    expect(themed("axis-domain")).toEqual([]);
    // An empty style is the same as none.
    expect(getRillTheme(false, undefined, { chart: {} })).toEqual(
      getRillTheme(false),
    );
  });

  it("styles the fill and the top line of an area chart", async () => {
    const themed = await draw("area_chart", timeSeries, light);
    const plain = await draw("area_chart", timeSeries, undefined);

    expect(values(themed("mark-area"), "fillOpacity")).toEqual([0.3]);
    expect(values(plain("mark-area"), "fillOpacity")).toEqual([0.8]);
    // A width of 0 takes the line off the area: both the one the area mark draws and the
    // chart's own line layer.
    expect(values(themed("mark-line"), "strokeWidth")).toEqual([0]);
    expect(values(plain("mark-line"), "strokeWidth")).toEqual([1, 1.5]);
  });

  it("lets a widget's vl_config set the width of an area chart's line layer", async () => {
    const own = await draw(
      "area_chart",
      { ...timeSeries, vl_config: '{"line":{"strokeWidth":6}}' } as ChartSpec,
      light,
    );

    // The area mark's own edge keeps the theme's width; the chart's line layer takes the widget's.
    expect(values(own("mark-line"), "strokeWidth").sort()).toEqual([0, 6]);
  });

  it("styles the line layer of an area chart with several measures", async () => {
    const measures = {
      metrics_view: "sales",
      x: { field: "day", type: "temporal" },
      y: {
        field: "revenue",
        type: "quantitative",
        fields: ["revenue", "cost"],
      },
      color: { type: "value", field: "rill_measures" },
    } as ChartSpec;

    const themed = await draw("area_chart", measures, light);
    const plain = await draw("area_chart", measures, undefined);

    expect(values(themed("mark-area"), "fillOpacity")).toEqual([0.3]);
    expect(values(themed("mark-line"), "strokeWidth")).toEqual([0]);
    expect(values(plain("mark-line"), "strokeWidth")).toEqual([1, 1.5]);
  });

  it("styles the line of a combo chart", async () => {
    const combo = {
      metrics_view: "sales",
      x: { field: "day", type: "temporal" },
      y1: { field: "revenue", type: "quantitative", mark: "bar" },
      y2: { field: "cost", type: "quantitative", mark: "line" },
    } as ChartSpec;

    const themed = await draw("combo_chart", combo, light);
    const plain = await draw("combo_chart", combo, undefined);
    const own = await draw(
      "combo_chart",
      { ...combo, vl_config: '{"line":{"strokeWidth":5}}' } as ChartSpec,
      light,
    );

    expect(values(themed("mark-line"), "strokeWidth")).toEqual([3]);
    // A combo chart draws its line heavier than Rill's default line.
    expect(values(plain("mark-line"), "strokeWidth")).toEqual([2]);
    expect(values(own("mark-line"), "strokeWidth")).toEqual([5]);
  });

  it("places and draws the legend, unless the chart places its own", async () => {
    const themed = await draw("line_chart", timeSeries, light);
    expect(values(themed("legend"), "orient")).toEqual(["bottom"]);
    expect(values(themed("legend-symbol"), "shape")).toEqual(["stroke"]);
    expect(values(themed("legend-label"), "fontSize")).toEqual([12]);
    expect(values(themed("legend-label"), "fill")).toEqual(["#334155"]);

    const own = await draw(
      "line_chart",
      {
        ...timeSeries,
        color: { field: "region", type: "nominal", legendOrientation: "right" },
      } as ChartSpec,
      light,
    );
    expect(values(own("legend"), "orient")).toEqual(["right"]);

    const plain = await draw("line_chart", timeSeries, undefined);
    expect(values(plain("legend"), "orient")).toEqual(["top"]);
  });

  it("places the legend where a widget's vl_config says, below its color field", async () => {
    const configured = {
      ...timeSeries,
      vl_config: '{"legend":{"orient":"left"}}',
    } as ChartSpec;

    const themed = await draw("line_chart", configured, light);
    expect(values(themed("legend"), "orient")).toEqual(["left"]);
    const plain = await draw("line_chart", configured, undefined);
    expect(values(plain("legend"), "orient")).toEqual(["left"]);

    const authored = await draw(
      "line_chart",
      {
        ...configured,
        color: { field: "region", type: "nominal", legendOrientation: "right" },
      } as ChartSpec,
      light,
    );
    expect(values(authored("legend"), "orient")).toEqual(["right"]);

    // A widget's vl_config can also place the legend in a corner of the chart.
    const cornered = await draw(
      "line_chart",
      { ...timeSeries, vl_config: '{"legend":{"orient":"top-left"}}' },
      light,
    );
    expect(values(cornered("legend"), "orient")).toEqual(["top-left"]);
  });

  it("applies each mode's own style", async () => {
    const lightGrid = (await draw("line_chart", timeSeries, light))(
      "axis-grid",
    );
    const darkGrid = (await draw("line_chart", timeSeries, dark, true))(
      "axis-grid",
    );

    expect(values(lightGrid, "stroke")).toEqual(["#e2e8f0"]);
    expect(values(darkGrid, "stroke")).toEqual(["#334155"]);
    // The dark style names a grid style instead of a dash pattern.
    expect(values(darkGrid, "strokeDash")).toEqual([[1, 3]]);
  });

  it("draws gridlines by what an axis holds, not where it sits", async () => {
    const category = { field: "region", type: "nominal" };
    const measure = { field: "revenue", type: "quantitative" };
    const bars = {
      metrics_view: "sales",
      x: category,
      y: measure,
    } as ChartSpec;
    const sideways = { ...bars, x: measure, y: category } as ChartSpec;
    const gridOf = async (spec: ChartSpec, grid: V1ThemeChartStyle["grid"]) =>
      (await draw("bar_chart", spec, { grid }))("axis-grid");
    // A gridline of a 400 by 300 chart is horizontal when it spans the width.
    const horizontal = (items: SceneNode[]) =>
      values(items, "x2").map((x2) => (x2 as number) > 0);

    // The measure's gridlines follow the measure to the x axis of a horizontal bar chart.
    const measureOnly = { measure: true, category: false };
    expect(horizontal(await gridOf(bars, measureOnly))).toEqual([true]);
    expect(horizontal(await gridOf(sideways, measureOnly))).toEqual([false]);

    expect(await gridOf(bars, { measure: false })).toEqual([]);
    expect(
      horizontal(await gridOf(bars, { measure: true, category: true })).sort(),
    ).toEqual([false, true]);
  });
});

describe("chartStyleOfVegaConfig", () => {
  it("reads back the style a config was built from", () => {
    const style: V1ThemeChartStyle = {
      fontFamily: "IBM Plex Sans, sans-serif",
      text: {
        title: { size: 16, weight: 700, color: "#0f172a" },
        axisLabel: { size: 12, weight: 500, color: "#475569" },
        axisTitle: { size: 13, weight: 600, color: "#0f172a" },
        legendLabel: { size: 12, weight: 300, color: "#334155" },
        legendTitle: { size: 13, weight: 700, color: "#0f172a" },
        valueLabel: { size: 10, weight: 500, color: "#1e293b" },
      },
      grid: {
        measure: false,
        category: true,
        dash: [4, 2],
        width: 2,
        color: "#e2e8f0",
      },
      axis: { line: true, lineWidth: 2, ticks: true, color: "#94a3b8" },
      line: { width: 3 },
      area: { fillOpacity: 0.3, lineWidth: 0 },
      bar: { opacity: 0.9, cornerRadius: 4 },
      point: { size: 60, opacity: 0.7 },
      legend: { position: "bottom", symbol: "line" },
    };

    expect(
      chartStyleOfVegaConfig(getRillTheme(false, undefined, { chart: style })),
    ).toEqual(style);
  });

  it("states Rill's defaults for a config built from no style", () => {
    expect(chartStyleOfVegaConfig(getRillTheme(false))).toMatchObject({
      fontFamily: "Inter, sans-serif",
      text: {
        axisLabel: { size: 11, weight: 400 },
        axisTitle: { size: 12, weight: 500 },
      },
      grid: { measure: true, category: false, dash: [1] },
      axis: { line: false, ticks: false },
      line: { width: 1.5 },
      area: { fillOpacity: 0.8, lineWidth: 1 },
      legend: { position: "top" },
    });
  });
});

describe("a theme's chart style on a Vega-Lite spec", () => {
  beforeAll(stubCanvasContext);

  const spec: TopLevelSpec = {
    data: { values: rows },
    width: 400,
    height: 300,
    mark: "line",
    encoding: {
      x: { field: "day", type: "temporal" },
      y: { field: "revenue", type: "quantitative" },
      color: { field: "region", type: "nominal" },
    },
  };
  const config = getRillTheme(false, undefined, { chart: light });

  it("styles a spec that states no styling of its own", async () => {
    const drawn = await render(spec, config);

    expect(values(drawn("axis-grid"), "strokeDash")).toEqual([[4, 2]]);
    expect(values(drawn("mark-line"), "strokeWidth")).toEqual([3]);
    expect(values(drawn("legend"), "orient")).toEqual(["bottom"]);
  });

  it("gives way to the spec's own config", async () => {
    const drawn = await render(
      {
        ...spec,
        config: {
          axisY: { gridDash: [8, 8] },
          line: { strokeWidth: 5 },
          legend: { orient: "left" },
        },
      },
      config,
    );

    expect(values(drawn("axis-grid"), "strokeDash")).toEqual([[8, 8]]);
    expect(values(drawn("mark-line"), "strokeWidth")).toEqual([5]);
    expect(values(drawn("legend"), "orient")).toEqual(["left"]);
    // What the spec's config leaves alone still comes from the theme.
    expect(values(drawn("axis-grid"), "stroke")).toEqual(["#e2e8f0"]);
    expect(values(drawn("axis-label"), "fontSize")).toEqual([12]);
  });
});
