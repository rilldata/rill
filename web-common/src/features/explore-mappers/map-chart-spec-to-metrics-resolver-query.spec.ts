import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import {
  type ChartTimeRange,
  type CreateChartSpec,
  type FlintCreateChartArgs,
  isFlintCreateChartArgs,
  mapChartSpecToMetricsResolverQuery,
  mapFlintChartArgsToMetricsResolverQuery,
} from "@rilldata/web-common/features/explore-mappers/map-chart-spec-to-metrics-resolver-query.ts";
import { convertMetricsResolverQueryToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-metrics-resolver-query-to-url-params.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { Expression } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { describe, expect, it } from "vitest";

const timeRange = {
  start: "2024-01-01T00:00:00Z",
  end: "2024-12-31T23:59:59Z",
};
const comparisonTimeRange = {
  start: "2023-01-01T00:00:00Z",
  end: "2023-12-31T23:59:59Z",
};
// The generated type declares `val` as an object, but lists and scalars are what the resolver accepts.
const where = {
  cond: { op: "in", exprs: [{ name: "country" }, { val: ["US", "CA"] }] },
} as unknown as Expression;

describe("mapChartSpecToMetricsResolverQuery", () => {
  it("maps a bar chart over a nominal dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      where,
      // Ignored: there is no temporal field to floor.
      time_grain: V1TimeGrain.TIME_GRAIN_DAY,
      color: "primary",
      x: { field: "advertiser_name", type: "nominal", limit: 20, sort: "-y" },
      y: { field: "total_bids", type: "quantitative", zeroBasedOrigin: true },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "advertiser_name" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      where,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("maps a line chart over time with a grain and a comparison", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      time_grain: V1TimeGrain.TIME_GRAIN_MONTH,
      x: { field: "__time", type: "temporal" },
      y: { field: "total_bids", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [
        {
          name: "__time",
          compute: { time_floor: { dimension: "__time", grain: "month" } },
        },
      ],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
    });
  });

  it("floors a temporal field by day when the chart has no grain", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "__time", type: "temporal" },
      y: { field: "total_bids", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec).dimensions).toEqual([
      {
        name: "__time",
        compute: { time_floor: { dimension: "__time", grain: "day" } },
      },
    ]);
  });

  it("maps a heatmap, skipping its explicit value order", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      color: { field: "total_bids", type: "quantitative" },
      x: {
        field: "day",
        type: "nominal",
        sort: ["Sunday", "Monday", "Tuesday"],
      },
      y: { field: "hour", type: "nominal", limit: 24, sort: "-color" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "day" }, { name: "hour" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("maps every measure of a multi-measure chart and ignores its measures legend", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      color: { field: "rill_measures", type: "value" },
      x: { field: "__time", type: "temporal" },
      y: {
        field: "clicks",
        fields: ["clicks", "impressions", "ctr"],
        type: "quantitative",
      },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [
        {
          name: "__time",
          compute: { time_floor: { dimension: "__time", grain: "day" } },
        },
      ],
      measures: [{ name: "clicks" }, { name: "impressions" }, { name: "ctr" }],
      time_range: timeRange,
    });
  });

  it("maps a horizontal bar chart sorted by its measure on x", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "total_bids", type: "quantitative" },
      y: { field: "advertiser_name", type: "nominal", sort: "-x" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "advertiser_name" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("sorts by the comparison delta when the chart compares a single dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      x: { field: "advertiser_name", type: "nominal", sort: "-y_delta" },
      y: { field: "total_bids", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "advertiser_name" }],
      measures: [
        { name: "total_bids" },
        {
          name: "total_bids_delta",
          compute: { comparison_delta: { measure: "total_bids" } },
        },
      ],
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      sort: [{ name: "total_bids_delta", desc: true }],
    });
  });

  it("falls back to the measure for a delta sort without a comparison", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "advertiser_name", type: "nominal", sort: "y_delta" },
      y: { field: "total_bids", type: "quantitative" },
    };
    const query = mapChartSpecToMetricsResolverQuery(spec);
    expect(query.measures).toEqual([{ name: "total_bids" }]);
    expect(query.sort).toEqual([{ name: "total_bids", desc: false }]);
  });

  it("falls back to the measure for a delta sort when a color split adds a second dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
      x: { field: "advertiser_name", type: "nominal", sort: "-y_delta" },
      y: { field: "total_bids", type: "quantitative" },
      color: { field: "device", type: "nominal", limit: 5 },
    };
    const query = mapChartSpecToMetricsResolverQuery(spec);
    expect(query.dimensions).toEqual([
      { name: "advertiser_name" },
      { name: "device" },
    ]);
    expect(query.measures).toEqual([{ name: "total_bids" }]);
    expect(query.sort).toEqual([{ name: "total_bids", desc: true }]);
  });

  it("maps a funnel in measures mode to measures only", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids",
      time_range: timeRange,
      breakdownMode: "measures",
      color: "name",
      mode: "width",
      measure: {
        field: "impressions",
        type: "quantitative",
        fields: ["impressions", "video_starts", "video_completes"],
      },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids",
      time_zone: "UTC",
      measures: [
        { name: "impressions" },
        { name: "video_starts" },
        { name: "video_completes" },
      ],
      time_range: timeRange,
    });
  });

  it("maps a funnel in dimension mode with its stage sorted by the measure", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids",
      time_range: timeRange,
      breakdownMode: "dimension",
      stage: { field: "funnel_stage", type: "nominal", sort: "-y" },
      measure: { field: "users", type: "quantitative" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids",
      time_zone: "UTC",
      dimensions: [{ name: "funnel_stage" }],
      measures: [{ name: "users" }],
      time_range: timeRange,
      sort: [{ name: "users", desc: true }],
    });
  });

  it("maps a donut chart sorted by its slices' dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      measure: { field: "total_bids", type: "quantitative" },
      color: { field: "device", type: "nominal", limit: 10, sort: "color" },
      innerRadius: 50,
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "device" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      sort: [{ name: "device", desc: false }],
    });
  });

  it("maps a scatter plot's measures on both axes and its point dimension", () => {
    const spec: CreateChartSpec = {
      metrics_view: "bids_metrics",
      time_range: timeRange,
      x: { field: "impressions", type: "quantitative" },
      y: { field: "clicks", type: "quantitative" },
      size: { field: "total_bids", type: "quantitative" },
      dimension: { field: "advertiser_name", type: "nominal", limit: 50 },
      color: "primary",
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "advertiser_name" }],
      measures: [
        { name: "impressions" },
        { name: "clicks" },
        { name: "total_bids" },
      ],
      time_range: timeRange,
    });
  });

  it("maps a combo chart's two measures sorted by the first", () => {
    const spec: CreateChartSpec = {
      metrics_view: "auction_metrics",
      time_range: timeRange,
      x: { field: "app_site_name", type: "nominal", sort: "-y" },
      y1: { field: "requests", type: "quantitative", mark: "bar" },
      y2: { field: "bid_rate", type: "quantitative", mark: "line" },
    };
    expect(mapChartSpecToMetricsResolverQuery(spec)).toEqual({
      metrics_view: "auction_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "app_site_name" }],
      measures: [{ name: "requests" }, { name: "bid_rate" }],
      time_range: timeRange,
      sort: [{ name: "requests", desc: true }],
    });
  });
});

// The replay route feeds the mapped query to the same explore mapper as a `query_metrics_view` call.
describe("mapChartSpecToMetricsResolverQuery opens an explore", () => {
  function urlParamsFor(spec: CreateChartSpec) {
    const urlParams = convertMetricsResolverQueryToUrlParams(
      { query: mapChartSpecToMetricsResolverQuery(spec) },
      AD_BIDS_TIMESTAMP_DIMENSION,
    );
    expect(
      convertURLSearchParamsToExploreState(
        urlParams,
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
        AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION,
        {},
      ).errors,
    ).toEqual([]);
    return Object.fromEntries(urlParams.entries());
  }

  it("opens a bar chart as a filtered dimension table sorted by its measure", () => {
    expect(
      urlParamsFor({
        metrics_view: AD_BIDS_METRICS_NAME,
        time_range: timeRange,
        where: {
          cond: {
            op: "in",
            exprs: [
              { name: AD_BIDS_DOMAIN_DIMENSION },
              { val: ["a.com", "b.com"] },
            ],
          },
        } as unknown as Expression,
        x: { field: AD_BIDS_PUBLISHER_DIMENSION, type: "nominal", sort: "-y" },
        y: { field: AD_BIDS_IMPRESSIONS_MEASURE, type: "quantitative" },
      }),
    ).toEqual({
      dims: AD_BIDS_PUBLISHER_DIMENSION,
      measures: AD_BIDS_IMPRESSIONS_MEASURE,
      f: `${AD_BIDS_DOMAIN_DIMENSION} IN ('a.com','b.com')`,
      tr: "2024-01-01T00:00:00.000Z,2024-12-31T23:59:59.000Z",
      tz: "UTC",
      view: "explore",
      expand_dim: AD_BIDS_PUBLISHER_DIMENSION,
      sort_by: AD_BIDS_IMPRESSIONS_MEASURE,
      sort_dir: "DESC",
      sort_type: "value",
    });
  });

  it("opens a line chart over time as a time dimension detail at the chart's grain", () => {
    expect(
      urlParamsFor({
        metrics_view: AD_BIDS_METRICS_NAME,
        time_range: timeRange,
        comparison_time_range: comparisonTimeRange,
        time_grain: V1TimeGrain.TIME_GRAIN_WEEK,
        x: { field: AD_BIDS_TIMESTAMP_DIMENSION, type: "temporal" },
        y: { field: AD_BIDS_IMPRESSIONS_MEASURE, type: "quantitative" },
      }),
    ).toEqual({
      measures: AD_BIDS_IMPRESSIONS_MEASURE,
      tr: "2024-01-01T00:00:00.000Z,2024-12-31T23:59:59.000Z",
      tz: "UTC",
      compare_tr: "2023-01-01T00:00:00.000Z,2023-12-31T23:59:59.000Z",
      // The time dimension is the chart's axis, not a dimension of the explore.
      view: "tdd",
      measure: AD_BIDS_IMPRESSIONS_MEASURE,
      grain: "week",
    });
  });

  it("opens a heatmap as a flat pivot of both dimensions", () => {
    expect(
      urlParamsFor({
        metrics_view: AD_BIDS_METRICS_NAME,
        time_range: timeRange,
        color: { field: AD_BIDS_IMPRESSIONS_MEASURE, type: "quantitative" },
        x: {
          field: AD_BIDS_PUBLISHER_DIMENSION,
          type: "nominal",
          sort: ["Facebook", "Google"],
        },
        y: { field: AD_BIDS_DOMAIN_DIMENSION, type: "nominal", sort: "-color" },
      }),
    ).toEqual({
      dims: `${AD_BIDS_PUBLISHER_DIMENSION},${AD_BIDS_DOMAIN_DIMENSION}`,
      measures: AD_BIDS_IMPRESSIONS_MEASURE,
      tr: "2024-01-01T00:00:00.000Z,2024-12-31T23:59:59.000Z",
      tz: "UTC",
      view: "pivot",
      rows: "",
      cols: `${AD_BIDS_PUBLISHER_DIMENSION},${AD_BIDS_DOMAIN_DIMENSION},${AD_BIDS_IMPRESSIONS_MEASURE}`,
      table_mode: "flat",
      sort_by: AD_BIDS_IMPRESSIONS_MEASURE,
      sort_dir: "DESC",
    });
  });
});

describe("mapFlintChartArgsToMetricsResolverQuery", () => {
  const kinds = {
    advertiser_name: "dimension",
    device: "dimension",
    __time: "time",
    total_bids: "measure",
    impressions: "measure",
  } as const;

  it("recognizes the Flint call shape", () => {
    expect(
      isFlintCreateChartArgs({
        chart_spec: { chartType: "Bar Chart", encodings: {} },
        data: { metrics_view: "bids_metrics" },
      }),
    ).toBe(true);
    expect(isFlintCreateChartArgs({ chart_type: "bar_chart", spec: {} })).toBe(
      false,
    );
    expect(isFlintCreateChartArgs(undefined)).toBe(false);
  });

  it("maps a sorted bar chart with a filter", () => {
    const args: FlintCreateChartArgs = {
      chart_spec: {
        chartType: "Bar Chart",
        encodings: {
          x: {
            field: "advertiser_name",
            sortBy: "y",
            sortOrder: "descending",
            limit: 10,
          },
          y: "total_bids",
        },
      },
      data: { metrics_view: "bids_metrics", time_range: timeRange, where },
    };
    expect(mapFlintChartArgsToMetricsResolverQuery(args, kinds)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "advertiser_name" }],
      measures: [{ name: "total_bids" }],
      time_range: timeRange,
      where,
      sort: [{ name: "total_bids", desc: true }],
    });
  });

  it("floors the time field at the grain and keeps the comparison range", () => {
    const args: FlintCreateChartArgs = {
      chart_spec: {
        chartType: "Line Chart",
        encodings: {
          x: "__time",
          y: ["total_bids", "impressions"],
          color: { field: "device" },
        },
      },
      data: {
        metrics_view: "bids_metrics",
        time_range: timeRange,
        comparison_time_range: comparisonTimeRange,
        time_grain: V1TimeGrain.TIME_GRAIN_WEEK,
      },
    };
    expect(mapFlintChartArgsToMetricsResolverQuery(args, kinds)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [
        {
          name: "__time",
          compute: { time_floor: { dimension: "__time", grain: "week" } },
        },
        { name: "device" },
      ],
      measures: [{ name: "total_bids" }, { name: "impressions" }],
      time_range: timeRange,
      comparison_time_range: comparisonTimeRange,
    });
  });

  it("sorts by the encoding that asks for it, not the first on its channel", () => {
    const args: FlintCreateChartArgs = {
      chart_spec: {
        chartType: "Line Chart",
        encodings: {
          x: "__time",
          y: [
            { field: "total_bids" },
            { field: "impressions", sortOrder: "descending" },
          ],
        },
      },
      data: { metrics_view: "bids_metrics" },
    };
    expect(mapFlintChartArgsToMetricsResolverQuery(args, kinds).sort).toEqual([
      { name: "impressions", desc: true },
    ]);
  });

  it("opens the explore in the chart's time zone, UTC when the chart names none", () => {
    const open = (time_range: ChartTimeRange) => {
      return convertMetricsResolverQueryToUrlParams({
        query: mapFlintChartArgsToMetricsResolverQuery(
          {
            chart_spec: {
              chartType: "Line Chart",
              encodings: {
                x: AD_BIDS_TIMESTAMP_DIMENSION,
                y: AD_BIDS_IMPRESSIONS_MEASURE,
              },
            },
            data: { metrics_view: AD_BIDS_METRICS_NAME, time_range },
          },
          {
            [AD_BIDS_TIMESTAMP_DIMENSION]: "time",
            [AD_BIDS_IMPRESSIONS_MEASURE]: "measure",
          },
        ),
      }).get(ExploreStateURLParams.TimeZone);
    };

    expect(open({ ...timeRange, time_zone: "Asia/Kolkata" })).toBe(
      "Asia/Kolkata",
    );
    // The server queries a chart without a zone in UTC, where the explore would open in its
    // own default zone and bucket the days differently.
    expect(open(timeRange)).toBe("UTC");
  });

  it("treats fields the result does not describe as dimensions and sorts by the channel's own field", () => {
    const args: FlintCreateChartArgs = {
      chart_spec: {
        chartType: "Bar Chart",
        encodings: {
          x: { field: "publisher", sortOrder: "ascending" },
          y: "total_bids",
        },
      },
      data: { metrics_view: "bids_metrics" },
    };
    expect(mapFlintChartArgsToMetricsResolverQuery(args, kinds)).toEqual({
      metrics_view: "bids_metrics",
      time_zone: "UTC",
      dimensions: [{ name: "publisher" }],
      measures: [{ name: "total_bids" }],
      sort: [{ name: "publisher", desc: false }],
    });
  });
});
