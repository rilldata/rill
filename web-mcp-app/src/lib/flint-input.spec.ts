import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/flint_bar_chart.json";
import comparison from "../../harness/fixtures/flint_line_comparison.json";
import pie from "../../harness/fixtures/flint_pie.json";
import waterfall from "../../harness/fixtures/flint_waterfall.json";
import { normalizeRows, type FlintChartResult } from "./contract";
import {
  buildFlintCompileInput,
  COMPARISON_SERIES_FIELD,
  comparisonChannel,
  deriveFlintFields,
  encodingFields,
  measureFieldsInRows,
  singleEncodingField,
  stripFlintChartSpec,
} from "./flint-input";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as FlintChartResult;

describe("stripFlintChartSpec", () => {
  it("drops the title, the subtitle and Rill's limit keys", () => {
    const stripped = stripFlintChartSpec(result(waterfall).input.chart_spec);
    expect(stripped).toEqual({
      chartType: "Waterfall Chart",
      encodings: { x: { field: "publisher" }, y: "spend_change" },
      chartProperties: { totals: "both" },
    });
    expect(stripped).not.toHaveProperty("subtitle");
  });

  it("keeps Flint's encoding keys and strips inside arrays", () => {
    const stripped = stripFlintChartSpec({
      chartType: "Line Chart",
      encodings: {
        x: { field: "__time", limit: 5 },
        y: [
          "total_bids",
          { field: "impressions", sortOrder: "ascending", limit: 3 },
        ],
      },
    });
    expect(stripped.encodings).toEqual({
      x: { field: "__time" },
      y: ["total_bids", { field: "impressions", sortOrder: "ascending" }],
    });
    expect(stripped).not.toHaveProperty("chartProperties");
  });
});

describe("stripFlintChartSpec channels", () => {
  it("drops the server's detail extension where the chart type has no detail channel", () => {
    const stripped = stripFlintChartSpec({
      chartType: "Scatter Plot",
      encodings: { x: "total_bids", y: "win_rate", detail: "advertiser_name" },
    });
    expect(stripped.encodings).toEqual({ x: "total_bids", y: "win_rate" });
    const kept = stripFlintChartSpec({
      chartType: "Line Chart",
      encodings: { x: "__time", y: "total_bids", detail: "advertiser_name" },
    });
    expect(kept.encodings).toEqual({
      x: "__time",
      y: "total_bids",
      detail: "advertiser_name",
    });
  });

  it("keeps every channel of a chart type Flint does not know", () => {
    const stripped = stripFlintChartSpec({
      chartType: "Unknown Chart",
      encodings: { x: "a", detail: "b" },
    });
    expect(stripped.encodings).toEqual({ x: "a", detail: "b" });
  });
});

describe("encoding helpers", () => {
  it("reads single fields and all fields", () => {
    expect(singleEncodingField("total_bids")).toBe("total_bids");
    expect(singleEncodingField({ field: "device" })).toBe("device");
    expect(singleEncodingField(["a", "b"])).toBeUndefined();
    expect(singleEncodingField(undefined)).toBeUndefined();
    expect(encodingFields(["a", { field: "b" }, { sortBy: "y" }])).toEqual([
      "a",
      "b",
    ]);
  });
});

describe("deriveFlintFields", () => {
  it("mirrors the server's derivation from presets, formats and types", () => {
    const fields = deriveFlintFields(
      {
        __time: { name: "__time", kind: "time", display_name: "__time" },
        device: { name: "device", kind: "dimension", display_name: "Device" },
        total_bids: {
          name: "total_bids",
          kind: "measure",
          display_name: "Total Bids",
          type: "INT64",
        },
        total_spend: {
          name: "total_spend",
          kind: "measure",
          type: "FLOAT64",
          format_preset: "currency_usd",
        },
        win_rate: {
          name: "win_rate",
          kind: "measure",
          type: "FLOAT64",
          format_preset: "percentage",
        },
        avg_bid_price: {
          name: "avg_bid_price",
          kind: "measure",
          type: "FLOAT64",
          format_d3: "$.2f",
        },
        ratio: { name: "ratio", kind: "measure", type: "FLOAT64" },
      },
      "TIME_GRAIN_MONTH",
    );
    expect(fields.semantic_types).toEqual({
      __time: "YearMonth",
      device: "Category",
      total_bids: "Count",
      total_spend: "Amount",
      win_rate: "Percentage",
      avg_bid_price: "Amount",
      ratio: "Number",
    });
    expect(fields.field_display_names).toEqual({
      __time: "Time",
      device: "Device",
      total_bids: "Total Bids",
      total_spend: "total_spend",
      win_rate: "win_rate",
      avg_bid_price: "avg_bid_price",
      ratio: "ratio",
    });
  });

  it("uses DateTime for grains finer than a day and when the grain is unknown", () => {
    const time = { name: "t", kind: "time" as const };
    expect(
      deriveFlintFields({ t: time }, "TIME_GRAIN_HOUR").semantic_types.t,
    ).toBe("DateTime");
    expect(deriveFlintFields({ t: time }, undefined).semantic_types.t).toBe(
      "DateTime",
    );
  });
});

describe("buildFlintCompileInput", () => {
  it("prefers the server's semantic maps and lists the measure columns", () => {
    const res = result(bar);
    const rows = normalizeRows(res.data);
    const compile = buildFlintCompileInput(
      res.input.chart_spec,
      res,
      rows,
      undefined,
    );
    expect(compile.spec).toEqual({
      chartType: "Bar Chart",
      encodings: {
        x: { field: "advertiser_name", sortBy: "y", sortOrder: "descending" },
        y: "total_bids",
      },
    });
    expect(compile.fields).toEqual({
      semantic_types: { advertiser_name: "Category", total_bids: "Count" },
      field_display_names: {
        advertiser_name: "Advertiser Name",
        total_bids: "Total Bids",
      },
    });
    expect(compile.measureFields).toEqual(["total_bids"]);
    expect(compile.comparison).toBeUndefined();
  });

  it("falls back to deriving the semantics when the server omitted them", () => {
    const res = { ...result(pie) };
    delete res.semantic_types;
    delete res.field_display_names;
    const compile = buildFlintCompileInput(
      res.input.chart_spec,
      res,
      normalizeRows(res.data),
      undefined,
    );
    expect(compile.fields.semantic_types).toEqual({
      device: "Category",
      total_bids: "Count",
    });
    expect(compile.fields.field_display_names.device).toBe("Device");
  });

  it("folds the previous period in as a colored series in long form", () => {
    const res = result(comparison);
    const rows = normalizeRows(res.data);
    expect(comparisonChannel(res.input.chart_spec, res.fields ?? {})).toEqual({
      channel: "y",
      measure: "total_bids",
    });
    expect(measureFieldsInRows(res.fields ?? {}, rows)).toEqual([
      "total_bids",
      "total_bids_prev",
    ]);
    const compile = buildFlintCompileInput(
      res.input.chart_spec,
      res,
      rows,
      "TIME_GRAIN_DAY",
    );
    expect(compile.spec.chartType).toBe("Line Chart");
    expect(compile.spec.encodings).toEqual({
      x: "__time",
      y: "total_bids",
      color: COMPARISON_SERIES_FIELD,
    });
    expect(compile.comparison).toEqual({
      channel: "color",
      measure: "total_bids",
    });
    expect(compile.rows).toHaveLength(rows.length * 2);
    expect(compile.rows[0]).toEqual({
      __time: rows[0].__time,
      total_bids: rows[0].total_bids,
      rill_period: "Total Bids",
    });
    expect(compile.rows[1]).toEqual({
      __time: rows[0].__time,
      total_bids: rows[0].total_bids_prev,
      rill_period: "Total Bids (previous period)",
    });
    expect(compile.measureFields).toEqual(["total_bids"]);
    expect(compile.fields.semantic_types.rill_period).toBe("Category");
    expect(compile.fields.field_display_names.rill_period).toBe("Period");
  });

  it("names the previous period itself when the server did not", () => {
    const res = { ...result(comparison) };
    delete res.semantic_types;
    delete res.field_display_names;
    const compile = buildFlintCompileInput(
      res.input.chart_spec,
      res,
      normalizeRows(res.data),
      "TIME_GRAIN_DAY",
    );
    expect(
      compile.rows.some(
        (row) => row.rill_period === "Total Bids (previous period)",
      ),
    ).toBe(true);
  });

  it("turns a bar chart with a comparison into grouped bars", () => {
    const res = result(comparison);
    const compile = buildFlintCompileInput(
      { chartType: "Bar Chart", encodings: { x: "__time", y: "total_bids" } },
      res,
      normalizeRows(res.data),
      "TIME_GRAIN_DAY",
    );
    expect(compile.spec.chartType).toBe("Grouped Bar Chart");
    expect(compile.spec.encodings?.group).toBe(COMPARISON_SERIES_FIELD);
    expect(compile.comparison).toEqual({
      channel: "group",
      measure: "total_bids",
    });
  });

  it("layers the areas of a compared area chart", () => {
    const res = result(comparison);
    const compile = buildFlintCompileInput(
      { chartType: "Area Chart", encodings: { x: "__time", y: "total_bids" } },
      res,
      normalizeRows(res.data),
      "TIME_GRAIN_DAY",
    );
    expect(compile.spec.chartProperties).toEqual({ stackMode: "layered" });
  });

  it("does not fold outside the fold set, for multi-field channels or a taken series channel", () => {
    const res = result(comparison);
    const rows = normalizeRows(res.data);
    const heatmap = buildFlintCompileInput(
      { chartType: "Heatmap", encodings: { x: "__time", y: "total_bids" } },
      res,
      rows,
      "TIME_GRAIN_DAY",
    );
    expect(heatmap.comparison).toBeUndefined();
    expect(heatmap.spec.encodings?.y).toBe("total_bids");
    expect(heatmap.rows).toBe(rows);
    const multi = buildFlintCompileInput(
      {
        chartType: "Line Chart",
        encodings: { x: "__time", y: ["total_bids", "total_bids"] },
      },
      res,
      rows,
      "TIME_GRAIN_DAY",
    );
    expect(multi.comparison).toBeUndefined();
    const colored = buildFlintCompileInput(
      {
        chartType: "Bar Chart",
        encodings: { x: "__time", y: "total_bids", color: "__time" },
      },
      res,
      rows,
      "TIME_GRAIN_DAY",
    );
    expect(colored.comparison).toBeUndefined();
    expect(colored.spec.chartType).toBe("Bar Chart");
  });
});
