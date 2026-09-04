import { describe, expect, it } from "vitest";
import {
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
  ],
  dimensions: [{ name: "publisher", displayName: "Publisher" }],
};

const rows = [
  { publisher: "Yahoo", bid_price_sum: 1200, total_records: 40 },
  { publisher: "Google", bid_price_sum: 1450, total_records: 55 },
  { publisher: "Microsoft", bid_price_sum: 980, total_records: 31 },
];

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
