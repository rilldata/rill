import * as vega from "vega";
import { compile, type TopLevelSpec } from "vega-lite";
import { beforeAll, describe, expect, it } from "vitest";
import { scaleVegaText } from "./text-scale";
import { getRillTheme } from "./vega-config";

// jsdom has no canvas: without a context, Vega estimates text widths instead
// of measuring them, which is all a headless render needs.
beforeAll(() => {
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    configurable: true,
    value: () => null,
  });
});

// A theme with a primary color keeps getRillTheme from resolving CSS variables,
// which jsdom does not have.
const theme = { primary: "#4f46e5" };

const chart: TopLevelSpec = {
  title: "Orders by region",
  data: {
    values: [
      { region: "North", orders: 28, channel: "Web" },
      { region: "South", orders: 55, channel: "Store" },
    ],
  },
  layer: [
    {
      mark: "bar",
      encoding: {
        x: { field: "region", type: "nominal" },
        y: { field: "orders", type: "quantitative" },
        color: { field: "channel", type: "nominal" },
      },
    },
    {
      mark: { type: "text", dy: -6 },
      encoding: {
        x: { field: "region", type: "nominal" },
        y: { field: "orders", type: "quantitative" },
        text: { field: "orders", type: "quantitative" },
      },
    },
  ],
};

/**
 * Draws the spec with Rill's chart config at a text scale and returns the font
 * sizes Vega draws with, by the role of the text.
 */
async function drawnFontSizes(
  spec: TopLevelSpec,
  textScale: number,
): Promise<Record<string, number[]>> {
  const config = getRillTheme(false, theme, { textScale });
  const scaled = scaleVegaText(spec, textScale);
  const view = new vega.View(vega.parse(compile(scaled, { config }).spec), {
    renderer: "none",
  });
  await view.runAsync();

  type Node = { role?: string; items?: Node[]; fontSize?: number };
  const sizes: Record<string, Set<number>> = {};
  const visit = (mark: Node) => {
    for (const item of mark.items ?? []) {
      if (mark.role && typeof item.fontSize === "number") {
        (sizes[mark.role] ??= new Set()).add(item.fontSize);
      }
      item.items?.forEach(visit);
    }
  };
  visit((view.scenegraph() as unknown as { root: Node }).root);
  return Object.fromEntries(
    Object.entries(sizes).map(([role, values]) => [role, [...values]]),
  );
}

describe("chart text scale", () => {
  it("draws every kind of chart text at the scaled size", async () => {
    const standard = await drawnFontSizes(chart, 1);
    const large = await drawnFontSizes(chart, 1.25);

    // The sizes of Rill's chart config, and Vega's own for the roles it leaves unset.
    expect(standard).toEqual({
      "axis-label": [11],
      "axis-title": [12],
      "legend-label": [11],
      "legend-title": [12],
      "title-text": [13],
      mark: [11],
    });
    expect(large).toEqual({
      "axis-label": [13.75],
      "axis-title": [15],
      "legend-label": [13.75],
      "legend-title": [15],
      "title-text": [16.25],
      mark: [13.75],
    });
  });

  it("scales the sizes a spec sets itself, over the config's", async () => {
    const spec: TopLevelSpec = {
      ...chart,
      config: {
        axisX: { labelFontSize: 14 },
        axisY: { labelFontSize: 14 },
      },
      layer: [
        chart.layer[0],
        { ...chart.layer[1], mark: { type: "text", fontSize: 20 } },
      ],
    };

    const sizes = await drawnFontSizes(spec, 1.5);

    expect(sizes["axis-label"]).toEqual([21]);
    expect(sizes.mark).toEqual([30]);
  });
});

describe("scaleVegaText", () => {
  it("returns the input itself at a scale of 1", () => {
    expect(scaleVegaText(chart, 1)).toBe(chart);
  });

  it("leaves a size computed from the chart's dimensions to the spec", () => {
    const spec = {
      mark: {
        type: "text",
        fontSize: { expr: "max(9, min(16, min(width, height) / 22))" },
      },
    };

    expect(scaleVegaText(spec, 1.5).mark.fontSize).toEqual(spec.mark.fontSize);
  });

  it("scales both branches of a conditional size", () => {
    const config = {
      axis: {
        labelFontSize: {
          condition: { test: "datum.index === 0", value: 14 },
          value: 10,
        },
      },
    };

    expect(scaleVegaText(config, 1.5).axis.labelFontSize).toEqual({
      condition: { test: "datum.index === 0", value: 21 },
      value: 15,
    });
  });

  it("leaves the data alone, where a font size key is a column", () => {
    const spec = {
      data: { values: [{ fontSize: 10, label: "a" }] },
      datasets: { sizes: [{ titleFontSize: 4 }] },
      mark: { type: "text", fontSize: 10 },
    };

    const scaled = scaleVegaText(spec, 2);

    expect(scaled.data).toBe(spec.data);
    expect(scaled.datasets).toBe(spec.datasets);
    expect(scaled.mark.fontSize).toBe(20);
  });
});
