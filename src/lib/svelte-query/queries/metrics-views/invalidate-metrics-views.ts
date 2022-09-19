import { MetaId } from "$lib/svelte-query/queries/metrics-views/metrics-views-metadata";
import { TimeSeriesId } from "$lib/svelte-query/queries/metrics-views/metrics-views-time-series";
import { TopListId } from "$lib/svelte-query/queries/metrics-views/metrics-views-top-list";
import { TotalsId } from "$lib/svelte-query/queries/metrics-views/metrics-views-totals";
import type { QueryClient } from "@sveltestack/svelte-query";

// invalidation helpers

export const invalidateMetricsView = async (
  queryClient: QueryClient,
  metricViewId: string
) => {
  // wait for meta to be invalidated
  await queryClient.refetchQueries([MetaId, metricViewId]);
  // invalidateMetricsViewData(queryClient, metricViewId);
};

export const invalidateMetricsViewData = (
  queryClient: QueryClient,
  metricViewId: string
) => {
  queryClient.refetchQueries([TopListId, metricViewId]);
  queryClient.refetchQueries([TotalsId, metricViewId]);
  queryClient.refetchQueries([TimeSeriesId, metricViewId]);
};
