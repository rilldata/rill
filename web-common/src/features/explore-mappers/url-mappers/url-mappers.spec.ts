import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { toEphemeralMeasuresParam } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-param";
import { ComparisonDeltaAbsoluteSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_INIT_DIMENSIONS,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { convertAggregationRequestToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-aggregation-request-to-url-params";
import { convertComparisonRequestToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-comparison-request-to-url-params";
import { convertMetricsResolverQueryToUrlParams } from "@rilldata/web-common/features/explore-mappers/url-mappers/convert-metrics-resolver-query-to-url-params";
import { getMappedExploreUrl } from "@rilldata/web-common/features/explore-mappers/url-mappers/get-mapped-url";
import {
  MetricsViewSpecDimensionType,
  type V1ExploreSpec,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
  V1MetricsViewComparisonMeasureType,
  V1Operation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema";
import { beforeEach, describe, expect, it } from "vitest";

// Requests without a time range or comparison measure filters do not query the runtime.
const NoopClient = {} as RuntimeClient;
const NO_TIME_METRICS_NAME = "AdBids_no_time_metrics";

function expectParsesWithoutErrors(urlParams: URLSearchParams) {
  const { errors } = convertURLSearchParamsToExploreState(
    urlParams,
    AD_BIDS_METRICS_INIT_WITH_TIME,
    AD_BIDS_EXPLORE_INIT,
  );
  expect(errors).toEqual([]);
}

describe("convertAggregationRequestToUrlParams", () => {
  const TestCases: {
    title: string;
    req: V1MetricsViewAggregationRequest;
    urlState?: string;
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
    {
      title: "pivot when forced overrides the saved view and sort",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [
          { name: AD_BIDS_PUBLISHER_DIMENSION },
          { name: AD_BIDS_DOMAIN_DIMENSION },
        ],
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        pivotOn: [AD_BIDS_DOMAIN_DIMENSION],
      },
      urlState: `?view=explore&expand_dim=${AD_BIDS_PUBLISHER_DIMENSION}&sort_by=${AD_BIDS_IMPRESSIONS_MEASURE}&sort_dir=DESC&sort_type=value`,
      forceOpenPivot: true,
      expected: {
        view: "pivot",
        expand_dim: AD_BIDS_PUBLISHER_DIMENSION,
        rows: AD_BIDS_PUBLISHER_DIMENSION,
        cols: `${AD_BIDS_DOMAIN_DIMENSION},${AD_BIDS_IMPRESSIONS_MEASURE}`,
        table_mode: "nest",
      },
    },
    {
      title: "saved view is kept when not forced to pivot",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
      },
      urlState: `?view=explore&expand_dim=${AD_BIDS_PUBLISHER_DIMENSION}`,
      expected: {
        view: "explore",
        expand_dim: AD_BIDS_PUBLISHER_DIMENSION,
      },
    },
  ];

  for (const { title, req, urlState, forceOpenPivot, expected } of TestCases) {
    it(title, async () => {
      const urlParams = await convertAggregationRequestToUrlParams({
        client: NoopClient,
        req,
        urlState,
        forceOpenPivot,
      });
      expect(Object.fromEntries(urlParams.entries())).toEqual(expected);
      expectParsesWithoutErrors(urlParams);
    });
  }
});

// Fixtures for the queries saved in alerts and reports.
describe("convertAggregationRequestToUrlParams with a runtime", () => {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  const client = new RuntimeClient({
    host: "http://localhost:9009",
    instanceId: "default",
  });

  beforeEach(() => {
    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockMetricsView(NO_TIME_METRICS_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
      "7D": { start: "2024-01-01T05:00:00Z", end: "2024-01-08T05:00:00Z" },
      P1D: { start: "2024-01-07T05:00:00Z", end: "2024-01-08T05:00:00Z" },
    });
  });

  const TestCases: {
    title: string;
    req: V1MetricsViewAggregationRequest;
    urlState?: string;
    ignoreFilters?: boolean;
    forceOpenPivot?: boolean;
    expected: Record<string, string>;
  }[] = [
    {
      title: "dimension table for a single dimension",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        sort: [{ desc: true, name: AD_BIDS_DOMAIN_DIMENSION }],
      },
      expected: {
        view: "explore",
        expand_dim: AD_BIDS_DOMAIN_DIMENSION,
      },
    },
    {
      title: "pivot for a single dimension",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        sort: [{ desc: true, name: AD_BIDS_DOMAIN_DIMENSION }],
      },
      forceOpenPivot: true,
      expected: {
        view: "pivot",
        rows: "",
        cols: AD_BIDS_DOMAIN_DIMENSION,
        table_mode: "flat",
        sort_by: AD_BIDS_DOMAIN_DIMENSION,
        sort_dir: "DESC",
      },
    },
    {
      // The time dimension has no place in the dimension table.
      title: "dimension table ignores the time dimension",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [
          { name: AD_BIDS_DOMAIN_DIMENSION },
          {
            name: AD_BIDS_TIMESTAMP_DIMENSION,
            timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
          },
        ],
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
      },
      expected: {
        measures: AD_BIDS_BID_PRICE_MEASURE,
        sort_by: AD_BIDS_BID_PRICE_MEASURE,
        sort_dir: "DESC",
        sort_type: "value",
        view: "explore",
        expand_dim: AD_BIDS_DOMAIN_DIMENSION,
      },
    },
    {
      title: "pivot sorted by the time dimension",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [
          { name: AD_BIDS_DOMAIN_DIMENSION },
          {
            name: AD_BIDS_TIMESTAMP_DIMENSION,
            timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
          },
        ],
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        sort: [{ desc: true, name: AD_BIDS_TIMESTAMP_DIMENSION }],
      },
      forceOpenPivot: true,
      expected: {
        view: "pivot",
        rows: "",
        cols: `${AD_BIDS_DOMAIN_DIMENSION},time.week,${AD_BIDS_BID_PRICE_MEASURE}`,
        table_mode: "flat",
        sort_by: "time.week",
        sort_dir: "DESC",
      },
    },
    {
      // The explore has no column for a measure's previous value.
      title: "time dimension details drop a sort by a previous value",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        measures: [
          { name: AD_BIDS_BID_PRICE_MEASURE },
          {
            name: `${AD_BIDS_BID_PRICE_MEASURE}__previous`,
            comparisonValue: { measure: AD_BIDS_BID_PRICE_MEASURE },
          },
        ],
        sort: [{ desc: true, name: `${AD_BIDS_BID_PRICE_MEASURE}__previous` }],
      },
      expected: {
        measures: AD_BIDS_BID_PRICE_MEASURE,
        view: "tdd",
        measure: AD_BIDS_BID_PRICE_MEASURE,
      },
    },
    {
      title: "ignores filters",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        where: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo", "Google"]),
        ]),
      },
      ignoreFilters: true,
      expected: {
        measures: AD_BIDS_BID_PRICE_MEASURE,
        view: "explore",
        expand_dim: AD_BIDS_DOMAIN_DIMENSION,
      },
    },
    {
      // The request only carries the expression; the saved filter marks the In List filter.
      title: "keeps an In List filter from the saved state",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        where: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo", "Google"]),
          createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["yahoo.com"]),
        ]),
      },
      urlState: `?view=explore&f=${encodeURIComponent(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN LIST ('Yahoo','Google') AND ${AD_BIDS_DOMAIN_DIMENSION} IN ('yahoo.com')`,
      )}`,
      expected: {
        view: "explore",
        f: `${AD_BIDS_PUBLISHER_DIMENSION} IN LIST ('Yahoo','Google') AND ${AD_BIDS_DOMAIN_DIMENSION} IN ('yahoo.com')`,
      },
    },
    {
      // Reports and alerts embed the ephemeral definition in the request itself.
      title: "recovers ephemeral measures from the request",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        measures: [
          {
            name: "bid_price_per_impression",
            expression: {
              expression: `${AD_BIDS_BID_PRICE_MEASURE} / ${AD_BIDS_IMPRESSIONS_MEASURE}`,
              displayName: "Bid price per impression",
            },
          },
        ],
      },
      expected: {
        [ExploreStateURLParams.EphemeralMeasures]: toEphemeralMeasuresParam([
          {
            name: "bid_price_per_impression",
            displayName: "Bid price per impression",
            expression: `${AD_BIDS_BID_PRICE_MEASURE} / ${AD_BIDS_IMPRESSIONS_MEASURE}`,
          },
        ]),
        measures: "bid_price_per_impression",
        view: "tdd",
        measure: "bid_price_per_impression",
      },
    },
    {
      // Midnight in New York resolves to 05:00Z, which is written back as midnight in the zone.
      title: "resolved time range is written as wall-clock times in the zone",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        timeRange: { expression: "7D", timeZone: "America/New_York" },
        comparisonTimeRange: { isoDuration: "P7D" },
      },
      expected: {
        tr: "2024-01-01T00:00:00.000Z,2024-01-08T00:00:00.000Z",
        tz: "America/New_York",
        compare_tr: "rill-PP",
        view: "explore",
        expand_dim: AD_BIDS_DOMAIN_DIMENSION,
      },
    },
    {
      // The default time dimension is left out, so the url stays the same as without one.
      // A saved `td` is cleared, since the query is what the alert or report ran.
      title: "time range on the default time dimension",
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        timeRange: {
          expression: "7D",
          timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
        },
      },
      // Saved state keeps its own view, so only the time params are written.
      urlState: `?${ExploreStateURLParams.TimeDimension}=${AD_BIDS_OFFSET_TIMESTAMP_DIMENSION}`,
      expected: {
        tr: "2024-01-01T05:00:00.000Z,2024-01-08T05:00:00.000Z",
        tz: "UTC",
      },
    },
  ];

  for (const {
    title,
    req,
    urlState,
    ignoreFilters,
    forceOpenPivot,
    expected,
  } of TestCases) {
    it(title, async () => {
      const urlParams = await convertAggregationRequestToUrlParams({
        client,
        req,
        urlState,
        ignoreFilters,
        forceOpenPivot,
        executionTime: "2024-01-08T05:00:00Z",
      });
      expect(Object.fromEntries(urlParams.entries())).toEqual(expected);
      expectParsesWithoutErrors(urlParams);
    });
  }

  // Alerts with comparison criteria, such as "change vs previous day", have no dashboard equivalent for the criteria.
  // The mapper runs the saved query itself to find the matching dimension values and pins them as an `in` filter.
  it("converts comparison criteria into a toplist filter", async () => {
    const deltaMeasure =
      AD_BIDS_BID_PRICE_MEASURE + ComparisonDeltaAbsoluteSuffix;
    mocks.mockMetricsViewAggregation(
      // The saved query is sent under the request's `metricsView` field.
      // Any other key for it fails the strict proto decode before the request is sent.
      new RegExp(
        `(?=[\\s\\S]*"metricsView":"${AD_BIDS_METRICS_NAME}")(?=[\\s\\S]*"having")`,
      ),
      {
        schema: { fields: [{ name: AD_BIDS_DOMAIN_DIMENSION }] },
        data: [
          { [AD_BIDS_DOMAIN_DIMENSION]: "yahoo.com" },
          { [AD_BIDS_DOMAIN_DIMENSION]: "google.com" },
        ],
      },
    );

    const urlParams = await convertAggregationRequestToUrlParams({
      client,
      req: {
        metricsView: AD_BIDS_METRICS_NAME,
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        measures: [
          { name: AD_BIDS_BID_PRICE_MEASURE },
          {
            name: deltaMeasure,
            comparisonDelta: { measure: AD_BIDS_BID_PRICE_MEASURE },
          },
        ],
        having: createAndExpression([
          createBinaryExpression(deltaMeasure, V1Operation.OPERATION_GT, 100),
        ]),
        timeRange: { isoDuration: "P1D" },
        comparisonTimeRange: { isoDuration: "P1D", isoOffset: "P1D" },
      },
      executionTime: "2024-01-08T05:00:00Z",
    });

    expect(urlParams.get(ExploreStateURLParams.Filters)).toBe(
      `${AD_BIDS_DOMAIN_DIMENSION} IN ('yahoo.com','google.com')`,
    );
    expect(urlParams.get(ExploreStateURLParams.ComparisonTimeRange)).toBe(
      "rill-PP",
    );
    expectParsesWithoutErrors(urlParams);
  });

  describe("time range on a time dimension other than the default", () => {
    const MetricsViewWithTwoTimeDimensions: V1MetricsViewSpec = {
      ...AD_BIDS_METRICS_INIT_WITH_TIME,
      dimensions: [
        ...AD_BIDS_INIT_DIMENSIONS,
        {
          name: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
          type: MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
        },
      ],
    };
    const ExploreWithOffsetTimestamp: V1ExploreSpec = {
      ...AD_BIDS_EXPLORE_INIT,
      dimensions: [
        ...AD_BIDS_EXPLORE_INIT.dimensions!,
        AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
      ],
    };
    const ExploreWithOffsetTimestampName = "AdBids_explore_offset_timestamp";

    beforeEach(() => {
      mocks.mockMetricsExplore(
        ExploreWithOffsetTimestampName,
        MetricsViewWithTwoTimeDimensions,
        ExploreWithOffsetTimestamp,
      );
      mocks.mockMetricsExplore(
        AD_BIDS_EXPLORE_NAME,
        MetricsViewWithTwoTimeDimensions,
        AD_BIDS_EXPLORE_INIT,
      );
    });

    function getMappedUrlForExplore(exploreName: string) {
      return getMappedExploreUrl(
        client,
        {
          type: "query",
          request: {
            exploreName,
            queryName: "MetricsViewAggregation",
            queryArgsJson: JSON.stringify({
              metricsView: AD_BIDS_METRICS_NAME,
              dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
              timeRange: {
                expression: "7D",
                timeDimension: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
              },
            } satisfies V1MetricsViewAggregationRequest),
            executionTime: "2024-01-08T05:00:00Z",
          },
          options: {},
        },
        { organization: "org", project: "project" },
      );
    }

    it("opens the explore on the query's time dimension", async () => {
      const url = await getMappedUrlForExplore(ExploreWithOffsetTimestampName);
      const urlParams = new URL(url, "http://localhost").searchParams;
      expect(Object.fromEntries(urlParams.entries())).toEqual({
        tr: "2024-01-01T05:00:00.000Z,2024-01-08T05:00:00.000Z",
        tz: "UTC",
        td: AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
        view: "explore",
        expand_dim: AD_BIDS_DOMAIN_DIMENSION,
      });

      const { partialExploreState, errors } =
        convertURLSearchParamsToExploreState(
          urlParams,
          MetricsViewWithTwoTimeDimensions,
          ExploreWithOffsetTimestamp,
        );
      expect(errors).toEqual([]);
      expect(partialExploreState.selectedTimeDimension).toBe(
        AD_BIDS_OFFSET_TIMESTAMP_DIMENSION,
      );
    });

    // The explore would drop the time dimension and show the range on the default one.
    it("fails when the explore does not expose the time dimension", async () => {
      await expect(
        getMappedUrlForExplore(AD_BIDS_EXPLORE_NAME),
      ).rejects.toThrow(
        `time range on "${AD_BIDS_OFFSET_TIMESTAMP_DIMENSION}"`,
      );
    });
  });

  describe("metrics view without a time dimension", () => {
    // The runtime ignores it too, so the explore still shows what the query saw.
    it("ignores a time range without a time dimension", async () => {
      const urlParams = await convertAggregationRequestToUrlParams({
        client,
        req: {
          metricsView: NO_TIME_METRICS_NAME,
          dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
          measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
          timeRange: { isoDuration: "P1D" },
        },
      });
      expect(Object.fromEntries(urlParams.entries())).toEqual({
        measures: AD_BIDS_BID_PRICE_MEASURE,
        view: "explore",
        expand_dim: AD_BIDS_DOMAIN_DIMENSION,
      });
    });

    // The runtime filters on it, but the explore has no time controls to show it.
    it("fails on a time range on another time dimension", async () => {
      await expect(
        convertAggregationRequestToUrlParams({
          client,
          req: {
            metricsView: NO_TIME_METRICS_NAME,
            dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
            measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
            timeRange: { isoDuration: "P1D", timeDimension: "created_at" },
          },
        }),
      ).rejects.toThrow(`time range on "created_at"`);
    });
  });
});

describe("convertComparisonRequestToUrlParams", () => {
  it("dimension table for an older alert", async () => {
    const urlParams = await convertComparisonRequestToUrlParams({
      client: NoopClient,
      req: {
        metricsViewName: AD_BIDS_METRICS_NAME,
        dimension: { name: AD_BIDS_DOMAIN_DIMENSION },
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        sort: [
          {
            name: AD_BIDS_BID_PRICE_MEASURE,
            sortType:
              V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
            desc: true,
          },
        ],
        where: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
        ]),
      },
    });
    expect(Object.fromEntries(urlParams.entries())).toEqual({
      f: `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Yahoo')`,
      measures: AD_BIDS_BID_PRICE_MEASURE,
      sort_by: AD_BIDS_BID_PRICE_MEASURE,
      sort_dir: "DESC",
      // Sorting by percent of total has the same effect as sorting by base value.
      sort_type: "percent",
      view: "explore",
      expand_dim: AD_BIDS_DOMAIN_DIMENSION,
    });
    expectParsesWithoutErrors(urlParams);
  });
});

describe("convertMetricsResolverQueryToUrlParams", () => {
  const TestCases: {
    title: string;
    query: MetricsResolverQuery;
    timeDimension?: string;
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
        tz: "UTC",
      },
    },
    {
      title: "absolute time range written as wall-clock times in the zone",
      query: {
        metrics_view: AD_BIDS_METRICS_NAME,
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        time_range: {
          start: "2024-01-01T05:00:00Z",
          end: "2024-02-01T05:00:00Z",
        },
        time_zone: "America/New_York",
      },
      expected: {
        measures: AD_BIDS_IMPRESSIONS_MEASURE,
        tr: "2024-01-01T00:00:00.000Z,2024-02-01T00:00:00.000Z",
        tz: "America/New_York",
      },
    },
    {
      // Rill time has no open-ended absolute range, so the open end is the watermark.
      title: "open-ended filter on the time dimension moves to the time range",
      query: {
        metrics_view: AD_BIDS_METRICS_NAME,
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        where: {
          cond: {
            op: "gte",
            exprs: [
              { name: AD_BIDS_TIMESTAMP_DIMENSION },
              { val: "2024-01-01T05:00:00Z" },
            ],
          },
        } as MetricsResolverQuery["where"],
        time_zone: "America/New_York",
      },
      timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
      expected: {
        measures: AD_BIDS_IMPRESSIONS_MEASURE,
        tr: "2024-01-01T00:00:00.000Z to watermark",
        tz: "America/New_York",
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

  for (const { title, query, timeDimension, expected } of TestCases) {
    it(title, () => {
      const urlParams = convertMetricsResolverQueryToUrlParams(
        { query },
        timeDimension,
      );
      expect(Object.fromEntries(urlParams.entries())).toEqual(expected);
      expectParsesWithoutErrors(urlParams);
    });
  }
});
