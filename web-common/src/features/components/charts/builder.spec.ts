import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
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
