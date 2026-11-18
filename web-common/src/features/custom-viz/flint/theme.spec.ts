import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import type {
  V1MetricsViewSpec,
  V1ThemeChartStyle,
} from "@rilldata/web-common/runtime-client";
import * as vega from "vega";
import { compile, type TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import catalog from "../examples/flint-charts-index.json";
import type { GalleryEntry } from "../examples/example-spec";
import { compileFlintSpec } from "./compile";
import { deriveFlintFields } from "./semantic-types";
import { flintTheme } from "./theme";

/** A project theme's variables for one mode, as the renderers receive them. */
const theme = {
  primary: "#0f62fe",
  "surface-card": "hsl(0 0% 9%)",
  "color-qualitative-1": "#0f62fe",
  "color-qualitative-2": "rgb(238, 83, 150)",
  "color-qualitative-10": "#08bdba",
  "color-sequential-1": "#edf5ff",
  "color-sequential-2": "#0043ce",
};

const style: V1ThemeChartStyle = {
  fontFamily: "IBM Plex Sans, sans-serif",
  text: {
    axisLabel: { size: 12, weight: 500, color: "slategray" },
    legendTitle: { weight: 700 },
  },
  grid: { category: true, style: "dashed", width: 2, color: "#e2e8f0" },
  axis: { line: true, lineWidth: 2, color: "#94a3b8" },
  line: { width: 3 },
  area: { fillOpacity: 0.3, lineWidth: 0 },
  bar: { cornerRadius: 4 },
  legend: { position: "bottom", symbol: "line" },
};

describe("flintTheme", () => {
  it("is absent when the project theme sets no chart style", () => {
    expect(flintTheme(false, theme, {})).toBeUndefined();
  });

  it("states the chart style in Flint's terms", () => {
    const flint = flintTheme(false, theme, { chart: style });

    expect(flint?.type?.axisLabel).toEqual({
      family: "IBM Plex Sans, sans-serif",
      size: 12,
      weight: 500,
      // Flint reads hex colors only.
      color: "#708090",
    });
    expect(flint?.type?.keyTitle?.weight).toBe(700);
    expect(flint?.structure?.grid).toEqual({
      measure: { presence: "full" },
      category: { presence: "full" },
      dash: [3, 3],
      weight: 2,
    });
    expect(flint?.structure?.axis?.measure).toEqual({
      line: "full",
      lineWeight: 2,
      ticks: "omit",
    });
    expect(flint?.structure?.axis?.categorical).toEqual(
      flint?.structure?.axis?.measure,
    );
    expect(flint?.ink?.structure).toEqual({ grid: "#e2e8f0", axis: "#94a3b8" });
    expect(flint?.marks).toEqual({
      strokeWeight: 3,
      fillOpacity: 0.3,
      areaEdge: { presence: "omit", width: 0 },
      cornerRadius: 4,
    });
    expect(flint?.legend).toEqual({ placement: ["bottom"], swatch: "line" });
  });

  it("completes the style with Rill's defaults, so Flint adds no look of its own", () => {
    const flint = flintTheme(false, theme, { chart: { line: { width: 3 } } });

    expect(flint?.marks?.strokeWeight).toBe(3);
    expect(flint?.marks?.fillOpacity).toBe(0.8);
    expect(flint?.type?.axisLabel).toMatchObject({
      family: "Inter, sans-serif",
      size: 11,
      weight: 400,
    });
    expect(flint?.type?.axisTitle).toMatchObject({ size: 12, weight: 500 });
    // Rill draws the gridlines values are read off, dotted, and no axis lines or ticks.
    expect(flint?.structure?.grid).toMatchObject({
      measure: { presence: "full" },
      category: { presence: "omit" },
      dash: [1],
    });
    expect(flint?.structure?.axis?.measure).toEqual({
      line: "omit",
      ticks: "omit",
    });
    expect(flint?.legend?.placement).toEqual(["top"]);
  });

  it("sets the chart in the theme's font unless the chart style names its own", () => {
    const themeFont = { fontFamily: "Lato", chart: { line: { width: 3 } } };
    expect(flintTheme(false, theme, themeFont)?.type?.axisLabel?.family).toBe(
      "Lato",
    );
    expect(
      flintTheme(false, theme, { fontFamily: "Lato", chart: style })?.type
        ?.axisLabel?.family,
    ).toBe("IBM Plex Sans, sans-serif");
  });

  it("hands Flint the project's palettes and the surface the chart sits on", () => {
    const flint = flintTheme(true, theme, { chart: { line: { width: 3 } } });

    expect(flint?.ink?.surface?.canvas).toBe("#171717");
    expect(flint?.ink?.series?.single).toBe("#0f62fe");
    expect(flint?.ink?.series?.categorical).toEqual([
      "#0f62fe",
      "#ee5396",
      "#08bdba",
    ]);
    expect(flint?.ink?.series?.sequential).toEqual({
      stops: ["#edf5ff", "#0043ce"],
    });
  });

  it("falls back to the surface of the mode when the theme names none", () => {
    const noSurface = Object.fromEntries(
      Object.entries(theme).filter(([name]) => name !== "surface-card"),
    );
    const light = flintTheme(false, noSurface, {
      chart: { line: { width: 3 } },
    });
    const dark = flintTheme(true, noSurface, { chart: { line: { width: 3 } } });

    expect(light?.ink?.surface?.canvas).toBe("#ffffff");
    expect(dark?.ink?.surface?.canvas).toBe("#1b1c21");
  });

  it("puts only what the theme states over a Flint preset", () => {
    const flint = flintTheme(false, theme, {
      chart: {
        extends: "economist",
        grid: { style: "solid" },
        line: { width: 3 },
      },
    });

    // The preset keeps its own type, surface and axes; the project's palettes go over its own.
    expect(flint).toEqual({
      extends: "economist",
      structure: { grid: { style: "solid" } },
      marks: { strokeWeight: 3 },
      ink: {
        series: {
          categorical: ["#0f62fe", "#ee5396", "#08bdba"],
          sequential: { stops: ["#edf5ff", "#0043ce"] },
        },
      },
    });
    expect(
      flintTheme(
        false,
        { primary: "#0f62fe" },
        { chart: { extends: "economist" } },
      ),
    ).toEqual({ extends: "economist" });
  });
});

describe("a Flint chart compiled with the project theme", () => {
  const metricsView: V1MetricsViewSpec = {
    timeDimension: "day",
    measures: [{ name: "revenue", displayName: "Revenue" }],
    dimensions: [{ name: "region", displayName: "Region" }],
  };
  // Two regions, and rows without one: a series Flint does not count.
  const rows = ["2026-01-01", "2026-01-02", "2026-01-03"].flatMap((day, i) =>
    [null, "EMEA", "APAC"].map((region, j) => ({
      day: `${day}T00:00:00Z`,
      region,
      revenue: 10 * (j + 1) + i,
    })),
  );

  function compileChart(
    chartType: string,
    chart: V1ThemeChartStyle | undefined,
    isDarkMode = false,
  ) {
    const { spec, error } = compileFlintSpec({
      spec: {
        chartType,
        encodings: { x: "day", y: "revenue", color: "region" },
      },
      rows,
      fields: deriveFlintFields(["day", "region", "revenue"], metricsView),
      size: { width: 560, height: 320 },
      theme: flintTheme(isDarkMode, theme, { chart }),
    });
    expect(error).toBeNull();
    return spec as Record<string, unknown>;
  }

  type SceneNode = {
    role?: string;
    marktype?: string;
    items?: SceneNode[];
    [property: string]: unknown;
  };

  /**
   * Draws a compiled chart with Rill's config for the same style, as `FlintChartRenderer` does,
   * and returns the distinct values a property takes across the items of a role or mark type.
   */
  async function drawn(
    spec: Record<string, unknown>,
    chart: V1ThemeChartStyle | undefined,
    key: string,
    property: string,
  ): Promise<unknown[]> {
    const config = getRillTheme(false, theme, { chart });
    const view = new vega.View(
      vega.parse(compile(spec as unknown as TopLevelSpec, { config }).spec),
      { renderer: "none" },
    );
    await view.runAsync();
    const found: SceneNode[] = [];
    const visit = (mark: SceneNode) => {
      const name = mark.role === "mark" ? `mark-${mark.marktype}` : mark.role;
      for (const item of mark.items ?? []) {
        if (name === key) found.push(item);
        item.items?.forEach(visit);
      }
    };
    visit((view.scenegraph() as unknown as { root: SceneNode }).root);
    return [...new Set(found.map((item) => JSON.stringify(item[property])))]
      .filter((value) => value !== undefined)
      .map((value) => JSON.parse(value) as unknown);
  }

  it("draws a line chart in the theme's style", async () => {
    const spec = compileChart("Line Chart", style);

    expect(await drawn(spec, style, "mark-line", "strokeWidth")).toEqual([3]);
    expect(await drawn(spec, style, "axis-grid", "strokeDash")).toEqual([
      [3, 3],
    ]);
    // Flint leaves out the gridline that the axis line is drawn over.
    const gridColors = await drawn(spec, style, "axis-grid", "stroke");
    expect(gridColors.filter((color) => color !== "transparent")).toEqual([
      "#e2e8f0",
    ]);
    // The config's exact text size stands, where Flint scales its own with the chart's width.
    expect(await drawn(spec, style, "axis-label", "fontSize")).toEqual([12]);
    expect(await drawn(spec, style, "axis-label", "font")).toEqual([
      "IBM Plex Sans, sans-serif",
    ]);
    expect(await drawn(spec, style, "legend", "orient")).toEqual(["bottom"]);
  });

  it("draws an area chart's fill and top line in the theme's style", async () => {
    const edged = { ...style, area: { fillOpacity: 0.3, lineWidth: 2 } };
    const spec = compileChart("Area Chart", edged);

    expect(await drawn(spec, edged, "mark-area", "fillOpacity")).toEqual([0.3]);
    expect(await drawn(spec, edged, "mark-line", "strokeWidth")).toEqual([2]);
  });

  it("stays transparent on the card it is drawn on", () => {
    const spec = compileChart("Line Chart", style, true);

    expect(spec.background).toBeUndefined();
    expect((spec.config as Record<string, unknown>).background).toBeUndefined();
    expect(spec.padding).toBeUndefined();
  });

  it("gives every series its own color, including one Flint does not count", async () => {
    const spec = compileChart("Line Chart", style);

    // Three lines in the theme's three qualitative colors. Flint counts two series and hands
    // the scale two inks, which would give the third line the first color again.
    expect(await drawn(spec, style, "mark-line", "stroke")).toEqual([
      "#0f62fe",
      "rgb(238, 83, 150)",
      "#08bdba",
    ]);
  });

  it("draws and colors that series in a stack", async () => {
    const spec = compileChart("Area Chart", style);

    // Flint lists the two series it counts as the color scale's domain, which leaves the
    // third without a fill.
    expect(await drawn(spec, style, "mark-area", "fill")).toEqual([
      "#0f62fe",
      "rgb(238, 83, 150)",
      "#08bdba",
    ]);
  });

  it("colors that series from a preset's own palette", async () => {
    const preset = { extends: "economist" };
    const economist = ["#006ba2", "#3ebcd2", "#ebb434"];

    for (const [chartType, mark, property] of [
      ["Line Chart", "mark-line", "stroke"],
      ["Area Chart", "mark-area", "fill"],
    ]) {
      const { spec, error } = compileFlintSpec({
        spec: {
          chartType,
          encodings: { x: "day", y: "revenue", color: "region" },
        },
        rows,
        fields: deriveFlintFields(["day", "region", "revenue"], metricsView),
        size: { width: 560, height: 320 },
        // A project without palettes of its own, so the preset's is drawn.
        theme: flintTheme(false, undefined, { chart: preset }),
      });
      expect(error).toBeNull();
      expect(
        await drawn(spec as Record<string, unknown>, preset, mark, property),
        chartType,
      ).toEqual(economist);
    }
  });

  it("compiles as before for a project without a chart style", () => {
    const spec = compileChart("Line Chart", undefined);
    const themed = compileChart("Line Chart", { line: { width: 3 } });

    expect(
      (spec.config as Record<string, Record<string, unknown>>).line,
    ).toBeUndefined();
    expect(
      (themed.config as Record<string, Record<string, unknown>>).line
        .strokeWidth,
    ).toBe(3);
  });

  it("keeps a preset's own design", () => {
    const preset = { extends: "economist", line: { width: 4 } };
    const config = compileChart("Line Chart", preset).config as Record<
      string,
      Record<string, unknown>
    >;

    // The preset's typography and surface are kept, where a chart without one defers to Rill's
    // config and is drawn on its card.
    expect(config.axisX.labelFont).toBeDefined();
    expect(config.background).toBeDefined();
    expect(config.line.strokeWidth).toBe(4);
    const plain = compileChart("Line Chart", style).config as Record<
      string,
      Record<string, unknown>
    >;
    expect(plain.axisX.labelFont).toBeUndefined();
  });

  it("reports a preset Flint does not ship", () => {
    const { spec, error } = compileFlintSpec({
      spec: { chartType: "Line Chart", encodings: { x: "day", y: "revenue" } },
      rows,
      fields: deriveFlintFields(["day", "region", "revenue"], metricsView),
      theme: flintTheme(false, theme, { chart: { extends: "econmist" } }),
    });

    expect(spec).toBeNull();
    expect(error).toContain("Unknown theme `econmist`");
  });
});

describe("the chart catalog compiled with a theme", () => {
  // A metrics view with enough fields for any example: each binding takes the next field of
  // its role.
  const metricsView: V1MetricsViewSpec = {
    timeDimension: "t",
    measures: ["m1", "m2", "m3", "m4", "m5"].map((name) => ({ name })),
    dimensions: ["d1", "d2", "d3"].map((name) => ({ name })),
  };
  const rows = Array.from({ length: 8 }, (_, day) =>
    ["A", "B", "C"].map((d1, i) => ({
      t: `2026-01-${String(day + 1).padStart(2, "0")}T00:00:00Z`,
      d1,
      d2: ["X", "Y"][(day + i) % 2],
      d3: `K${day % 3}`,
      m1: 10 + day * (i + 1),
      m2: 40 - day + i * 3,
      m3: 5 + ((day * 7 + i) % 11),
      m4: 20 + ((day * 3 + i * 5) % 13),
      m5: 15 + i,
    })),
  ).flat();
  const fields = deriveFlintFields(Object.keys(rows[0]), metricsView);

  function compileExample(
    entry: GalleryEntry,
    chart: V1ThemeChartStyle | undefined,
    isDarkMode = false,
  ) {
    const next = { measure: 0, dimension: 0 };
    const encodings = Object.fromEntries(
      entry.bindings.map(({ channel, role }) => [
        channel,
        role === "time_dimension"
          ? "t"
          : role === "measure"
            ? `m${++next.measure}`
            : `d${++next.dimension}`,
      ]),
    );
    return compileFlintSpec({
      spec: {
        chartType: entry.chartType,
        encodings,
        chartProperties: entry.chartProperties,
      },
      rows,
      fields,
      size: { width: 560, height: 320 },
      measureFields: metricsView.measures?.map((m) => m.name as string),
      theme: flintTheme(isDarkMode, theme, { chart }),
    });
  }

  const examples = catalog.examples as GalleryEntry[];
  const themes: [string, V1ThemeChartStyle, boolean][] = [
    ["in light mode", style, false],
    ["in dark mode", style, true],
    ["over a preset", { ...style, extends: "economist" }, false],
  ];

  it.each(themes)("compiles every example %s", (_, chart, isDarkMode) => {
    for (const entry of examples) {
      const plain = compileExample(entry, undefined);
      const themed = compileExample(entry, chart, isDarkMode);
      // A theme changes how a chart looks, not whether it can be drawn.
      expect(themed.error, entry.name).toBe(plain.error);
      if (!themed.spec) continue;
      const config = getRillTheme(isDarkMode, theme, { chart });
      expect(
        () => compile(themed.spec as TopLevelSpec, { config }),
        entry.name,
      ).not.toThrow();
    }
  });
});
