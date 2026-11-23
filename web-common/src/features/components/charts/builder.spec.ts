import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import { createPositionEncoding } from "@rilldata/web-common/features/components/charts/builder";
import { generateVLBarChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/bar-chart/spec";
import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import { generateVLLineChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/line-chart/spec";
import { generateVLMultiMetricChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/multi-metric-chart";
import { generateVLStackedBarChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/stacked-bar/default";
import type { ChartSpec } from "@rilldata/web-common/features/components/charts/types";
import chroma from "chroma-js";
import type { VisualizationSpec } from "svelte-vega";
import { splitAccessPath } from "vega-util";
import { parseExpression } from "vega-expression";
import type { TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";

function chartDataForDimensionValue(value: string): ChartDataResult {
  return {
    data: [],
    isFetching: false,
    fields: {
      created_date: {
        field: "created_date",
        displayName: "Time",
      },
      post_count: {
        name: "post_count",
        displayName: "Post Count",
      },
      post_title: {
        name: "post_title",
        displayName: "Post Title",
      },
    },
    domainValues: {
      post_title: [value],
    },
    isDarkMode: false,
    hasComparison: false,
    theme: {
      primary: chroma("#1d4ed8"),
      secondary: chroma("#7c3aed"),
    },
  };
}

describe("chart builder Vega field escaping", () => {
  it("preserves arbitrary flat field names after Vega path escaping", () => {
    const fieldNames = [
      "Alpha\n\n[Beta](https://x.test/p)",
      'Quote "q" \\ path',
      "a[b].c",
    ];

    for (const fieldName of fieldNames) {
      expect(splitAccessPath(sanitizeValueForVega(fieldName))).toEqual([
        fieldName,
      ]);
    }
  });

  it("compiles pivoted tooltip fields for markdown dimension values", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");
    const dimensionValue = "Alpha\n\n[Beta](https://x.test/p)";

    const spec = generateVLLineChartSpec(
      {
        metrics_view: "reddit_posts",
        x: { field: "created_date", type: "temporal" },
        y: { field: "post_count", type: "quantitative" },
        color: {
          field: "post_title",
          type: "nominal",
          values: [dimensionValue],
        },
      },
      chartDataForDimensionValue(dimensionValue),
    );

    const vegaSpec = compile(spec as TopLevelSpec).spec;
    const tooltipExpression = findStringValue(
      vegaSpec,
      (value) => value.includes("timeFormat") && value.includes("Alpha"),
    );

    expect(tooltipExpression).toBeDefined();
    expect(() => parseExpression(tooltipExpression!)).not.toThrow();
    expect(JSON.stringify(vegaSpec)).toContain(
      "Alpha [Beta](https://x.test/p)",
    );
  });
});

describe("axis label layout", () => {
  const stores = ["MMC Online", "STVT San Antonio", "STVT Corpus Christi"];
  const storeData: ChartDataResult = {
    ...chartDataForDimensionValue("unused"),
    fields: { store: { name: "store", displayName: "Store" } },
    domainValues: { store: stores },
  };
  // Without a canvas context the widest label ("STVT Corpus Christi", 19 chars) is estimated at
  // 19 * 6.2 = 117.8px, plus 8px of padding, rounded up.
  const widestLabelPx = 126;
  // Rows for the charts that are drawn below: every store in two regions, with two measures.
  const regions = ["North", "South"];
  const storeRows = stores.flatMap((store, i) =>
    regions.map((region, j) => ({
      store,
      region,
      post_count: 10 * (i + 1) + j,
      comment_count: 4 * (i + 1) + j,
    })),
  );
  const drawnData: ChartDataResult = {
    ...storeData,
    data: storeRows,
    fields: {
      store: { name: "store", displayName: "Store" },
      region: { name: "region", displayName: "Region" },
      post_count: { name: "post_count", displayName: "Post Count" },
      comment_count: { name: "comment_count", displayName: "Comment Count" },
    },
    domainValues: { store: stores, region: regions },
  };

  it("lays out an unset categorical x-axis with resize-aware expressions", () => {
    stubCanvasContext();
    const axis = createPositionEncoding(
      { field: "store", type: "nominal" },
      storeData,
      "x",
    ).axis as Record<string, unknown>;

    expect(axis.labelAngle).toEqual({
      expr: `((width / 3) >= ${widestLabelPx} ? 0 : ((width / 3) >= 16 ? -45 : -90))`,
    });
    // The side of the axis is read when the label is drawn, so it follows `vl_config` too.
    expect(axis.labelAlign).toEqual({
      expr: `(width / 3) >= ${widestLabelPx} ? "center" : (item.mark.group.orient === "top" ? "left" : "right")`,
    });
    expect(axis.labelBaseline).toEqual({
      expr: `(width / 3) >= ${widestLabelPx} ? (item.mark.group.orient === "top" ? "bottom" : "top") : "middle"`,
    });
    expect(axis.labelLimit).toEqual({
      expr: `(width / 3) >= ${widestLabelPx} ? (width / 3) - 4 : max(60, height * 0.5)`,
    });
    expect(axis.labelOverlap).toBe(false);
    for (const value of [
      axis.labelAngle,
      axis.labelAlign,
      axis.labelBaseline,
      axis.labelLimit,
    ]) {
      expect(() =>
        parseExpression((value as { expr: string }).expr),
      ).not.toThrow();
    }
  });

  it("reaches the bar chart's x-axis through its own spec builder", () => {
    stubCanvasContext();
    const spec = generateVLBarChartSpec(
      {
        metrics_view: "stores",
        x: { field: "store", type: "nominal" },
        y: { field: "post_count", type: "quantitative" },
      },
      storeData,
    ) as { encoding: { x: { axis: Record<string, unknown> } } };

    expect(spec.encoding.x.axis.labelAngle).toEqual({
      expr: `((width / 3) >= ${widestLabelPx} ? 0 : ((width / 3) >= 16 ? -45 : -90))`,
    });
    expect(spec.encoding.x.axis.labelOverlap).toBe(false);
  });

  it("truncates an explicit upright angle to the band instead of thinning labels", () => {
    const axis = createPositionEncoding(
      { field: "store", type: "nominal", labelAngle: 0, limit: 2 },
      storeData,
      "x",
    ).axis as Record<string, unknown>;

    expect(axis.labelAngle).toBe(0);
    expect(axis.labelLimit).toEqual({ expr: "(width / 2) - 4" });
    expect(axis.labelOverlap).toBe(false);
  });

  it("caps an explicit rotated angle at a share of the chart height", () => {
    const axis = createPositionEncoding(
      { field: "store", type: "nominal", labelAngle: -90 },
      storeData,
      "x",
    ).axis as Record<string, unknown>;

    expect(axis.labelAngle).toBe(-90);
    expect(axis.labelLimit).toEqual({ expr: "max(60, height * 0.5)" });
  });

  it("counts categories from the data rows when no domain is precomputed", () => {
    const axis = createPositionEncoding(
      { field: "store", type: "nominal", labelAngle: 0 },
      {
        ...storeData,
        domainValues: undefined,
        data: [{ store: "a" }, { store: "b" }, { store: "a" }],
      },
      "x",
    ).axis as Record<string, unknown>;

    expect(axis.labelLimit).toEqual({ expr: "(width / 2) - 4" });
  });

  it("leaves temporal and y-axis labels alone", () => {
    const temporalAxis = createPositionEncoding(
      { field: "created_date", type: "temporal" },
      storeData,
      "x",
    ).axis as Record<string, unknown>;
    const yAxis = createPositionEncoding(
      { field: "store", type: "nominal", labelAngle: 30 },
      storeData,
      "y",
    ).axis as Record<string, unknown>;

    expect(temporalAxis).not.toHaveProperty("labelAngle");
    expect(temporalAxis).not.toHaveProperty("labelLimit");
    expect(yAxis.labelAngle).toBe(30);
    expect(yAxis).not.toHaveProperty("labelLimit");
    expect(yAxis).not.toHaveProperty("labelAlign");
  });

  it("leaves the alignment of an explicit angle to Vega-Lite and only centers a rotated label's line", () => {
    const axisFor = (labelAngle: number) =>
      createPositionEncoding(
        { field: "store", type: "nominal", labelAngle },
        storeData,
        "x",
      ).axis as Record<string, unknown>;

    for (const upright of [0, 180]) {
      expect(axisFor(upright)).not.toHaveProperty("labelAlign");
      expect(axisFor(upright)).not.toHaveProperty("labelBaseline");
    }
    for (const rotated of [-45, -90, 30]) {
      expect(axisFor(rotated)).not.toHaveProperty("labelAlign");
      expect(axisFor(rotated).labelBaseline).toBe("middle");
    }
  });

  // Each chart is drawn by Vega at a width that puts the automatic layout in one of its three
  // states, or with one of the editor's orientation presets.
  describe.each([
    {
      orientation: "auto, upright",
      labelAngle: undefined,
      width: 900,
      angle: 0,
    },
    {
      orientation: "auto, angled",
      labelAngle: undefined,
      width: 260,
      angle: -45,
    },
    {
      orientation: "auto, vertical",
      labelAngle: undefined,
      width: 70,
      angle: -90,
    },
    { orientation: "horizontal", labelAngle: 0, width: 400, angle: 0 },
    { orientation: "angled", labelAngle: -45, width: 400, angle: -45 },
    { orientation: "vertical", labelAngle: -90, width: 400, angle: -90 },
  ])("drawn $orientation", ({ labelAngle, width, angle }) => {
    const base: CartesianChartSpec = {
      metrics_view: "stores",
      x: { field: "store", type: "nominal", labelAngle },
      y: { field: "post_count", type: "quantitative" },
    };
    const byRegion: CartesianChartSpec = {
      ...base,
      color: { field: "region", type: "nominal" },
    };
    const twoMeasures: CartesianChartSpec = {
      ...base,
      color: { field: "rill_measures", type: "value" },
      y: {
        field: "post_count",
        fields: ["post_count", "comment_count"],
        type: "quantitative",
      },
    };

    it.each([
      ["bar", () => generateVLBarChartSpec(base, drawnData)],
      ["grouped bar", () => generateVLBarChartSpec(byRegion, drawnData)],
      ["stacked bar", () => generateVLStackedBarChartSpec(byRegion, drawnData)],
      [
        "bar with two measures",
        () =>
          generateVLMultiMetricChartSpec(twoMeasures, drawnData, "grouped_bar"),
      ],
    ])(
      "puts each label of a %s under the center of its bars",
      async (_, spec) => {
        const { labels, barCenters } = await drawBars(spec(), storeRows, width);

        expect(labels.map((label) => label.category)).toEqual(stores);
        for (const label of labels) {
          const barCenter = barCenters.get(label.category);
          expect(barCenter).toBeDefined();
          expect(((label.angle % 360) + 360) % 360).toBe((angle + 360) % 360);
          // The label is anchored on the middle of its bars.
          expect(label.anchor).toBeCloseTo(barCenter!, 0);
          if (angle === 0) {
            // An upright label extends equally to both sides of that point.
            expect(label.center).toBeCloseTo(barCenter!, 0);
          } else {
            // A rotated label ends there, with the middle of its line on the point.
            expect([label.align, label.baseline]).toEqual(["right", "middle"]);
          }
        }
      },
    );
  });

  // The axis can be moved above the chart from the field or from the widget's `vl_config`.
  describe.each([
    ["the field's `axisOrient`", { axisOrient: "top" as const }, undefined],
    ["`vl_config`", {}, '{"axisX":{"orient":"top"}}'],
  ])("drawn with the axis on top through %s", (_, xOptions, vl_config) => {
    it.each([
      {
        orientation: "auto, upright",
        labelAngle: undefined,
        width: 900,
        angle: 0,
      },
      {
        orientation: "auto, angled",
        labelAngle: undefined,
        width: 260,
        angle: -45,
      },
      { orientation: "angled", labelAngle: -45, width: 400, angle: -45 },
      { orientation: "vertical", labelAngle: -90, width: 400, angle: -90 },
    ])(
      "keeps $orientation labels above the bars they belong to",
      async ({ labelAngle, width, angle }) => {
        const config: ChartSpec = {
          metrics_view: "stores",
          vl_config,
          x: { field: "store", type: "nominal", labelAngle, ...xOptions },
          y: { field: "post_count", type: "quantitative" },
        };
        const spec = generateVLBarChartSpec(config, drawnData);
        const { labels, barCenters } = await drawBars(
          spec,
          storeRows,
          width,
          "top",
        );

        expect(labels.map((label) => label.category)).toEqual(stores);
        for (const label of labels) {
          expect(((label.angle % 360) + 360) % 360).toBe((angle + 360) % 360);
          expect(label.anchor).toBeCloseTo(barCenters.get(label.category)!, 0);
          // Nothing of the label reaches down into the plot, whose top edge is at 0.
          expect(label.bottom).toBeLessThanOrEqual(0);
          if (angle === 0) {
            expect(label.center).toBeCloseTo(
              barCenters.get(label.category)!,
              0,
            );
          } else {
            // A rotated label starts on the point and runs up and away from the axis.
            expect([label.align, label.baseline]).toEqual(["left", "middle"]);
          }
        }
      },
    );
  });

  it("keeps the bar widths that a widget's offset padding gives while centering the group", async () => {
    const config: ChartSpec = {
      metrics_view: "stores",
      vl_config:
        '{"scale":{"offsetBandPaddingInner":0.6,"offsetBandPaddingOuter":0.3}}',
      x: { field: "store", type: "nominal", labelAngle: 0 },
      y: { field: "post_count", type: "quantitative" },
      color: { field: "region", type: "nominal" },
    };
    const spec = generateVLBarChartSpec(config, drawnData);
    const { labels, barCenters, barWidths, bandwidth } = await drawBars(
      spec,
      storeRows,
      600,
    );

    // The two bars of a store share its band. The configured padding leaves each bar's slot 40%
    // of its step, and a bar fills 90% of the slot.
    const step = bandwidth / (regions.length - 0.6 + 2 * 0.3);
    expect(barWidths).toHaveLength(stores.length * regions.length);
    for (const barWidth of barWidths) {
      expect(barWidth).toBeCloseTo(0.9 * 0.4 * step, 1);
    }
    for (const label of labels) {
      expect(label.anchor).toBeCloseTo(barCenters.get(label.category)!, 0);
    }
  });
});

interface SceneItem {
  x: number;
  y: number;
  width: number;
  align?: string;
  baseline?: string;
  angle?: number;
  orient?: string;
  bounds: { x1: number; x2: number; y2: number };
  datum: Record<string, unknown>;
  items?: SceneMark[];
}

interface SceneMark {
  marktype: string;
  role: string;
  name?: string;
  items: SceneItem[];
}

/**
 * Draws a bar chart spec of stores with Vega at the given view width and returns, in the plot's
 * coordinates, where each label of the x-axis on the given side sits, and where the middle of
 * each store's bars is, with the width of every bar and of a store's band.
 */
async function drawBars(
  spec: VisualizationSpec,
  rows: Record<string, unknown>[],
  width: number,
  axisSide: "bottom" | "top" = "bottom",
) {
  stubCanvasContext();
  const { compile } = await import("vega-lite");
  const vega = await import("vega");
  for (const measure of ["post_count", "comment_count"]) {
    vega.expressionFunction(sanitizeFieldName(measure), String);
  }

  const view = new vega.View(vega.parse(compile(spec as TopLevelSpec).spec), {
    renderer: "none",
  });
  view.data("metrics-view", rows);
  await view.runAsync();
  // The spec takes its size from a container. Without one the size is set on the view, which
  // then fits the plot and its axes inside it, as it does on a resize.
  await view.width(width).height(300).runAsync();

  const frame = (view.scenegraph() as unknown as { root: SceneMark }).root
    .items[0];
  const labels: {
    category: string;
    anchor: number;
    center: number;
    bottom: number;
    angle: number;
    align?: string;
    baseline?: string;
  }[] = [];
  const barExtents = new Map<string, [number, number]>();
  const barWidths: number[] = [];

  for (const mark of frame.items ?? []) {
    if (mark.role === "axis") {
      const axis = mark.items[0];
      if (axis.orient !== axisSide) continue;
      const labelMark = axis.items?.find((m) => m.role === "axis-label");
      for (const label of labelMark?.items ?? []) {
        labels.push({
          category: String(label.datum.value),
          anchor: axis.x + label.x,
          center: axis.x + (label.bounds.x1 + label.bounds.x2) / 2,
          bottom: axis.y + label.bounds.y2,
          angle: label.angle ?? 0,
          align: label.align,
          baseline: label.baseline,
        });
      }
    } else if (mark.marktype === "rect" && mark.name !== "layer_0_marks") {
      // `layer_0_marks` is the transparent hover target; the rest are the bars.
      for (const bar of mark.items) {
        const category = String(bar.datum.store);
        barWidths.push(bar.width);
        const [start, end] = barExtents.get(category) ?? [Infinity, -Infinity];
        barExtents.set(category, [
          Math.min(start, bar.x),
          Math.max(end, bar.x + bar.width),
        ]);
      }
    }
  }

  const barCenters = new Map(
    [...barExtents].map(([category, [start, end]]) => [
      category,
      (start + end) / 2,
    ]),
  );
  const bandwidth = (
    view.scale("x") as unknown as { bandwidth: () => number }
  ).bandwidth();
  return { labels, barCenters, barWidths, bandwidth };
}

function findStringValue(
  value: unknown,
  predicate: (value: string) => boolean,
): string | undefined {
  if (typeof value === "string") {
    return predicate(value) ? value : undefined;
  }

  if (!value || typeof value !== "object") return undefined;

  for (const child of Object.values(value)) {
    const match = findStringValue(child, predicate);
    if (match) return match;
  }

  return undefined;
}

function stubCanvasContext() {
  if (typeof HTMLCanvasElement === "undefined") return;
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    configurable: true,
    value: () => null,
  });
}
