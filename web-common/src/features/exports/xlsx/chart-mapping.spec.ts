import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts/types";
import {
  WorkbookChartType,
  WorkbookColumnType,
  WorkbookSeriesMark,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { hasNativeChart, mapChart } from "./chart-mapping";

const fields: ChartDataResult["fields"] = {
  country: { name: "country", displayName: "Country" },
  device: { name: "device", displayName: "Device" },
  impressions: {
    name: "impressions",
    displayName: "Impressions",
    formatPreset: "humanize",
    expression: "count(*)",
  },
  spend: {
    name: "spend",
    displayName: "Spend",
    formatPreset: "currency_usd",
    expression: "sum(spend)",
  },
  cpa: { name: "cpa", displayName: "CPA", formatD3: ".2f", expression: "1" },
  timestamp: { field: "timestamp", displayName: "Time" },
};

function data(
  rows: Record<string, unknown>[],
  extra: Partial<ChartDataResult> = {},
): ChartDataResult {
  return {
    data: rows,
    isFetching: false,
    fields,
    theme: {
      primary: "#000",
      secondary: "#fff",
    } as unknown as ChartDataResult["theme"],
    isDarkMode: false,
    ...extra,
  };
}

describe("mapChart", () => {
  it("maps a vertical bar chart to a column chart over a long table", () => {
    const { table, chart } = mapChart({
      chartType: "bar_chart",
      spec: {
        metrics_view: "mv",
        x: { field: "country", type: "nominal", showAxisTitle: true },
        y: { field: "impressions", type: "quantitative", showAxisTitle: true },
      } as CanvasChartSpec,
      data: data([
        { country: "US", impressions: 10 },
        { country: "CA", impressions: 4 },
      ]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(table.columns.map((c) => c.displayName)).toStrictEqual([
      "Country",
      "Impressions",
    ]);
    expect(table.rows).toStrictEqual([
      ["US", 10],
      ["CA", 4],
    ]);
    expect(chart?.type).toBe(WorkbookChartType.COLUMN);
    expect(chart?.categoryColumn).toBe("country");
    expect(chart?.series.map((s) => s.column)).toStrictEqual(["impressions"]);
    expect(chart?.legendPosition).toBe("none");
    expect(chart?.xAxisTitle).toBe("Country");
    expect(chart?.yAxisTitle).toBe("Impressions");
    expect(chart?.numberFormat).toBe("#,##0");
  });

  it("maps a horizontal bar chart to a bar chart", () => {
    const { chart } = mapChart({
      chartType: "stacked_bar",
      spec: {
        metrics_view: "mv",
        x: { field: "impressions", type: "quantitative" },
        y: { field: "country", type: "nominal" },
      } as CanvasChartSpec,
      data: data([{ country: "US", impressions: 10 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(chart?.type).toBe(WorkbookChartType.BAR);
    expect(chart?.stacked).toBe(true);
    expect(chart?.categoryColumn).toBe("country");
  });

  it("pivots a split chart wide, one series per split value, with typed time categories", () => {
    const { table, chart } = mapChart({
      chartType: "line_chart",
      spec: {
        metrics_view: "mv",
        x: { field: "timestamp", type: "temporal" },
        y: { field: "spend", type: "quantitative" },
        color: { field: "device", type: "nominal", legendOrientation: "top" },
      } as CanvasChartSpec,
      data: data(
        [
          { timestamp: new Date(2026, 0, 1), device: "mobile", spend: 1.5 },
          { timestamp: new Date(2026, 0, 1), device: "desktop", spend: 2 },
          { timestamp: new Date(2026, 0, 2), device: "mobile", spend: 3 },
        ],
        { domainValues: { device: ["desktop", "mobile"] } },
      ),
      timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      timeZone: "UTC",
    });
    expect(table.columns.map((c) => [c.displayName, c.type])).toStrictEqual([
      ["Time", WorkbookColumnType.DATE],
      ["desktop", WorkbookColumnType.NUMBER],
      ["mobile", WorkbookColumnType.NUMBER],
    ]);
    expect(table.rows).toStrictEqual([
      ["2026-01-01T00:00:00", 2, 1.5],
      ["2026-01-02T00:00:00", null, 3],
    ]);
    expect(chart?.type).toBe(WorkbookChartType.LINE);
    expect(chart?.series.map((s) => s.column)).toStrictEqual([
      "spend_0",
      "spend_1",
    ]);
    expect(chart?.legendPosition).toBe("top");
    expect(chart?.numberFormat).toBe(`"$"#,##0.00`);
  });

  it("adds previous-period columns when the chart compares periods", () => {
    const { table, chart } = mapChart({
      chartType: "stacked_bar_normalized",
      spec: {
        metrics_view: "mv",
        x: { field: "country", type: "nominal" },
        y: {
          field: "impressions",
          type: "quantitative",
          fields: ["impressions", "spend"],
        },
      } as CanvasChartSpec,
      data: data(
        [{ country: "US", impressions: 10, spend: 5, impressions_prev: 8 }],
        {
          hasComparison: true,
        },
      ),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(table.columns.map((c) => c.displayName)).toStrictEqual([
      "Country",
      "Impressions",
      "Spend",
      "Impressions (previous)",
    ]);
    expect(table.rows).toStrictEqual([["US", 10, 5, 8]]);
    expect(chart?.percentStacked).toBe(true);
    expect(chart?.legendPosition).toBe("bottom");
  });

  it("maps pie and donut charts", () => {
    const { chart } = mapChart({
      chartType: "donut_chart",
      spec: {
        metrics_view: "mv",
        measure: { field: "impressions", type: "quantitative" },
        color: { field: "country", type: "nominal" },
      } as CanvasChartSpec,
      data: data([{ country: "US", impressions: 10 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(chart?.type).toBe(WorkbookChartType.DOUGHNUT);
    expect(chart?.categoryColumn).toBe("country");
    expect(chart?.legendPosition).toBe("right");
  });

  it("maps a combo chart with the second measure on a secondary line axis", () => {
    const { table, chart } = mapChart({
      chartType: "combo_chart",
      spec: {
        metrics_view: "mv",
        x: { field: "country", type: "nominal" },
        y1: { field: "spend", type: "quantitative", mark: "bar" },
        y2: { field: "cpa", type: "quantitative", mark: "line" },
      } as CanvasChartSpec,
      data: data([{ country: "US", spend: 10, cpa: 2.5 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(table.columns.map((c) => c.name)).toStrictEqual([
      "country",
      "spend",
      "cpa",
    ]);
    expect(chart?.type).toBe(WorkbookChartType.COMBO);
    expect(
      chart?.series.map((s) => [s.column, s.mark, s.secondaryAxis]),
    ).toStrictEqual([
      ["spend", WorkbookSeriesMark.BAR, false],
      ["cpa", WorkbookSeriesMark.LINE, true],
    ]);
  });

  it("keeps a combo chart native when its colour is the measures pseudo-field", () => {
    const { table, chart } = mapChart({
      chartType: "combo_chart",
      spec: {
        metrics_view: "mv",
        x: { field: "country", type: "nominal" },
        y1: { field: "spend", type: "quantitative", mark: "bar" },
        y2: { field: "cpa", type: "quantitative", mark: "line" },
        color: { field: "measures", type: "value", legendOrientation: "top" },
      } as unknown as CanvasChartSpec,
      data: data([{ country: "US", spend: 10, cpa: 2.5 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(table.columns.map((c) => c.name)).toStrictEqual([
      "country",
      "spend",
      "cpa",
    ]);
    expect(chart?.type).toBe(WorkbookChartType.COMBO);
    expect(chart?.legendPosition).toBe("top");
  });

  it("moves long legends to the right", () => {
    const rows = Array.from({ length: 8 }, (_, i) => ({
      country: "US",
      device: `d${i}`,
      spend: i,
    }));
    const { chart } = mapChart({
      chartType: "stacked_bar",
      spec: {
        metrics_view: "mv",
        x: { field: "country", type: "nominal" },
        y: { field: "spend", type: "quantitative" },
        color: {
          field: "device",
          type: "nominal",
          legendOrientation: "bottom",
        },
      } as CanvasChartSpec,
      data: data(rows),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(chart?.series).toHaveLength(8);
    expect(chart?.legendPosition).toBe("right");
  });

  it("maps scatter plots to a single series over the x measure", () => {
    const { table, chart } = mapChart({
      chartType: "scatter_plot",
      spec: {
        metrics_view: "mv",
        x: { field: "spend", type: "quantitative" },
        y: { field: "cpa", type: "quantitative" },
        dimension: { field: "country", type: "nominal" },
      } as CanvasChartSpec,
      data: data([{ country: "US", spend: 10, cpa: 2.5 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(table.columns.map((c) => c.name)).toStrictEqual([
      "country",
      "spend",
      "cpa",
    ]);
    expect(chart?.type).toBe(WorkbookChartType.SCATTER);
    expect(chart?.categoryColumn).toBe("spend");
    expect(chart?.numberFormat).toBe("0.00");
  });

  it("exports only the data of charts without a native equivalent", () => {
    const heatmap = mapChart({
      chartType: "heatmap",
      spec: {
        metrics_view: "mv",
        x: { field: "country", type: "nominal" },
        y: { field: "device", type: "nominal" },
        color: { field: "impressions", type: "quantitative" },
      } as CanvasChartSpec,
      data: data([{ country: "US", device: "mobile", impressions: 10 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(heatmap.chart).toBeUndefined();
    expect(heatmap.table.columns.map((c) => c.name)).toStrictEqual([
      "country",
      "device",
      "impressions",
    ]);

    const funnel = mapChart({
      chartType: "funnel_chart",
      spec: {
        metrics_view: "mv",
        stage: { field: "country", type: "nominal" },
        measure: { field: "impressions", type: "quantitative" },
      } as CanvasChartSpec,
      data: data([{ country: "US", impressions: 10 }]),
      timeGrain: undefined,
      timeZone: "UTC",
    });
    expect(funnel.chart).toBeUndefined();
    expect(funnel.table.rows).toStrictEqual([["US", 10]]);
  });
});

describe("hasNativeChart", () => {
  it("excludes heatmaps, funnels and split combo charts", () => {
    expect(
      hasNativeChart("bar_chart", { metrics_view: "mv" } as CanvasChartSpec),
    ).toBe(true);
    expect(
      hasNativeChart("heatmap", { metrics_view: "mv" } as CanvasChartSpec),
    ).toBe(false);
    expect(
      hasNativeChart("funnel_chart", { metrics_view: "mv" } as CanvasChartSpec),
    ).toBe(false);
    expect(
      hasNativeChart("combo_chart", { metrics_view: "mv" } as CanvasChartSpec),
    ).toBe(true);
    expect(
      hasNativeChart("combo_chart", {
        metrics_view: "mv",
        color: { field: "device", type: "nominal" },
      } as CanvasChartSpec),
    ).toBe(false);
  });
});
