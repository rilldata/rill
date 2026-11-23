import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/bar_chart.json";
import empty from "../../harness/fixtures/empty.json";
import comparison from "../../harness/fixtures/line_comparison.json";
import line from "../../harness/fixtures/line_chart.json";
import truncated from "../../harness/fixtures/truncated.json";
import truncatedRows from "../../harness/fixtures/truncated_rows.json";
import { adaptResult, MAX_TABLE_ROWS } from "./adapt";
import type { ChartResult, WireFlintChartSpec, WireRow } from "./contract";
import { controlOptions } from "./controls";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as ChartResult;

const light = { isDark: false, locale: "en" };

describe("adaptResult", () => {
  it("builds the measures, the compile input and the table for the canonical bar chart", () => {
    const chart = adaptResult(result(bar), light);
    expect(chart.chartType).toBe("Bar Chart");
    expect(chart.chartSpec).toBe(result(bar).input.chart_spec);
    expect(chart.rowCount).toBe(8);
    expect(chart.measures).toEqual([
      expect.objectContaining({
        name: "total_bids",
        displayName: "Total Bids",
        formatPreset: "humanize",
      }),
    ]);
    // Flint sees the spec without Rill's `limit`, and the server's semantics.
    expect(chart.flint.spec.encodings?.x).toEqual({
      field: "advertiser_name",
      sortBy: "y",
      sortOrder: "descending",
    });
    expect(chart.flint.fields.semantic_types.total_bids).toBe("Count");
    expect(chart.table.columns.map((column) => column.label)).toEqual([
      "Advertiser Name",
      "Total Bids",
    ]);
    expect(chart.hasComparison).toBe(false);
    expect(chart.title).toBe("Total Bids by Advertiser Name");
    expect(chart.caption).toBe("Sep 9 – Sep 15, 2023 · Bids");
    expect(chart.openUrl).toBe(result(bar).open_url);
  });

  it("resolves the project theme per mode", () => {
    const lightChart = adaptResult(result(bar), light);
    const darkChart = adaptResult(result(bar), { isDark: true });
    expect(lightChart.theme?.primary).toBe("#4f46e5");
    expect(darkChart.theme?.primary).toBe("#818cf8");
  });

  it("falls back to the spec's title, then the chart type, when the result has no title", () => {
    const untitled = { ...result(bar), title: undefined };
    expect(adaptResult(untitled, light).title).toBe("Bar Chart");
    const chartSpec = { ...untitled.input.chart_spec, title: "Top bidders" };
    const chart = adaptResult(untitled, { ...light, chartSpec });
    expect(chart.title).toBe("Top bidders");
    expect(chart.timeRangeLabel).toBe("Sep 9 – Sep 15, 2023");
  });

  it("converts temporal values to dates and zero-fills missing buckets", () => {
    const chart = adaptResult(result(line), light);
    expect(chart.timeGrain).toBe("TIME_GRAIN_DAY");
    expect(chart.temporalField).toBe("__time");
    // Six rows arrive, the seventh (Sep 12) is filled with zero.
    expect(chart.rowCount).toBe(7);
    const rows = chart.rows as Array<{ __time: Date; total_bids: number }>;
    rows.forEach((row) => expect(row.__time).toBeInstanceOf(Date));
    expect(rows.map((row) => row.__time.getDate())).toEqual([
      9, 10, 11, 12, 13, 14, 15,
    ]);
    // Same wall-clock time as the UTC bucket start, in the system zone.
    expect(rows[0].__time.getHours()).toBe(0);
    expect(rows[3].total_bids).toBe(0);
    expect(chart.captionParts).toEqual([
      "Sep 9 – Sep 15, 2023",
      "by day",
      "Bids",
    ]);
  });

  it("fills the whole range when only a top-N cap applied", () => {
    // `truncated` without `rows_truncated` means a nominal channel was capped; the time axis
    // is complete, so the leading and trailing empty buckets are zeros like any other.
    const data = (result(line).data ?? []).filter((row) =>
      ["11", "13"].includes(String(row.__time).slice(8, 10)),
    );
    const chart = adaptResult(
      { ...result(line), data, truncated: true },
      light,
    );
    const dates = (chart.rows as Array<{ __time: Date }>).map((row) =>
      row.__time.getDate(),
    );
    expect(dates).toEqual([9, 10, 11, 12, 13, 14, 15]);
    expect(chart.rowsTruncated).toBe(false);
  });

  it("does not zero-fill before the first bucket of a row-truncated series", () => {
    // The server keeps the newest rows, so Sep 9 and 10 were cut rather than empty. Sep 12 is
    // a hole inside the data and Sep 14 and 15 come after the newest row; those are zeros.
    const data = (result(line).data ?? []).filter((row) =>
      ["11", "13"].includes(String(row.__time).slice(8, 10)),
    );
    const chart = adaptResult(
      { ...result(line), data, truncated: true, rows_truncated: true },
      light,
    );
    const rows = chart.rows as Array<{ __time: Date; total_bids: number }>;
    expect(rows.map((row) => row.__time.getDate())).toEqual([
      11, 12, 13, 14, 15,
    ]);
    expect(rows.map((row) => row.total_bids === 0)).toEqual([
      false,
      true,
      false,
      true,
      true,
    ]);
    expect(chart.rowsTruncated).toBe(true);
  });

  it("folds the comparison into a second series and says so in the caption", () => {
    const chart = adaptResult(result(comparison), light);
    expect(chart.hasComparison).toBe(true);
    expect(chart.flint.comparison).toEqual({
      channel: "color",
      measure: "total_bids",
    });
    expect(chart.captionParts).toContain("vs. previous period");
    expect(chart.table.columns.map((column) => column.key)).toEqual([
      "__time",
      "total_bids",
      "total_bids_prev",
    ]);
    expect(chart.table.columns[2].label).toBe("Total Bids (prev)");
  });

  it("captions only the comparison the chart draws", () => {
    // A chart type that cannot fold the previous period in draws the current period only;
    // the table still shows the previous values the rows carry.
    const chartSpec: WireFlintChartSpec = {
      chartType: "Heatmap",
      encodings: { x: "__time", color: "total_bids" },
    };
    const chart = adaptResult(result(comparison), { ...light, chartSpec });
    expect(chart.flint.comparison).toBeUndefined();
    expect(chart.hasComparison).toBe(false);
    expect(chart.captionParts).not.toContain("vs. previous period");
    expect(chart.table.columns.map((column) => column.key)).toContain(
      "total_bids_prev",
    );
  });

  it("coerces numeric strings in measure columns", () => {
    const chart = adaptResult(
      {
        ...result(bar),
        data: [
          { advertiser_name: "Acme", total_bids: "12345" },
          { advertiser_name: "Beta", total_bids: 9876 },
        ],
      },
      light,
    );
    expect(chart.rows).toEqual([
      { advertiser_name: "Acme", total_bids: 12345 },
      { advertiser_name: "Beta", total_bids: 9876 },
    ]);
  });

  it("reports empty results", () => {
    // A zero-row result carries an empty `data` array.
    expect(result(empty).data).toEqual([]);
    expect(adaptResult(result(empty), light).rowCount).toBe(0);
  });

  it("tells apart a top-N cap from cut rows", () => {
    expect(adaptResult(result(bar), light).truncationNote).toBeUndefined();

    const capped = adaptResult(result(truncated), light);
    expect(capped.truncated).toBe(true);
    expect(capped.rowsTruncated).toBe(false);
    // The x limit the backend applied is on the spec.
    expect(capped.truncationNote).toBe("Showing the top 10 values.");
    expect(capped.truncationWarning).toContain("Lower the x limit");

    const cut = adaptResult(result(truncatedRows), light);
    expect(cut.rowsTruncated).toBe(true);
    expect(cut.truncationNote).toBe("Showing the most recent periods.");
    // The first three days were cut, not empty, so nothing is filled before Sep 12.
    expect(cut.rowCount).toBe(4);

    // Cut rows of a chart without a time axis are neither periods nor top values.
    const points = adaptResult(
      { ...result(bar), truncated: true, rows_truncated: true },
      light,
    );
    expect(points.truncationNote).toBe("Showing a subset of the data.");
  });

  it("formats table cells with Rill's measure formatters and caps the rows", () => {
    const chart = adaptResult(result(bar), light);
    const measure = chart.table.columns.find(
      (column) => column.key === "total_bids",
    )!;
    expect(measure.format(128430)).toBe("128.4k");
    expect(measure.format(null)).toBe("");

    const manyRows = Array.from({ length: MAX_TABLE_ROWS + 50 }, (_, i) => ({
      advertiser_name: `adv ${i}`,
      total_bids: i,
    }));
    const big = adaptResult({ ...result(bar), data: manyRows }, light);
    expect(big.table.rows).toHaveLength(MAX_TABLE_ROWS);
    expect(big.table.totalRows).toBe(MAX_TABLE_ROWS + 50);
  });

  it("uses the time zone from the result before the option", () => {
    const zoned = adaptResult(
      {
        ...result(line),
        resolved_time_range: {
          start: "2023-09-09T00:00:00Z",
          end: "2023-09-16T00:00:00Z",
          time_zone: "America/New_York",
        },
      },
      { ...light, timeZone: "Asia/Tokyo" },
    );
    expect(zoned.timeZone).toBe("America/New_York");
    expect(
      adaptResult(result(line), { ...light, timeZone: "Asia/Tokyo" }).timeZone,
    ).toBe("Asia/Tokyo");
    expect(adaptResult(result(line), light).timeZone).toBe("UTC");
  });
});

describe("adaptResult on hand-built results", () => {
  // Four daily buckets: Sep 9 to Sep 12.
  const range = { start: "2023-09-09T00:00:00Z", end: "2023-09-13T00:00:00Z" };
  const day = (n: number) => `2023-09-${String(n).padStart(2, "0")}T00:00:00Z`;

  function flintResult(
    chartSpec: WireFlintChartSpec,
    data: WireRow[],
    domain?: Record<string, unknown[]>,
  ): ChartResult {
    return {
      widget: "chart",
      chart_type: chartSpec.chartType,
      input: {
        chart_spec: chartSpec,
        data: { metrics_view: "bids", time_range: range },
      },
      data,
      time_grain: "TIME_GRAIN_DAY",
      fields: {
        __time: { name: "__time", kind: "time" },
        device: { name: "device", kind: "dimension" },
        publisher: { name: "publisher", kind: "dimension" },
        bids: { name: "bids", kind: "measure" },
      },
      domain_values: domain,
      resolved_time_range: range,
      metrics_view: { name: "bids", time_dimension: "__time" },
    };
  }

  const countBy = (rows: WireRow[], field: string) =>
    rows.reduce<Record<string, number>>((counts, row) => {
      const key = String(row[field]);
      counts[key] = (counts[key] ?? 0) + 1;
      return counts;
    }, {});

  const firstAppearance = (rows: WireRow[], field: string) =>
    Array.from(new Set(rows.map((row) => row[field])));

  it("zero-fills every series of a time series, as Rill's charts do", () => {
    const chart = adaptResult(
      flintResult(
        {
          chartType: "Line Chart",
          encodings: { x: "__time", y: "bids", color: "device" },
        },
        [
          { __time: day(9), device: "Mobile", bids: 5 },
          { __time: day(10), device: "Desktop", bids: 3 },
          { __time: day(11), device: "Mobile", bids: 4 },
        ],
        { device: ["Mobile", "Desktop"] },
      ),
      light,
    );
    expect(countBy(chart.rows, "device")).toEqual({ Mobile: 4, Desktop: 4 });
    expect(chart.rows.filter((row) => row.bids === 0)).toHaveLength(5);
    // The table shows the filled rows too, in time order.
    expect(chart.table.totalRows).toBe(8);
  });

  it("zero-fills the previous period of a comparison as its own series", () => {
    const chart = adaptResult(
      flintResult(
        { chartType: "Area Chart", encodings: { x: "__time", y: "bids" } },
        [
          { __time: day(9), bids: 5, bids_prev: 2 },
          { __time: day(12), bids: 4, bids_prev: 1 },
        ],
      ),
      light,
    );
    expect(chart.flint.comparison).toBeDefined();
    const periods = countBy(chart.rows, "rill_period");
    expect(Object.values(periods)).toEqual([4, 4]);
    expect(chart.rows.filter((row) => row.bids === 0)).toHaveLength(4);
  });

  it("leaves rows split by another dimension, and charts that are not time series", () => {
    const rows = [
      { __time: day(9), device: "Mobile", publisher: "news", bids: 5 },
      { __time: day(11), device: "Mobile", publisher: "sports", bids: 4 },
    ];
    const faceted = adaptResult(
      flintResult(
        {
          chartType: "Line Chart",
          encodings: {
            x: "__time",
            y: "bids",
            color: "device",
            column: "publisher",
          },
        },
        rows,
      ),
      light,
    );
    expect(faceted.rows).toHaveLength(2);
    const timeOnColor = flintResult(
      {
        chartType: "Line Chart",
        encodings: { x: "__time", y: "bids", color: "order_date" },
      },
      [
        { __time: day(9), order_date: day(1), bids: 5 },
        { __time: day(11), order_date: day(2), bids: 4 },
      ],
    );
    timeOnColor.fields!.order_date = { name: "order_date", kind: "time" };
    expect(adaptResult(timeOnColor, light).rows).toHaveLength(2);
    const heatmap = adaptResult(
      flintResult(
        {
          chartType: "Heatmap",
          encodings: { x: "__time", y: "device", color: "bids" },
        },
        rows,
      ),
      light,
    );
    expect(heatmap.rows).toHaveLength(2);
  });

  it("orders time rows by the server's top-N order, so a refresh keeps every series' color", () => {
    const spec: WireFlintChartSpec = {
      chartType: "Line Chart",
      encodings: { x: "__time", y: "bids", color: "device" },
    };
    const domain = { device: ["Android", "iOS", "Roku OS"] };
    // A bucket lists its series in no particular order, and each query orders them differently.
    const byDay = adaptResult(
      flintResult(
        spec,
        [
          { __time: day(9), device: "Roku OS", bids: 1 },
          { __time: day(9), device: "Android", bids: 9 },
          { __time: day(9), device: "iOS", bids: 4 },
        ],
        domain,
      ),
      light,
    );
    const byWeek = adaptResult(
      flintResult(
        spec,
        [
          { __time: day(9), device: "iOS", bids: 30 },
          { __time: day(9), device: "Roku OS", bids: 8 },
          { __time: day(9), device: "Android", bids: 60 },
        ],
        domain,
      ),
      light,
    );
    expect(firstAppearance(byDay.rows, "device")).toEqual(domain.device);
    expect(firstAppearance(byWeek.rows, "device")).toEqual(domain.device);
    // The table keeps the result's own order.
    expect(byDay.table.rows.slice(0, 3).map((row) => row.device)).toEqual([
      "Roku OS",
      "Android",
      "iOS",
    ]);
  });

  it("handles a time field other than the metrics view's time dimension", () => {
    const spec: WireFlintChartSpec = {
      chartType: "Line Chart",
      encodings: { x: "order_date", y: "bids" },
    };
    const base = flintResult(spec, [
      { order_date: day(9), bids: 5 },
      { order_date: day(11), bids: 4 },
    ]);
    const result: ChartResult = {
      ...base,
      fields: {
        ...base.fields,
        order_date: {
          name: "order_date",
          kind: "time",
          display_name: "Order Date",
          type: "DATE",
        },
      },
    };
    const chart = adaptResult(result, light);
    expect(chart.temporalField).toBe("order_date");
    // Zero-filled and converted to dates like the time dimension.
    expect(chart.rows).toHaveLength(4);
    expect(chart.rows.every((row) => row.order_date instanceof Date)).toBe(
      true,
    );
    expect(
      chart.table.columns.find((column) => column.key === "order_date")?.label,
    ).toBe("Order Date");
    expect(
      controlOptions(spec, result.fields, "__time", range)?.grains,
    ).toContain("TIME_GRAIN_DAY");
  });

  it("says the cut values of a pie are combined, rather than left out", () => {
    const pie = (limit: number, domain: string[]) =>
      adaptResult(
        {
          ...flintResult(
            {
              chartType: "Pie Chart",
              encodings: { color: { field: "device", limit }, size: "bids" },
            },
            [
              { device: "Mobile", bids: 9 },
              { device: "Desktop", bids: 5 },
              { device: "Other", bids: 3 },
            ],
            { device: domain },
          ),
          truncated: true,
        },
        light,
      );
    const combined = pie(2, ["Mobile", "Desktop", "Other"]);
    expect(combined.truncationNote).toBe(
      "The remaining values are combined into “Other”.",
    );
    // The combined slice stays last.
    expect(combined.rows.at(-1)?.device).toBe("Other");
    // Without the extra slice the note says what the top-N kept.
    expect(pie(3, ["Mobile", "Desktop", "Other"]).truncationNote).toBe(
      "Showing the top 3 values.",
    );
  });

  it("keeps the server's row order when no time axis orders the rows", () => {
    const chart = adaptResult(
      flintResult(
        {
          chartType: "Bar Chart",
          encodings: { x: "publisher", y: "bids", color: "device" },
        },
        [
          { publisher: "news", device: "Desktop", bids: 9 },
          { publisher: "news", device: "Mobile", bids: 5 },
          { publisher: "sports", device: "Mobile", bids: 4 },
        ],
        { device: ["Mobile", "Desktop"], publisher: ["news", "sports"] },
      ),
      light,
    );
    expect(chart.rows.map((row) => row.device)).toEqual([
      "Desktop",
      "Mobile",
      "Mobile",
    ]);
  });
});
