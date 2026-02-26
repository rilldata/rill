import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";
import type { RpcStatus, V1Message } from "@rilldata/web-common/runtime-client";
import {
  getMetricsViewAndExploreSpecsQueryOptions,
  type MetricsViewAndExploreSpecs,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { getQueryFromUrl } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { maybeGetExplorePageUrlSearchParams } from "@rilldata/web-common/features/explore-mappers/utils.ts";
import { getUrlForExplore } from "@rilldata/web-common/features/explore-mappers/generate-explore-link.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import { page } from "$app/stores";
import type { Page } from "@sveltejs/kit";
import {
  MessageType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";

const LEGACY_DASHBOARD_CITATION_URL_PATHNAME_REGEX = /\/-\/open-query\/?$/;
const DASHBOARD_CITATION_URL_PATHNAME_REGEX =
  /\/-\/ai\/[^/]+\/message\/([^/]+)\/-\/open\/?$/;

/**
 * Creates a store that contains a mapper function that maps a metrics resolver query to url with explore params.
 * Uses {@link getMetricsViewAndExploreSpecsQueryOptions} to get ListResources query to get metrics view and explore specs.
 * Calls {@link mapMetricsResolverQueryToDashboard} to get partial explore state from a metrics resolver query.
 * Then calls {@link convertPartialExploreStateToUrlParams} to convert the partial explore to url with explore params.
 */
export function getMetricsResolverQueryToUrlMapperStore(
  conversation: Conversation,
): Readable<{
  error: Error | null;
  isLoading: boolean;
  data?: (url: URL) => string;
}> {
  const resourcesQuery = createQuery(
    getMetricsViewAndExploreSpecsQueryOptions(conversation.runtimeClient),
    queryClient,
  );
  const conversationQuery = conversation.getConversationQuery();

  return derived(
    [page, resourcesQuery, conversationQuery],
    ([page, resourcesResp, conversationResp]) => {
      const isError = conversationResp.error || resourcesResp.error;
      const isLoading = conversationResp.isLoading || resourcesResp.isLoading;
      const hasData = conversationResp.data?.messages && resourcesResp.data;
      if (isError || isLoading || !hasData) {
        return {
          error: resourcesResp.error,
          isLoading: resourcesResp.isLoading,
        };
      }

      const metricsViewAndExploreSpecs = resourcesResp.data;
      const messages = conversationResp.data?.messages ?? [];

      const mapper = (url: URL): string =>
        mapMetricsResolverQueryToUrl(
          conversation.runtimeClient.instanceId,
          url,
          page,
          metricsViewAndExploreSpecs,
          messages,
        );

      return {
        isLoading: false,
        error: null,
        data: mapper,
      };
    },
  );
}

export function mapMetricsResolverQueryToUrl(
  instanceId: string,
  url: URL,
  pageState: Page,
  {
    metricsViewSpecsMap,
    exploreForMetricViewsMap,
    exploreSpecsMap,
  }: MetricsViewAndExploreSpecs,
  messages: V1Message[],
) {
  let query: MetricsResolverQuery;
  const isLegacyCitationUrl =
    LEGACY_DASHBOARD_CITATION_URL_PATHNAME_REGEX.test(url.pathname) &&
    url.searchParams.has("query");
  if (isLegacyCitationUrl) {
    try {
      query = getQueryFromUrl(url);
    } catch {
      return url.href;
    }
  } else {
    const citationUrlPathnameMatch = DASHBOARD_CITATION_URL_PATHNAME_REGEX.exec(
      url.pathname,
    );
    if (!citationUrlPathnameMatch?.[1]) return url.href;

    const callId = citationUrlPathnameMatch[1];
    if (!callId) return url.href;
    const callMessage = messages.find((m) => m.id === callId);
    if (!callMessage?.content?.[0]?.toolCall?.input) return url.href;
    const isMetricsQueryCall =
      callMessage.type === MessageType.CALL &&
      callMessage.tool === ToolName.QUERY_METRICS_VIEW;
    if (!isMetricsQueryCall) return url.href;
    query = callMessage.content?.[0]?.toolCall?.input;
  }

  const metricsViewName = query.metrics_view ?? "";
  const metricsViewSpec = metricsViewSpecsMap.get(metricsViewName);
  const exploreName = exploreForMetricViewsMap.get(metricsViewName);
  if (!metricsViewSpec || !exploreName) return url.href;
  const exploreSpec = exploreSpecsMap.get(exploreName);
  if (!exploreSpec) return url.href;

  const partialExploreState = mapMetricsResolverQueryToDashboard(
    metricsViewSpec,
    exploreSpec,
    query,
  );

  const urlSearchParams = maybeGetExplorePageUrlSearchParams(
    instanceId,
    partialExploreState,
    metricsViewSpec,
    exploreSpec,
  );
  if (!urlSearchParams) return url.href;

  const exploreUrl = getUrlForExplore(
    exploreName,
    pageState.params.organization,
    pageState.params.project,
  );
  urlSearchParams.forEach((value, key) => {
    exploreUrl.searchParams.set(key, value);
  });

  return exploreUrl.href;
}
