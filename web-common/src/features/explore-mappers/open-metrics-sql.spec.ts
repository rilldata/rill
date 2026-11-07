import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { beforeEach, describe, expect, it, vi } from "vitest";

const { convert, generateExploreLink, listResources, queryResolver } =
  vi.hoisted(() => ({
    convert: vi.fn(),
    generateExploreLink: vi.fn(),
    listResources: vi.fn(),
    queryResolver: vi.fn(),
  }));

vi.mock("@rilldata/web-common/runtime-client", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  runtimeServiceListResources: listResources,
  runtimeServiceQueryResolver: queryResolver,
}));

vi.mock("@rilldata/web-common/runtime-client/v2/gen/query-service", () => ({
  queryServiceConvertMetricsSQLToQuery: convert,
}));
vi.mock("@rilldata/web-common/features/explore-mappers/open-query.ts", () => ({
  getExploreSpecs: () =>
    Promise.resolve({
      metricsViewSpec: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      exploreSpec: AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION,
    }),
  generateExploreLink,
}));

import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getExploreUrlForMetricsSQL } from "./open-metrics-sql";

const metricsView =
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS_TIME_DIMENSION.metricsView;
const having = {
  cond: {
    op: "gt",
    exprs: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }, { val: 10 }],
  },
};
const args = {
  client: {} as RuntimeClient,
  sql: "SELECT publisher, impressions FROM AdBids HAVING impressions > 10",
  exploreName: "AdBids_explore",
  executionTime: "2022-02-01T07:00:00Z",
  organization: "org",
  project: "proj",
};

beforeEach(() => {
  queryClient.clear();
  listResources.mockReset();
  queryResolver.mockReset();
  queryResolver.mockResolvedValue({
    data: [{ [AD_BIDS_PUBLISHER_DIMENSION]: "Google", impressions: 12 }],
  });
  convert.mockReset();
  generateExploreLink.mockReset();
  generateExploreLink.mockResolvedValue(
    "http://localhost/org/proj/explore/AdBids_explore?f=x",
  );
});

describe("getExploreUrlForMetricsSQL", () => {
  it("compiles the SQL at the execution time and opens the explore", async () => {
    convert.mockResolvedValue({
      query: {
        metrics_view: metricsView,
        dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        having,
      },
    });
    const url = await getExploreUrlForMetricsSQL(args);
    expect(convert.mock.calls[0][1]).toEqual({
      sql: args.sql,
      executionTime: "2022-02-01T07:00:00Z",
    });
    expect(url).toBe("http://localhost/org/proj/explore/AdBids_explore?f=x");
    // The HAVING reaches the explore as the values that pass it, found by running the SQL
    // as the alert did: a measure filter would compute it without the WHERE.
    expect(queryResolver.mock.calls[0][1]).toEqual({
      resolver: "metrics_sql",
      resolverProperties: { sql: args.sql },
      resolverArgs: { execution_time: "2022-02-01T07:00:00Z" },
    });
    const exploreState = generateExploreLink.mock.calls[0][1];
    expect(exploreState.whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ]),
    );
  });

  it("opens the shared explore with an emailed link's token", async () => {
    convert.mockResolvedValue({
      query: {
        metrics_view: metricsView,
        dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
        measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        where: {
          cond: {
            op: "eq",
            exprs: [{ name: AD_BIDS_PUBLISHER_DIMENSION }, { val: "Yahoo" }],
          },
        },
        having,
      },
    });
    const url = await getExploreUrlForMetricsSQL({ ...args, token: "tkn" });
    expect(new URL(url).pathname).toBe(
      "/org/proj/-/share/tkn/explore/AdBids_explore",
    );
    // The token already filters by the WHERE; the HAVING it doesn't carry stays.
    const exploreState = generateExploreLink.mock.calls[0][1];
    expect(exploreState.whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ]),
    );
  });

  it("refuses an explore on another metrics view", async () => {
    convert.mockResolvedValue({
      query: { metrics_view: "other", measures: [] },
    });
    await expect(getExploreUrlForMetricsSQL(args)).rejects.toThrow(
      "does not show metrics view",
    );
  });

  it("refuses a HAVING the explore can't show, rather than drop it", async () => {
    // Without a dimension, or over several, values of one dimension aren't the same rows.
    for (const dimensions of [
      [],
      [
        { name: AD_BIDS_PUBLISHER_DIMENSION },
        { name: AD_BIDS_DOMAIN_DIMENSION },
      ],
      [{ name: "not_in_the_explore" }],
    ]) {
      convert.mockResolvedValue({
        query: {
          metrics_view: metricsView,
          dimensions,
          measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
          having,
        },
      });
      await expect(getExploreUrlForMetricsSQL(args)).rejects.toThrow("HAVING");
    }
    expect(queryResolver).not.toHaveBeenCalled();
  });

  describe("without an explore named", () => {
    const explore = (name: string, mv = metricsView) => ({
      meta: { name: { name } },
      explore: { state: { validSpec: { metricsView: mv } } },
    });
    beforeEach(() => {
      convert.mockResolvedValue({
        query: {
          metrics_view: metricsView,
          dimensions: [{ name: AD_BIDS_PUBLISHER_DIMENSION }],
          measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
        },
      });
    });

    it("opens the only explore on the metrics view", async () => {
      listResources.mockResolvedValue({
        resources: [explore("other_explore", "other"), explore("the_one")],
      });
      await getExploreUrlForMetricsSQL({ ...args, exploreName: "" });
      expect(generateExploreLink.mock.calls[0][4]).toBe("the_one");
    });

    it("doesn't guess among several", async () => {
      listResources.mockResolvedValue({
        resources: [explore("one"), explore("two")],
      });
      await expect(
        getExploreUrlForMetricsSQL({ ...args, exploreName: "" }),
      ).rejects.toThrow("found 2");
    });
  });
});
