/**
 * TODO: Instead of writing this file by hand, a better approach would be to use an OpenAPI spec and
 * autogenerate `svelte-query`-specific client code. One such tool is: https://orval.dev/guides/svelte-query
 */

import type {
  MetricsViewTotalsRequest,
  MetricsViewTotalsResponse,
} from "$common/rill-developer-service/MetricsViewActions";
import { fetchUrl } from "$lib/svelte-query/queries/fetch-url";
import { useQuery } from "@sveltestack/svelte-query";

// POST /api/v1/metrics-views/{view-name}/totals

export const getMetricsViewTotals = async (
  metricViewId: string,
  request: MetricsViewTotalsRequest
): Promise<MetricsViewTotalsResponse> => {
  return fetchUrl(`metrics-views/${metricViewId}/totals`, "POST", request);
};

export const TotalsId = `v1/metrics-view/totals`;
export const getTotalsQueryKey = (
  metricViewId: string,
  request: MetricsViewTotalsRequest
) => {
  return [TotalsId, metricViewId, request];
};

export const useTotalsQuery = (
  metricViewId: string,
  request: MetricsViewTotalsRequest
) => {
  const totalsQueryKey = getTotalsQueryKey(metricViewId, request);
  const totalsQueryFn = () => getMetricsViewTotals(metricViewId, request);

  return useQuery<MetricsViewTotalsResponse, Error>(
    totalsQueryKey,
    totalsQueryFn,
    {
      staleTime: 30 * 1000,
      enabled: !!(metricViewId && request.measures && request.time),
    }
  );
};
