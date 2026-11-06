import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/bar_chart.json";
import grouped from "../../harness/fixtures/grouped_bar.json";
import comparison from "../../harness/fixtures/line_comparison.json";
import pie from "../../harness/fixtures/pie.json";
import type { ChartResult, WireFlintChartSpec } from "./contract";
import {
  controlOptions,
  describeView,
  grainsForRange,
  patchInput,
  previousPeriod,
  seriesFieldOf,
  temporalFieldOf,
  TOP_N_PRESETS,
  topNChoices,
  topNTarget,
} from "./controls";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as ChartResult;

const week = { start: "2023-09-09T00:00:00Z", end: "2023-09-16T00:00:00Z" };

describe("grainsForRange", () => {
  it("offers grains that split the range into a readable number of buckets", () => {
    expect(grainsForRange(week, "TIME_GRAIN_HOUR")).toEqual([
      "TIME_GRAIN_HOUR",
      "TIME_GRAIN_DAY",
    ]);
    expect(grainsForRange(week, "TIME_GRAIN_DAY")).toEqual(["TIME_GRAIN_DAY"]);
  });
});

describe("topNChoices", () => {
  it("offers the presets plus the current limit when it is not one", () => {
    expect(topNChoices(undefined, TOP_N_PRESETS)).toEqual([5, 10, 20]);
    expect(topNChoices(10, TOP_N_PRESETS)).toEqual([5, 10, 20]);
    expect(topNChoices(15, TOP_N_PRESETS)).toEqual([5, 10, 15, 20]);
    expect(topNChoices(40, TOP_N_PRESETS)).toEqual([5, 10, 20, 40]);
    expect(topNChoices(3, [4, 6, 8])).toEqual([3, 4, 6, 8]);
  });
});

describe("previousPeriod", () => {
  it("shifts the range back by its own length", () => {
    expect(previousPeriod(week)).toEqual({
      start: "2023-09-02T00:00:00.000Z",
      end: "2023-09-09T00:00:00Z",
    });
  });
});

describe("temporalFieldOf", () => {
  it("finds the time field on x, by kind or by the metrics view's time dimension", () => {
    const res = result(comparison);
    expect(temporalFieldOf(res.input.chart_spec, res.fields)).toBe("__time");
    expect(temporalFieldOf(res.input.chart_spec, undefined, "__time")).toBe(
      "__time",
    );
    expect(
      temporalFieldOf(result(bar).input.chart_spec, result(bar).fields),
    ).toBeUndefined();
  });
});

describe("seriesFieldOf", () => {
  it("finds the dimension the series are colored by", () => {
    const res = result(grouped);
    expect(seriesFieldOf(res.input.chart_spec, res.fields)).toBe("device");
    const slices = result(pie);
    expect(seriesFieldOf(slices.input.chart_spec, slices.fields)).toBe(
      "device",
    );
    // A measure on color is a gradient, not a series.
    expect(
      seriesFieldOf(
        {
          chartType: "Heatmap",
          encodings: { x: "publisher", color: "impressions" },
        },
        res.fields,
      ),
    ).toBeUndefined();
  });
});

describe("topNTarget", () => {
  it("targets the first dimension channel with the limit in effect", () => {
    const res = result(bar);
    expect(topNTarget(res.input.chart_spec, res.fields)).toEqual({
      channel: "x",
      limit: 10,
    });
    const slices = result(pie);
    expect(topNTarget(slices.input.chart_spec, slices.fields)).toEqual({
      channel: "color",
      limit: 10,
    });
  });

  it("skips measures and time fields", () => {
    const res = result(comparison);
    expect(topNTarget(res.input.chart_spec, res.fields)).toBeUndefined();
  });
});

describe("controlOptions", () => {
  it("offers grains for a time axis and comparison for line charts", () => {
    const res = result(comparison);
    expect(
      controlOptions(
        res.input.chart_spec,
        res.fields,
        "__time",
        week,
        "TIME_GRAIN_HOUR",
      ),
    ).toEqual({
      grains: ["TIME_GRAIN_HOUR", "TIME_GRAIN_DAY"],
      topN: [],
      comparison: true,
    });
  });

  it("offers top-N without grains for a categorical bar chart", () => {
    const res = result(bar);
    expect(
      controlOptions(res.input.chart_spec, res.fields, "__time", week),
    ).toEqual({
      grains: [],
      topN: [5, 10, 20],
      comparison: true,
    });
  });

  it("offers the comparison only when the spec can draw it", () => {
    const res = result(comparison);
    const options = (
      encodings: WireFlintChartSpec["encodings"],
      chartType = "Line Chart",
    ) =>
      controlOptions(
        { chartType, encodings },
        { ...res.fields, device: { name: "device", kind: "dimension" } },
        "__time",
        week,
      )?.comparison;
    expect(options({ x: "__time", y: "total_bids" })).toBe(true);
    // The series channel is taken by a dimension.
    expect(options({ x: "__time", y: "total_bids", color: "device" })).toBe(
      false,
    );
    // Several measures on y are already a series fold.
    expect(options({ x: "__time", y: ["total_bids", "total_bids"] })).toBe(
      false,
    );
    // A chart type that does not fold.
    expect(options({ size: "total_bids", color: "device" }, "Pie Chart")).toBe(
      false,
    );
    // A grouped bar chart's clusters already take the series channel.
    const grouping = result(grouped);
    expect(
      controlOptions(
        grouping.input.chart_spec,
        grouping.fields,
        undefined,
        week,
      )?.comparison,
    ).toBe(false);
  });

  it("offers facet limits up to the server's cap of 8", () => {
    const res = result(comparison);
    const spec: WireFlintChartSpec = {
      chartType: "Line Chart",
      encodings: {
        x: "__time",
        y: "total_bids",
        column: { field: "device", limit: 6 },
      },
    };
    const fields = {
      ...res.fields,
      device: { name: "device", kind: "dimension" as const },
    };
    expect(topNTarget(spec, fields, "__time")).toEqual({
      channel: "column",
      limit: 6,
    });
    const topN = controlOptions(spec, fields, "__time", week)?.topN ?? [];
    expect(topN.length).toBeGreaterThan(1);
    expect(Math.max(...topN)).toBeLessThanOrEqual(8);
  });

  it("returns nothing when no control applies", () => {
    const res = result(pie);
    const spec = { ...res.input.chart_spec, encodings: { size: "total_bids" } };
    expect(controlOptions(spec, res.fields, undefined, week)).toBeUndefined();
  });
});

describe("patchInput", () => {
  it("sets the grain, the top-N limit and the comparison range without touching the input", () => {
    const res = result(bar);
    const before = JSON.stringify(res.input);
    const patched = patchInput(
      res.input,
      res.input.chart_spec,
      { grain: "TIME_GRAIN_WEEK", topN: 5, comparison: true },
      week,
      res.fields,
    );
    expect(patched.data.time_grain).toBe("TIME_GRAIN_WEEK");
    expect(patched.chart_spec.encodings.x).toEqual(
      expect.objectContaining({ field: "advertiser_name", limit: 5 }),
    );
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
    const patched = patchInput(res.input, spec, { topN: 20 }, week, res.fields);
    expect(patched.chart_spec.encodings.x).toEqual({
      field: "advertiser_name",
      limit: 20,
    });
  });

  it("sends no comparison for a spec that cannot draw one, which the server would reject", () => {
    const res = result(grouped);
    const input = {
      ...res.input,
      data: { ...res.input.data, comparison_time_range: week },
    };
    const patched = patchInput(
      input,
      input.chart_spec,
      { comparison: true },
      week,
      res.fields,
    );
    expect(patched.data.comparison_time_range).toBeUndefined();
  });

  it("keeps the caller's semantic type overrides", () => {
    const res = result(bar);
    const input = { ...res.input, semantic_types: { total_bids: "Amount" } };
    const patched = patchInput(
      input,
      input.chart_spec,
      { topN: 5 },
      week,
      res.fields,
    );
    expect(patched.semantic_types).toEqual({ total_bids: "Amount" });
    expect(
      patchInput(res.input, res.input.chart_spec, {}, week, res.fields),
    ).not.toHaveProperty("semantic_types");
  });

  it("removes the comparison range when comparison is switched off", () => {
    const res = result(comparison);
    expect(res.input.data.comparison_time_range).toBeDefined();
    const patched = patchInput(
      res.input,
      res.input.chart_spec,
      { comparison: false },
      week,
      res.fields,
    );
    expect(patched.data.comparison_time_range).toBeUndefined();
  });
});

describe("describeView", () => {
  it("states the chart type, the bindings and every control", () => {
    const res = result(bar);
    expect(
      describeView(
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
      describeView(
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
