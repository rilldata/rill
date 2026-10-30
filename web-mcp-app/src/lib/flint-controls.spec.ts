import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/flint_bar_chart.json";
import comparison from "../../harness/fixtures/flint_line_comparison.json";
import grouped from "../../harness/fixtures/flint_grouped_bar.json";
import pie from "../../harness/fixtures/flint_pie.json";
import type { FlintChartResult } from "./contract";
import {
  describeFlintView,
  flintControlOptions,
  flintTemporalField,
  flintTopNTarget,
  patchFlintInput,
} from "./flint-controls";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as FlintChartResult;

const week = { start: "2023-09-09T00:00:00Z", end: "2023-09-16T00:00:00Z" };

describe("flintTemporalField", () => {
  it("finds the time field on x, by kind or by the metrics view's time dimension", () => {
    const res = result(comparison);
    expect(flintTemporalField(res.input.chart_spec, res.fields)).toBe("__time");
    expect(flintTemporalField(res.input.chart_spec, undefined, "__time")).toBe(
      "__time",
    );
    expect(
      flintTemporalField(result(bar).input.chart_spec, result(bar).fields),
    ).toBeUndefined();
  });
});

describe("flintTopNTarget", () => {
  it("targets the first dimension channel with the limit in effect", () => {
    const res = result(bar);
    expect(flintTopNTarget(res.input.chart_spec, res.fields)).toEqual({
      channel: "x",
      limit: 10,
    });
    const slices = result(pie);
    expect(flintTopNTarget(slices.input.chart_spec, slices.fields)).toEqual({
      channel: "color",
      limit: 10,
    });
  });

  it("skips measures and time fields", () => {
    const res = result(comparison);
    expect(flintTopNTarget(res.input.chart_spec, res.fields)).toBeUndefined();
  });
});

describe("flintControlOptions", () => {
  it("offers grains for a time axis and comparison for line charts", () => {
    const res = result(comparison);
    expect(
      flintControlOptions(
        res.input.chart_spec,
        res.fields,
        "__time",
        week,
        "TIME_GRAIN_HOUR",
      ),
    ).toEqual({
      grains: ["TIME_GRAIN_HOUR", "TIME_GRAIN_DAY"],
      topN: false,
      comparison: true,
    });
  });

  it("offers top-N without grains for a categorical bar chart", () => {
    const res = result(bar);
    expect(
      flintControlOptions(res.input.chart_spec, res.fields, "__time", week),
    ).toEqual({
      grains: [],
      topN: true,
      comparison: true,
    });
  });

  it("returns nothing when no control applies", () => {
    const res = result(pie);
    const spec = { ...res.input.chart_spec, encodings: { size: "total_bids" } };
    expect(
      flintControlOptions(spec, res.fields, undefined, week),
    ).toBeUndefined();
  });
});

describe("patchFlintInput", () => {
  it("sets the grain, the top-N limit and the comparison range without touching the input", () => {
    const res = result(grouped);
    const before = JSON.stringify(res.input);
    const patched = patchFlintInput(
      res.input,
      res.input.chart_spec,
      { grain: "TIME_GRAIN_WEEK", topN: 5, comparison: true },
      week,
      res.fields,
    );
    expect(patched.data.time_grain).toBe("TIME_GRAIN_WEEK");
    expect(patched.chart_spec.encodings.x).toEqual({
      field: "publisher",
      limit: 5,
    });
    expect(patched.data.comparison_time_range?.end).toBe(week.start);
    expect(patched.data.comparison_time_range?.start).toMatch(
      /^2023-09-02T00:00:00/,
    );
    expect(JSON.stringify(res.input)).toBe(before);
  });

  it("turns a string encoding into an object to hold the limit", () => {
    const res = result(bar);
    const spec = {
      ...res.input.chart_spec,
      encodings: { x: "advertiser_name", y: "total_bids" },
    };
    const patched = patchFlintInput(
      res.input,
      spec,
      { topN: 20 },
      week,
      res.fields,
    );
    expect(patched.chart_spec.encodings.x).toEqual({
      field: "advertiser_name",
      limit: 20,
    });
  });

  it("removes the comparison range when comparison is switched off", () => {
    const res = result(comparison);
    expect(res.input.data.comparison_time_range).toBeDefined();
    const patched = patchFlintInput(
      res.input,
      res.input.chart_spec,
      { comparison: false },
      week,
      res.fields,
    );
    expect(patched.data.comparison_time_range).toBeUndefined();
  });
});

describe("describeFlintView", () => {
  it("states the chart type, the bindings and every control", () => {
    const res = result(bar);
    expect(
      describeFlintView(
        "Total Bids by Advertiser Name",
        res.input.chart_spec,
        { topN: 10, comparison: false },
        "Sep 9 – Sep 15, 2023",
      ),
    ).toBe(
      "Chart now shows Total Bids by Advertiser Name: Bar Chart with x = advertiser_name, y = total_bids; time grain n/a, top 10 values, comparison off, Sep 9 – Sep 15, 2023.",
    );
  });

  it("lists folded series and the grain unit", () => {
    expect(
      describeFlintView(
        "Bids",
        {
          chartType: "Line Chart",
          encodings: { x: "__time", y: ["total_bids", "total_bids_prev"] },
        },
        { grain: "TIME_GRAIN_DAY", comparison: true },
      ),
    ).toBe(
      "Chart now shows Bids: Line Chart with x = __time, y = total_bids + total_bids_prev; time grain day, top values n/a, comparison on.",
    );
  });
});
