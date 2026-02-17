import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { goto } from "$app/navigation";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { getMetricsViewAndExploreSpecsQueryOptions } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { maybeGetExplorePageUrlSearchParams } from "@rilldata/web-common/features/explore-mappers/utils.ts";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client.ts";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import { page } from "$app/stores";
import { getUrlForExplore } from "@rilldata/web-common/features/explore-mappers/generate-explore-link.ts";
import { getQueryFromUrl } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";

const LEGACY_DASHBOARD_CITATION_URL_PATHNAME_REGEX = /\/-\/open-query\/?$/;
const DASHBOARD_CITATION_URL_PATHNAME_REGEX = /\/-\/ai\/.*?\/call\/.*?\/?$/;

/**
 * Adds a click handler to the given node, that intercepts clicks to links and uses svelte's goto.
 * Links added directly to the node are not intercepted by svelte so we need to manually intercept them.
 * Also adds a check to make sure url is actually a local link before using goto.
 */
export function enhanceCitationLinks(
  node: HTMLElement,
  conversation: Conversation,
) {
  const isEmbedded = EmbedStore.isEmbedded();
  const mapperStore =
    getMetricsResolverQueryToUrlParamsMapperStore(conversation);

  function handleClick(e: MouseEvent) {
    if (!e.target || !(e.target instanceof HTMLElement)) return; // typesafety
    if (e.metaKey || e.ctrlKey) {
      if (isEmbedded) {
        // Prevent cmd/ctrl+click in embedded context. The iframe href used for embedding cannot be opened as is.
        e.preventDefault();
      }
      return;
    }

    const href = e.target.getAttribute("href") ?? "";
    const parsedUrl = URL.parse(href);

    const mapper = get(mapperStore).data;
    const mappedHref = parsedUrl && mapper ? mapper(parsedUrl) : href;

    const isLocalLink = window.location.origin === parsedUrl?.origin;
    const isPartialLink = mappedHref.startsWith("/");
    const shouldUseGoto = isLocalLink || isPartialLink;
    if (!shouldUseGoto) return;

    e.preventDefault();
    void goto(mappedHref);
  }

  node.addEventListener("click", handleClick);
  return {
    destroy() {
      node.removeEventListener("click", handleClick);
    },
  };
}

/**
 * Creates a store that contains a mapper function that maps a metrics resolver query to url params.
 * Uses {@link getMetricsViewAndExploreSpecsQueryOptions} to get ListResources query to get metrics view and explore specs.
 * Calls {@link mapMetricsResolverQueryToDashboard} to get partial explore state from a metrics resolver query.
 * Then calls {@link convertPartialExploreStateToUrlParams} to convert the partial explore to url params.
 */
function getMetricsResolverQueryToUrlParamsMapperStore(
  conversation: Conversation,
): Readable<{
  error: ErrorType<RpcStatus> | null;
  isLoading: boolean;
  data?: (url: URL) => string;
}> {
  const resourcesQuery = createQuery(
    getMetricsViewAndExploreSpecsQueryOptions(),
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

      const { metricsViewSpecsMap, exploreSpecsMap, exploreForMetricViewsMap } =
        resourcesResp.data;
      const messages = conversationResp.data?.messages ?? [];

      const mapper = (url: URL): string => {
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
          const isNewCitationUrl = DASHBOARD_CITATION_URL_PATHNAME_REGEX.test(
            url.pathname,
          );
          if (!isNewCitationUrl) return url.href;

          const callId = url.pathname.split("/").pop();
          if (!callId) return url.href;
          const callMessage = messages.find((m) => m.id === callId);
          if (!callMessage?.content?.[0]?.toolCall?.input) return url.href;
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
          partialExploreState,
          metricsViewSpec,
          exploreSpec,
        );
        if (!urlSearchParams) return url.href;

        const exploreUrl = getUrlForExplore(
          exploreName,
          page.params.organization,
          page.params.project,
        );
        urlSearchParams.forEach((value, key) => {
          exploreUrl.searchParams.set(key, value);
        });

        return exploreUrl.href;
      };

      return {
        isLoading: false,
        error: null,
        data: mapper,
      };
    },
  );
}
