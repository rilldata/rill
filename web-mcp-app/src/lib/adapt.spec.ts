import { TOTAL_DOMAIN_KEY } from "@rilldata/web-common/features/components/charts/circular/constants";
import { get } from "svelte/store";
import { describe, expect, it } from "vitest";
import area from "../../harness/fixtures/area_chart.json";
import bar from "../../harness/fixtures/bar_chart.json";
import combo from "../../harness/fixtures/combo_chart.json";
import comparison from "../../harness/fixtures/comparison.json";
import donut from "../../harness/fixtures/donut_chart.json";
import empty from "../../harness/fixtures/empty.json";
import funnel from "../../harness/fixtures/funnel_chart.json";
import heatmap from "../../harness/fixtures/heatmap.json";
import line from "../../harness/fixtures/line_chart.json";
import multi from "../../harness/fixtures/multi_measure_bar.json";
import truncated from "../../harness/fixtures/truncated.json";
import truncatedRows from "../../harness/fixtures/truncated_rows.json";
import { adaptChartResult, MAX_TABLE_ROWS } from "./adapt";
import type { RillChartResult } from "./contract";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as RillChartResult;

const light = { isDark: false, locale: "en" };

describe("adaptChartResult", () => {
  it("builds measures, fields and domain values for the canonical bar chart", () => {
    const chart = adaptChartResult(result(bar), light);
    expect(chart.chartType).toBe("bar_chart");
    expect(chart.rowCount).toBe(8);
    expect(chart.measures).toEqual([
      expect.objectContaining({
        name: "total_bids",
        displayName: "Total Bids",
        formatPreset: "humanize",
      }),
    ]);
    expect(chart.snapshot.fields.advertiser_name).toEqual(
      expect.objectContaining({ displayName: "Advertiser Name" }),
    );
    expect(chart.snapshot.domainValues?.advertiser_name).toEqual(
      result(bar).domain_values?.advertiser_name,
    );
    expect(chart.snapshot.hasComparison).toBe(false);
    expect(chart.title).toBe("Total Bids per Advertiser Name");
    expect(chart.caption).toBe("Sep 9 – Sep 15, 2023 · Bids");
    expect(chart.openUrl).toBe(result(bar).open_url);
    expect(get(chart.chartData)).toBe(chart.snapshot);
  });

  it("resolves the project theme per mode", () => {
    const lightChart = adaptChartResult(result(bar), light);
    const darkChart = adaptChartResult(result(bar), { isDark: true });
    expect(lightChart.theme?.primary).toBe("#4f46e5");
    expect(darkChart.theme?.primary).toBe("#818cf8");
    expect(lightChart.snapshot.theme.primary.hex()).toBe("#4f46e5");
    expect(darkChart.snapshot.isDarkMode).toBe(true);
  });

  it("falls back to the humanized chart type when the result has no title", () => {
    const chart = adaptChartResult({ ...result(bar), title: undefined }, light);
    expect(chart.title).toBe("Bar chart");
    expect(chart.timeRangeLabel).toBe("Sep 9 – Sep 15, 2023");
  });

  it("converts temporal values to dates and zero-fills missing buckets", () => {
    const chart = adaptChartResult(result(line), light);
    expect(chart.timeGrain).toBe("TIME_GRAIN_DAY");
    expect(chart.temporalField).toBe("__time");
    // Six rows arrive, the seventh (Sep 12) is filled with zero.
    expect(chart.rowCount).toBe(7);
    const rows = chart.snapshot.data as Array<{
      __time: Date;
      total_bids: number;
    }>;
    rows.forEach((row) => expect(row.__time).toBeInstanceOf(Date));
    expect(rows.map((row) => row.__time.getDate())).toEqual([
      9, 10, 11, 12, 13, 14, 15,
    ]);
    // Same wall-clock time as the UTC bucket start, in the system zone.
    expect(rows[0].__time.getHours()).toBe(0);
    expect(rows[3].total_bids).toBe(0);
    expect(chart.snapshot.fields.__time).toEqual(
      expect.objectContaining({ timeUnit: "yearmonthdate", format: "%b %d" }),
    );
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
    const chart = adaptChartResult(
      { ...result(line), data, truncated: true },
      light,
    );
    const dates = (chart.snapshot.data as Array<{ __time: Date }>).map((row) =>
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
    const chart = adaptChartResult(
      { ...result(line), data, truncated: true, rows_truncated: true },
      light,
    );
    const rows = chart.snapshot.data as Array<{
      __time: Date;
      total_bids: number;
    }>;
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

  it("zero-fills per series when a color dimension is present", () => {
    const chart = adaptChartResult(result(area), light);
    expect(chart.rowCount).toBe(21);
    const rows = chart.snapshot.data as Array<{
      __time: Date;
      device: string;
      total_spend: number;
    }>;
    const mobileSep12 = rows.find(
      (row) => row.device === "Mobile" && row.__time.getDate() === 12,
    );
    expect(mobileSep12?.total_spend).toBe(0);
    expect(chart.snapshot.domainValues?.device).toEqual([
      "Mobile",
      "Desktop",
      "Tablet",
    ]);
  });

  it("detects comparison columns", () => {
    const chart = adaptChartResult(result(comparison), light);
    expect(chart.hasComparison).toBe(true);
    expect(chart.captionParts).toContain("vs. previous period");
    expect(chart.table.columns.map((column) => column.key)).toEqual([
      "__time",
      "total_bids",
      "total_bids_prev",
    ]);
    expect(chart.table.columns[2].label).toBe("Total Bids (prev)");
  });

  it("labels the multi-measure color domain with display names", () => {
    const chart = adaptChartResult(result(multi), light);
    expect(chart.snapshot.domainValues?.rill_measures).toEqual([
      "Total Bids",
      "Impressions",
    ]);
    expect(chart.measures.map((measure) => measure.name)).toEqual([
      "total_bids",
      "impressions",
    ]);
  });

  it("labels the combo color domain with both measures", () => {
    const chart = adaptChartResult(result(combo), light);
    expect(chart.snapshot.domainValues?.rill_measures).toEqual([
      "Total Bids",
      "Win Rate",
    ]);
  });

  it("carries the donut total and synthesizes an Other slice when asked", () => {
    const chart = adaptChartResult(result(donut), light);
    expect(chart.snapshot.domainValues?.[TOTAL_DOMAIN_KEY]).toEqual([151220.9]);
    expect(chart.snapshot.domainValues?.__other_value).toBeUndefined();

    const withOther = adaptChartResult(
      {
        ...result(donut),
        spec: { ...result(donut).spec, show_other: true },
      },
      light,
    );
    // 151220.9 total minus the 138781.9 the five visible publishers add up to.
    expect(withOther.snapshot.domainValues?.__other_value?.[0]).toBeCloseTo(
      12439,
      3,
    );
    expect(withOther.snapshot.domainValues?.publisher).toContain("Other");
  });

  it("falls back to first appearance for heatmap axes without domain values", () => {
    const chart = adaptChartResult(result(heatmap), light);
    expect(chart.snapshot.domainValues?.publisher).toEqual([
      "news.example",
      "sports.example",
      "video.example",
      "music.example",
      "shop.example",
    ]);
    expect(chart.snapshot.domainValues?.device).toEqual([
      "Mobile",
      "Desktop",
      "Tablet",
    ]);
  });

  it("passes funnel stages through untouched", () => {
    const chart = adaptChartResult(result(funnel), light);
    expect(chart.snapshot.domainValues?.campaign_stage).toEqual([
      "Impression",
      "Click",
      "Add to cart",
      "Purchase",
    ]);
  });

  it("coerces numeric strings in measure columns", () => {
    const chart = adaptChartResult(
      {
        ...result(bar),
        data: [
          { advertiser_name: "Acme", total_bids: "12345" },
          { advertiser_name: "Beta", total_bids: 9876 },
        ],
      },
      light,
    );
    expect(chart.snapshot.data).toEqual([
      { advertiser_name: "Acme", total_bids: 12345 },
      { advertiser_name: "Beta", total_bids: 9876 },
    ]);
  });

  it("reports empty results", () => {
    // A zero-row result carries an empty `data` array.
    expect(result(empty).data).toEqual([]);
    expect(adaptChartResult(result(empty), light).rowCount).toBe(0);
  });

  it("tells apart a top-N cap from cut rows", () => {
    expect(adaptChartResult(result(bar), light).truncationNote).toBeUndefined();

    const capped = adaptChartResult(result(truncated), light);
    expect(capped.truncated).toBe(true);
    expect(capped.rowsTruncated).toBe(false);
    // The x limit the backend applied is on the spec.
    expect(capped.truncationNote).toBe("Showing the top 10 values.");
    expect(capped.truncationWarning).toContain("Lower the x limit");

    const cut = adaptChartResult(result(truncatedRows), light);
    expect(cut.rowsTruncated).toBe(true);
    expect(cut.truncationNote).toBe("Showing the most recent periods.");
    // The first three days were cut, not empty, so nothing is filled before Sep 12.
    expect(cut.rowCount).toBe(4);

    // Cut rows of a chart without a time axis are neither periods nor top values.
    const points = adaptChartResult(
      { ...result(bar), truncated: true, rows_truncated: true },
      light,
    );
    expect(points.truncationNote).toBe("Showing a subset of the data.");
  });

  it("formats table cells with Rill's measure formatters and caps the rows", () => {
    const chart = adaptChartResult(result(bar), light);
    const measure = chart.table.columns.find(
      (column) => column.key === "total_bids",
    )!;
    expect(measure.format(128430)).toBe("128.4k");
    expect(measure.format(null)).toBe("");

    const manyRows = Array.from({ length: MAX_TABLE_ROWS + 50 }, (_, i) => ({
      advertiser_name: `adv ${i}`,
      total_bids: i,
    }));
    const big = adaptChartResult({ ...result(bar), data: manyRows }, light);
    expect(big.table.rows).toHaveLength(MAX_TABLE_ROWS);
    expect(big.table.totalRows).toBe(MAX_TABLE_ROWS + 50);
  });

  it("uses the time zone from the result before the option", () => {
    const zoned = adaptChartResult(
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
      adaptChartResult(result(line), { ...light, timeZone: "Asia/Tokyo" })
        .timeZone,
    ).toBe("Asia/Tokyo");
    expect(adaptChartResult(result(line), light).timeZone).toBe("UTC");
  });
});
