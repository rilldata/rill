import { combineChartDataQuery } from "@rilldata/web-common/features/components/charts/query-util";
import type { V1MetricsViewAggregationResponse } from "@rilldata/web-common/runtime-client";
import { createQuery, QueryClient } from "@tanstack/svelte-query";
import { derived, get, writable } from "svelte/store";
import { describe, expect, it } from "vitest";

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
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
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

  it("does not report fetching when the data query cannot run", () => {
    // For example, a spec without a measure:
    // the provider never enables its data query, nor the top-N query it would wait on.
    const queryClient = new QueryClient();
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
