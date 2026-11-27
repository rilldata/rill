import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import type { TopLevelSpec } from "vega-lite";
import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "./CartesianChartProvider";
import { generateVLMultiMetricChartSpec } from "./multi-metric-chart";
import { at, chartData, stubCanvasContext } from "./test-fixtures";

const base: CartesianChartSpec = {
  metrics_view: "reddit_posts",
  color: { field: "rill_measures", type: "value" },
  x: { field: "post_title", type: "nominal", sort: "-y" },
  y: {
    field: "post_count",
    fields: ["post_count", "comment_count"],
    type: "quantitative",
    zeroBasedOrigin: true,
  },
};

// The same chart with the measures on x and the dimension on y.
const horizontal: CartesianChartSpec = {
  ...base,
  x: {
    field: "post_count",
    fields: ["post_count", "comment_count"],
    type: "quantitative",
    zeroBasedOrigin: true,
  },
  y: { field: "post_title", type: "nominal", sort: "-x" },
};

describe("generateVLMultiMetricChartSpec orientation", () => {
  it("draws grouped measures along yOffset when the measures are on x", () => {
    const spec = generateVLMultiMetricChartSpec(
      horizontal,
      chartData(),
      "grouped_bar",
    );

    expect(at(spec, "encoding.y.field")).toBe("post_title");
    expect(at(spec, "encoding.x")).toBeUndefined();

    expect(at(spec, "layer.1.encoding.x.field")).toBe("value");
    expect(at(spec, "layer.1.encoding.y")).toBeUndefined();
    expect(at(spec, "layer.1.encoding.yOffset.field")).toBe("Measure");
    expect(at(spec, "layer.1.encoding.xOffset")).toBeUndefined();
    // Grouped bars fill their slot; the gap is the offset scale's padding.
    expect(at(spec, "layer.1.mark.height")).toEqual({ band: 1 });
    expect(at(spec, "layer.1.encoding.yOffset.scale")).toEqual({
      paddingInner: 0.1,
      paddingOuter: 0.05,
    });
    expect(at(spec, "layer.0.params.0.select.encodings")).toEqual(["y"]);
  });

  it("stacks measures along x and offsets comparison periods on yOffset", () => {
    const spec = generateVLMultiMetricChartSpec(
      horizontal,
      chartData({ hasComparison: true }),
      "stacked_bar",
    );

    expect(at(spec, "layer.1.encoding.x.stack")).toBe("zero");
    expect(at(spec, "layer.1.encoding.yOffset.field")).toBe("period");
  });

  it("puts the normalized axis on x", () => {
    const spec = generateVLMultiMetricChartSpec(
      horizontal,
      chartData(),
      "stacked_bar_normalized",
    );

    expect(at(spec, "layer.1.encoding.x.stack")).toBe("normalize");
    expect(at(spec, "layer.1.encoding.x.axis.format")).toBe(".0%");
  });

  it("renders the same vertical layout for the line and area variants", () => {
    // A measure on x is rejected for line and area charts upstream; the
    // builder still reads the fields by role so nothing breaks.
    for (const markType of ["line", "stacked_area"] as const) {
      expect(
        generateVLMultiMetricChartSpec(horizontal, chartData(), markType),
      ).toEqual(generateVLMultiMetricChartSpec(base, chartData(), markType));
    }
  });

  it("compiles to a valid Vega spec", async () => {
    stubCanvasContext();
    const { compile } = await import("vega-lite");

    for (const markType of [
      "grouped_bar",
      "stacked_bar",
      "stacked_bar_normalized",
    ] as const) {
      const spec = generateVLMultiMetricChartSpec(
        horizontal,
        chartData({ hasComparison: true }),
        markType,
      );
      expect(() => compile(spec as TopLevelSpec)).not.toThrow();
    }
  });
});

describe("generateVLMultiMetricChartSpec with dotted measure names", () => {
  it("escapes the folded measures and keeps the raw names as output columns", () => {
    const measures = ["post.count", "comment.count"];
    const data = chartData();
    for (const measure of measures) {
      data.fields[measure] = { name: measure, displayName: measure };
    }

    const spec = generateVLMultiMetricChartSpec(
      {
        ...base,
        y: {
          field: measures[0],
          fields: measures,
          type: "quantitative",
          zeroBasedOrigin: true,
        },
      },
      data,
      "grouped_bar",
    );
    expect(at(spec, "transform.0.fold")).toEqual([
      "post\\.count",
      "comment\\.count",
    ]);
    expect(at(spec, "transform.0.as")).toEqual(["Measure", "value"]);
  });
});

describe("generateVLMultiMetricChartSpec value axis format", () => {
  type MeasureFormat = Pick<
    MetricsViewSpecMeasure,
    "formatPreset" | "formatD3"
  >;
  const measures = ["click_share", "revenue_share"];

  const vertical: CartesianChartSpec = {
    metrics_view: "store_sales",
    color: { field: "rill_measures", type: "value" },
    x: { field: "device", type: "nominal" },
    y: {
      field: measures[0],
      fields: measures,
      type: "quantitative",
      zeroBasedOrigin: true,
    },
  };

  // Horizontal bars carry the measures, and their axis, on x.
  const horizontalBars: CartesianChartSpec = {
    ...vertical,
    x: vertical.y,
    y: vertical.x,
  };

  const withVegaConfig = (config: CartesianChartSpec, vegaConfig: object) => ({
    ...config,
    vl_config: JSON.stringify(vegaConfig),
  });

  function dataWith(formats: MeasureFormat[]) {
    const data = chartData();
    data.fields = { device: { name: "device", displayName: "Device" } };
    measures.forEach((name, i) => {
      data.fields[name] = { name, displayName: name, ...formats[i] };
    });
    data.domainValues = {};
    return data;
  }

  const percentage: MeasureFormat = { formatPreset: "percentage" };

  it.each<[string, MeasureFormat[], boolean]>([
    ["two percentage presets", [percentage, percentage], true],
    [
      "a percentage preset and a d3 percent format",
      [percentage, { formatD3: ".1%" }],
      true,
    ],
    [
      "a percentage preset and a d3 rounded percent format",
      [{ formatD3: ".2p" }, percentage],
      true,
    ],
    [
      "a d3 percent format that overrides another preset",
      [{ formatD3: ".2%", formatPreset: "currency_usd" }, percentage],
      true,
    ],
    [
      "a plain d3 format that overrides a percentage preset",
      [{ formatD3: ",.2f", formatPreset: "percentage" }, percentage],
      false,
    ],
    [
      "a d3 format that cannot be parsed",
      [{ formatD3: "not a format", formatPreset: "percentage" }, percentage],
      false,
    ],
    ["a percentage and a plain number", [percentage, {}], false],
    [
      "a percentage and a currency",
      [percentage, { formatPreset: "currency_usd" }],
      false,
    ],
    // Unset strings arrive as "" from the runtime's REST API.
    ["unformatted measures", [{ formatPreset: "", formatD3: "" }, {}], false],
  ])("formats the axis of %s", (_name, formats, isPercent) => {
    const spec = generateVLMultiMetricChartSpec(vertical, dataWith(formats));
    const axis = at(spec, "layer.1.encoding.y.axis") as Record<string, unknown>;
    expect(axis.format).toBe(isPercent ? ",%" : undefined);
    expect("labelExpr" in axis).toBe(!isPercent);
  });

  it("applies to every variant that draws the measures on one axis", () => {
    const data = dataWith([percentage, percentage]);
    const axisFormat = (
      config: CartesianChartSpec,
      markType: Parameters<typeof generateVLMultiMetricChartSpec>[2],
      path: string,
    ) => at(generateVLMultiMetricChartSpec(config, data, markType), path);

    expect(
      axisFormat(vertical, "stacked_bar", "layer.1.encoding.y.axis.format"),
    ).toBe(",%");
    expect(axisFormat(vertical, "line", "layer.0.encoding.y.axis.format")).toBe(
      ",%",
    );
    expect(
      axisFormat(vertical, "stacked_area", "layer.0.encoding.y.axis.format"),
    ).toBe(",%");

    expect(
      axisFormat(
        horizontalBars,
        "grouped_bar",
        "layer.1.encoding.x.axis.format",
      ),
    ).toBe(",%");

    // A normalized stack shows each measure's share of the total, not its values.
    expect(
      axisFormat(vertical, "stacked_bar_normalized", "layer.1.encoding.y.axis"),
    ).toEqual({ title: null, format: ".0%" });
  });

  // A chart's own Vega-Lite config, as set through `vl_config`.
  const suffix = { labelExpr: "datum.label + ' sec'" };
  it.each<[string, CartesianChartSpec, object, boolean]>([
    [
      "yields to a label expression for the y axis",
      vertical,
      { axisY: suffix },
      false,
    ],
    [
      "yields to a format for quantitative axes",
      vertical,
      { axisQuantitative: { format: ".1%" } },
      false,
    ],
    ["yields to a number format", vertical, { numberFormat: ".2f" }, false],
    [
      "yields to a label expression for the x axis of horizontal bars",
      horizontalBars,
      { axisX: suffix },
      false,
    ],
    [
      "ignores a label expression for the category axis",
      vertical,
      { axisX: suffix },
      true,
    ],
    [
      "ignores a label expression for the category axis of horizontal bars",
      horizontalBars,
      { axisY: suffix },
      true,
    ],
    [
      "ignores axis settings other than the label text",
      vertical,
      { axisY: { labelColor: "gray" } },
      true,
    ],
  ])("%s in `vl_config`", (_name, chart, vegaConfig, isFormatted) => {
    const spec = generateVLMultiMetricChartSpec(
      withVegaConfig(chart, vegaConfig),
      dataWith([percentage, percentage]),
    );
    const channel = chart === horizontalBars ? "x" : "y";
    expect(at(spec, `layer.1.encoding.${channel}.axis.format`)).toBe(
      isFormatted ? ",%" : undefined,
    );
  });

  describe("rendered tick labels", () => {
    /**
     * Renders the chart with Rill's Vega theme, whose number format is what
     * an axis without a format of its own falls back to, and returns the
     * labels of the value axis.
     */
    async function tickLabels(
      formats: MeasureFormat[],
      rows: Record<string, unknown>[],
      config: CartesianChartSpec = vertical,
      markType: Parameters<
        typeof generateVLMultiMetricChartSpec
      >[2] = "grouped_bar",
    ): Promise<string[]> {
      stubCanvasContext();
      const { compile } = await import("vega-lite");
      const vega = await import("vega");
      // The app registers a formatter per measure for tooltips (see Chart.svelte).
      for (const name of measures) {
        vega.expressionFunction(sanitizeFieldName(name), (value: unknown) =>
          String(value),
        );
      }

      const spec = generateVLMultiMetricChartSpec(
        config,
        dataWith(formats),
        markType,
      );
      const view = new vega.View(
        vega.parse(
          compile(spec as TopLevelSpec, { config: getRillTheme(false, {}) })
            .spec,
        ),
        { renderer: "none" },
      )
        .width(500)
        .height(300);
      view.data("metrics-view", rows);
      await view.runAsync();

      // The scenegraph nests marks and their items; a mark carries the role, an item the text.
      type SceneNode = { role?: string; text?: unknown; items?: SceneNode[] };
      const devices = new Set(rows.map((row) => row.device));
      const labels: string[] = [];
      const visit = (node: SceneNode) => {
        for (const item of node.items ?? []) {
          if (node.role === "axis-label" && !devices.has(item.text)) {
            labels.push(String(item.text));
          }
          visit(item);
        }
      };
      // At runtime `scenegraph()` returns the Scenegraph wrapper, not the root mark.
      visit((view.scenegraph() as unknown as { root: SceneNode }).root);
      return labels;
    }

    // d3 writes negative numbers with a minus sign, not a hyphen.
    const minus = "\u2212";
    // A line chart whose axis does not start at zero, for ranges far from it.
    const narrow: CartesianChartSpec = {
      ...vertical,
      y: { ...vertical.y!, zeroBasedOrigin: false },
    };
    const row = (device: string, clicks: number, revenue: number) => ({
      device,
      click_share: clicks,
      revenue_share: revenue,
    });

    it("keeps percentages when the values stay under 100%", async () => {
      // The theme's SI format labelled these ticks "0m" to "800m".
      const labels = await tickLabels(
        [percentage, percentage],
        [row("Desktop", 0.9, 0.6), row("Mobile", 0.1, 0.4)],
      );
      expect(labels).toEqual(["0%", "20%", "40%", "60%", "80%"]);
    });

    it("keeps the same format when the data range changes", async () => {
      // Reaching 100% used to switch the labels to "0.0" through "1.0".
      expect(
        await tickLabels([percentage, percentage], [row("Desktop", 1, 1)]),
      ).toEqual(["0%", "20%", "40%", "60%", "80%", "100%"]);

      expect(
        await tickLabels(
          [percentage, percentage],
          [row("Desktop", -0.25, 0.4), row("Mobile", 0.1, -0.05)],
        ),
      ).toEqual([
        `${minus}20%`,
        `${minus}10%`,
        "0%",
        "10%",
        "20%",
        "30%",
        "40%",
      ]);

      expect(
        await tickLabels(
          [percentage, percentage],
          [row("Desktop", 1.8, 3.1), row("Mobile", 0.4, 2.2)],
        ),
      ).toEqual(["0%", "50%", "100%", "150%", "200%", "250%", "300%", "350%"]);

      // The second measure's values have not arrived yet.
      expect(
        await tickLabels(
          [percentage, percentage],
          [{ device: "Desktop", click_share: 0.9 }],
        ),
      ).toEqual(["0%", "20%", "40%", "60%", "80%"]);
    });

    it("picks the precision from the ticks, not from the measures' formats", async () => {
      // A ".0%" value format would label every one of these ticks "0%".
      const rows = [row("Desktop", 0.0009, 0.0004)];
      expect(await tickLabels([{ formatD3: ".0%" }, percentage], rows)).toEqual(
        ["0.00%", "0.02%", "0.04%", "0.06%", "0.08%"],
      );
      expect(
        await tickLabels(
          [{ formatD3: ".0%" }, percentage],
          rows,
          vertical,
          "stacked_bar",
        ),
      ).toEqual([
        "0.00%",
        "0.02%",
        "0.04%",
        "0.06%",
        "0.08%",
        "0.10%",
        "0.12%",
      ]);
    });

    it("writes fractions as plain decimals when the measures are not all percentages", async () => {
      // The theme's SI format labelled these ticks "0m" to "800m" and "0µ" to "800µ".
      expect(
        await tickLabels(
          [{}, percentage],
          [row("Desktop", 0.9, 0.6), row("Mobile", 0.1, 0.4)],
        ),
      ).toEqual(["0", "0.2", "0.4", "0.6", "0.8"]);
      expect(
        await tickLabels([{}, {}], [row("Desktop", 0.0009, 0.0004)]),
      ).toEqual(["0", "0.0002", "0.0004", "0.0006", "0.0008"]);
      expect(await tickLabels([{}, {}], [row("Desktop", -0.4, 0.6)])).toEqual([
        `${minus}0.4`,
        `${minus}0.2`,
        "0",
        "0.2",
        "0.4",
        "0.6",
      ]);

      // Ticks half a millionth apart: six significant digits would repeat labels.
      const labels = await tickLabels(
        [{}, {}],
        [row("Desktop", 0.500001, 0.500003)],
        narrow,
        "line",
      );
      expect(labels.sort()).toEqual([
        "0.500001",
        "0.5000015",
        "0.500002",
        "0.5000025",
        "0.500003",
      ]);
    });

    it("keeps the labels that the chart's `vl_config` asks for", async () => {
      // A label expression in the encoding would drop this suffix.
      expect(
        await tickLabels(
          [{}, {}],
          [row("Desktop", 9000, 6000), row("Mobile", 1000, 4000)],
          withVegaConfig(vertical, { axisY: suffix }),
        ),
      ).toEqual(["0k sec", "2k sec", "4k sec", "6k sec", "8k sec"]);

      expect(
        await tickLabels(
          [percentage, percentage],
          [row("Desktop", 0.9, 0.6), row("Mobile", 0.1, 0.4)],
          withVegaConfig(horizontalBars, { axisX: { format: ".1%" } }),
        ),
      ).toEqual(["0.0%", "20.0%", "40.0%", "60.0%", "80.0%"]);
    });

    it("leaves the theme's labels alone from one upward", async () => {
      expect(
        await tickLabels(
          [{}, {}],
          [row("Desktop", 9000, 6000), row("Mobile", 1000, 4000)],
        ),
      ).toEqual(["0k", "2k", "4k", "6k", "8k"]);

      // A narrow range keeps the precision Vega derives from the tick step.
      expect(
        await tickLabels(
          [{}, {}],
          [row("Desktop", 1001000, 1002000)],
          narrow,
          "line",
        ),
      ).toEqual([
        "1.0010M",
        "1.0012M",
        "1.0014M",
        "1.0016M",
        "1.0018M",
        "1.0020M",
      ]);
    });
  });
});
