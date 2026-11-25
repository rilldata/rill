import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { getUrlFromAggregationRequest } from "@rilldata/web-common/features/explore-mappers/url-mappers/get-url-from-aggregation-request";
import { mapMetricsResolverQueryToUrlSearch } from "@rilldata/web-common/features/explore-mappers/url-mappers/map-metrics-resolver-query-to-url-search";
import {
  type V1MetricsViewAggregationRequest,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema";
import { describe, expect, it } from "vitest";

// Requests without a time range or comparison measure filters do not query the runtime.
const NoopClient = {} as RuntimeClient;

function expectParsesWithoutErrors(urlParams: URLSearchParams) {
  const { errors } = convertURLSearchParamsToExploreState(
    urlParams,
    AD_BIDS_METRICS_INIT_WITH_TIME,
    AD_BIDS_EXPLORE_INIT,
    {},
  );
  expect(errors).toEqual([]);
}

describe("getUrlFromAggregationRequest", () => {
  const TestCases: {
    title: string;
    req: V1MetricsViewAggregationRequest;
    forceOpenPivot?: boolean;
    expected: Record<string, string>;
  }[] = [
    {
      title: "dimension table for a request with a dimension",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
        measures: [
          { name: AD_BIDS_IMPRESSIONS_MEASURE },
          { name: AD_BIDS_IMPRESSIONS_MEASURE + "__delta_rel" },
        ],
        sort: [{ name: AD_BIDS_IMPRESSIONS_MEASURE, desc: true }],
      },
      expected: {
        measures: AD_BIDS_IMPRESSIONS_MEASURE,
        sort_by: AD_BIDS_IMPRESSIONS_MEASURE,
        sort_dir: "DESC",
        sort_type: "value",
        view: "explore",
        expand_dim: AD_BIDS_PUBLISHER_DIMENSION,
      },
    },
    {
      title: "time dimension details for a request without dimensions",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
      },
      expected: {
        measures: AD_BIDS_BID_PRICE_MEASURE,
        view: "tdd",
        measure: AD_BIDS_BID_PRICE_MEASURE,
      },
    },
    {
      title: "pivot when forced",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [
          {
            name: AD_BIDS_TIMESTAMP_DIMENSION,
            timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
            alias: "day",
          },
          { name: AD_BIDS_DOMAIN_DIMENSION },
        ],
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        pivotOn: ["day"],
        sort: [{ name: "day", desc: false }],
      },
      forceOpenPivot: true,
      expected: {
        view: "pivot",
        rows: AD_BIDS_DOMAIN_DIMENSION,
        cols: `time.day,${AD_BIDS_IMPRESSIONS_MEASURE}`,
        table_mode: "nest",
        sort_by: "time.day",
        sort_dir: "ASC",
      },
    },
  ];

  for (const { title, req, forceOpenPivot, expected } of TestCases) {
    it(title, async () => {
      const urlParams = await getUrlFromAggregationRequest({
        client: NoopClient,
        req,
        forceOpenPivot,
      });
      expect(Object.fromEntries(urlParams.entries())).toEqual(expected);
      expectParsesWithoutErrors(urlParams);
    });
  }
});

describe("mapMetricsResolverQueryToUrlSearch", () => {
  const TestCases: {
    title: string;
    query: MetricsResolverQuery;
    expected: Record<string, string>;
  }[] = [
    {
      title: "dimension table sorted by a comparison",
      query: {
        metrics_view: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
        measures: [
          { name: AD_BIDS_IMPRESSIONS_MEASURE },
          {
            name: "impressions_delta",
            compute: {
              comparison_delta: { measure: AD_BIDS_IMPRESSIONS_MEASURE },
            },
          },
        ],
        sort: [{ name: "impressions_delta", desc: true }],
        time_range: { expression: "P7D" },
        comparison_time_range: { expression: "rill-PW" },
      },
      expected: {
        dims: AD_BIDS_PUBLISHER_DIMENSION,
        measures: AD_BIDS_IMPRESSIONS_MEASURE,
        tr: "P7D",
        compare_tr: "rill-PW",
        view: "explore",
        expand_dim: AD_BIDS_PUBLISHER_DIMENSION,
        sort_by: AD_BIDS_IMPRESSIONS_MEASURE,
        sort_dir: "DESC",
        sort_type: "delta_abs",
      },
    },
    {
      title: "time dimension details split by a dimension",
      query: {
        metrics_view: AD_BIDS_METRICS_NAME,
        dimensions: [
          {
            name: "ts_day",
            compute: {
              time_floor: {
                dimension: AD_BIDS_TIMESTAMP_DIMENSION,
                grain: "day",
              },
            },
          },
          { name: AD_BIDS_DOMAIN_DIMENSION },
        ],
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        // A filter on the time dimension moves to the time range.
        where: {
          cond: {
            op: "and",
            exprs: [
              {
                cond: {
                  op: "gte",
                  exprs: [
                    { name: AD_BIDS_TIMESTAMP_DIMENSION },
                    { val: "2024-01-01T00:00:00Z" },
                  ],
                },
              },
              {
                cond: {
                  op: "lt",
                  exprs: [
                    { name: AD_BIDS_TIMESTAMP_DIMENSION },
                    { val: "2024-02-01T00:00:00Z" },
                  ],
                },
              },
            ],
          },
        } as MetricsResolverQuery["where"],
      },
      expected: {
        dims: AD_BIDS_DOMAIN_DIMENSION,
        measures: AD_BIDS_BID_PRICE_MEASURE,
        tr: "2024-01-01T00:00:00.000Z,2024-02-01T00:00:00.000Z",
        view: "tdd",
        measure: AD_BIDS_BID_PRICE_MEASURE,
        compare_dim: AD_BIDS_DOMAIN_DIMENSION,
        grain: "day",
      },
    },
    {
      title: "pivot for several dimensions",
      query: {
        metrics_view: AD_BIDS_METRICS_NAME,
        dimensions: [
          { name: AD_BIDS_PUBLISHER_DIMENSION },
          { name: AD_BIDS_DOMAIN_DIMENSION },
        ],
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        sort: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
      },
      expected: {
        dims: `${AD_BIDS_PUBLISHER_DIMENSION},${AD_BIDS_DOMAIN_DIMENSION}`,
        measures: AD_BIDS_IMPRESSIONS_MEASURE,
        tr: "inf",
        view: "pivot",
        rows: "",
        cols: `${AD_BIDS_PUBLISHER_DIMENSION},${AD_BIDS_DOMAIN_DIMENSION},${AD_BIDS_IMPRESSIONS_MEASURE}`,
        table_mode: "flat",
        sort_by: AD_BIDS_DOMAIN_DIMENSION,
        sort_dir: "ASC",
      },
    },
  ];

  for (const { title, query, expected } of TestCases) {
    it(title, () => {
      const urlParams = mapMetricsResolverQueryToUrlSearch({ query });
      expect(Object.fromEntries(urlParams.entries())).toEqual(expected);
      expectParsesWithoutErrors(urlParams);
    });
  }
});
