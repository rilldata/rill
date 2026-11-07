import { combineChartDataQuery } from "@rilldata/web-common/features/components/charts/query-util";
import type { V1MetricsViewAggregationResponse } from "@rilldata/web-common/runtime-client";
import {
  createQuery,
  keepPreviousData,
  QueryClient,
} from "@tanstack/svelte-query";
import { derived, get, writable } from "svelte/store";
import { describe, expect, it } from "vitest";

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
}

function deferredResponse() {
  let resolve!: (value: V1MetricsViewAggregationResponse) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<V1MetricsViewAggregationResponse>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

// TanStack Query delivers results to its stores from a `setTimeout` callback.
async function flush() {
  for (let i = 0; i < 3; i++) {
    await new Promise((resolve) => setTimeout(resolve, 0));
  }
}

// Mirrors a chart provider:
// the data query waits for the component to be visible and for its top-N query.
function createChartQueries() {
  const queryClient = createTestQueryClient();
  const visible = writable(false);
  const topN = deferredResponse();
  const data = deferredResponse();

  const topNQuery = createQuery(
    derived(visible, ($visible) => ({
      queryKey: ["top-n"],
      queryFn: () => topN.promise,
      enabled: $visible,
    })),
    queryClient,
  );
  const dataQuery = createQuery(
    derived([visible, topNQuery], ([$visible, $topNQuery]) => ({
      queryKey: ["data", $topNQuery.data],
      queryFn: () => data.promise,
      enabled: $visible && $topNQuery.data !== undefined,
    })),
    queryClient,
  );
  const chartQuery = combineChartDataQuery(dataQuery, [topNQuery], visible);
  // Subscribing mounts the query observers, as the chart component does.
  const unsubscribe = chartQuery.subscribe(() => {});

  return { visible, topN, data, dataQuery, chartQuery, unsubscribe };
}

// Another chart making the same top-N request, which runs once it scrolls into view.
function createOtherChartTopNQuery(
  queryClient: QueryClient,
  queryFn: () => Promise<V1MetricsViewAggregationResponse>,
) {
  const visible = writable(false);
  const topNQuery = createQuery(
    derived(visible, ($visible) => ({
      queryKey: ["top-n"],
      queryFn,
      enabled: $visible,
    })),
    queryClient,
  );
  return { visible, topNQuery };
}

describe("combineChartDataQuery", () => {
  it("reports fetching until the data query has run", async () => {
    const { visible, topN, data, dataQuery, chartQuery, unsubscribe } =
      createChartQueries();

    // Off-screen: nothing runs yet.
    expect(get(chartQuery)).toMatchObject({
      data: undefined,
      isFetching: true,
    });

    // On-screen: the top-N query runs while the data query waits, disabled.
    visible.set(true);
    await flush();
    expect(get(dataQuery)).toMatchObject({
      data: undefined,
      isFetching: false,
    });
    expect(get(chartQuery)).toMatchObject({
      data: undefined,
      isFetching: true,
    });

    topN.resolve({ data: [{ country: "US" }] });
    await flush();
    expect(get(dataQuery).isFetching).toBe(true);
    expect(get(chartQuery)).toMatchObject({
      data: undefined,
      isFetching: true,
    });

    data.resolve({ data: [] });
    await flush();
    expect(get(chartQuery)).toEqual({
      data: { data: [] },
      isFetching: false,
      error: null,
    });

    unsubscribe();
  });

  it("reports the error of a failed dependency", async () => {
    const { visible, topN, chartQuery, unsubscribe } = createChartQueries();

    visible.set(true);
    topN.reject(new Error("top-N query failed"));
    await flush();
    expect(get(chartQuery).isFetching).toBe(false);
    expect(get(chartQuery).error?.message).toBe("top-N query failed");

    unsubscribe();
  });

  it("ignores another chart fetching a dependency this chart does not run", async () => {
    // For example, a chart with explicit color values skips its top-N query,
    // but another chart's identical top-N request shares that query's cache entry.
    const queryClient = createTestQueryClient();
    const topN = deferredResponse();
    const data = deferredResponse();
    const topNQuery = createQuery(
      { queryKey: ["top-n"], queryFn: () => topN.promise, enabled: false },
      queryClient,
    );
    const dataQuery = createQuery(
      { queryKey: ["data"], queryFn: () => data.promise },
      queryClient,
    );
    const chartQuery = combineChartDataQuery(
      dataQuery,
      [topNQuery],
      writable(true),
    );
    const otherChart = createOtherChartTopNQuery(
      queryClient,
      () => topN.promise,
    );
    const unsubscribers = [
      chartQuery.subscribe(() => {}),
      otherChart.topNQuery.subscribe(() => {}),
    ];

    data.resolve({ data: [] });
    await flush();
    otherChart.visible.set(true);
    await flush();
    expect(get(topNQuery).isFetching).toBe(true);
    expect(get(chartQuery).isFetching).toBe(false);

    topN.reject(new Error("top-N query failed"));
    await flush();
    expect(get(topNQuery).error?.message).toBe("top-N query failed");
    expect(get(chartQuery)).toEqual({
      data: { data: [] },
      isFetching: false,
      error: null,
    });

    unsubscribers.forEach((unsubscribe) => unsubscribe());
  });

  it("ignores another chart refetching a shared dependency", async () => {
    // Two charts with the same top-N request share its cache entry,
    // and the second refetches the stale entry when it scrolls into view.
    const queryClient = createTestQueryClient();
    const topNResponses = [deferredResponse(), deferredResponse()];
    let topNFetches = 0;
    const topNQueryFn = () => topNResponses[topNFetches++].promise;
    const data = deferredResponse();
    let dataFetches = 0;

    const topNQuery = createQuery(
      { queryKey: ["top-n"], queryFn: topNQueryFn },
      queryClient,
    );
    const dataQuery = createQuery(
      derived(topNQuery, ($topNQuery) => ({
        queryKey: ["data", $topNQuery.data],
        queryFn: () => {
          dataFetches++;
          return data.promise;
        },
        enabled: $topNQuery.data !== undefined,
        placeholderData: keepPreviousData,
      })),
      queryClient,
    );
    const chartQuery = combineChartDataQuery(
      dataQuery,
      [topNQuery],
      writable(true),
    );
    const otherChart = createOtherChartTopNQuery(queryClient, topNQueryFn);
    const unsubscribers = [
      chartQuery.subscribe(() => {}),
      otherChart.topNQuery.subscribe(() => {}),
    ];

    topNResponses[0].resolve({ data: [{ country: "US" }] });
    await flush();
    data.resolve({ data: [] });
    await flush();
    expect(get(chartQuery).isFetching).toBe(false);

    otherChart.visible.set(true);
    await flush();
    expect(topNFetches).toBe(2);
    expect(get(chartQuery).isFetching).toBe(false);

    topNResponses[1].resolve({ data: [{ country: "US" }] });
    await flush();
    expect(dataFetches).toBe(1);
    expect(get(chartQuery)).toEqual({
      data: { data: [] },
      isFetching: false,
      error: null,
    });

    unsubscribers.forEach((unsubscribe) => unsubscribe());
  });

  it("does not report fetching when the data query cannot run", () => {
    // For example, a spec without a measure:
    // the provider never enables its data query, nor the top-N query it would wait on.
    const queryClient = createTestQueryClient();
    const topNQuery = createQuery(
      {
        queryKey: ["top-n"],
        queryFn: () => deferredResponse().promise,
        enabled: false,
      },
      queryClient,
    );
    const dataQuery = createQuery(
      {
        queryKey: ["data"],
        queryFn: () => deferredResponse().promise,
        enabled: false,
      },
      queryClient,
    );

    const chartQuery = combineChartDataQuery(
      dataQuery,
      [topNQuery],
      writable(true),
    );
    expect(get(chartQuery)).toEqual({
      data: undefined,
      isFetching: false,
      error: null,
    });
  });
});
