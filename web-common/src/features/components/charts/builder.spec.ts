import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import { createPositionEncoding } from "@rilldata/web-common/features/components/charts/builder";
import { generateVLBarChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/bar-chart/spec";
import { generateVLLineChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/line-chart/spec";
import chroma from "chroma-js";
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
    expect(axis.labelLimit).toEqual({
      expr: `(width / 3) >= ${widestLabelPx} ? (width / 3) - 4 : max(60, height * 0.5)`,
    });
    expect(axis.labelOverlap).toBe(false);
    for (const value of [axis.labelAngle, axis.labelLimit]) {
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
  });
});

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
