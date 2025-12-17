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

const DASHBOARD_CITATION_URL_PATHNAME_REGEX = /\/-\/open-query\/?$/;

/**
 * Adds a click handler to the given node, that intercepts clicks to links and uses svelte's goto.
 * Links added directly to the node are not intercepted by svelte so we need to manually intercept them.
 * Also adds a check to make sure url is actually a local link before using goto.
 * @param node
 */
export function enhanceCitationLinks(node: HTMLElement) {
  const isEmbedded = EmbedStore.isEmbedded();
  const mapperStore = getMetricsResolverQueryToUrlParamsMapperStore();

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
function getMetricsResolverQueryToUrlParamsMapperStore(): Readable<{
  error: ErrorType<RpcStatus> | null;
  isLoading: boolean;
  data?: (url: URL) => string;
}> {
  const resourcesQuery = createQuery(
    getMetricsViewAndExploreSpecsQueryOptions(),
    queryClient,
  );

  return derived([page, resourcesQuery], ([page, resourcesResp]) => {
    if (resourcesResp.error || resourcesResp.isLoading || !resourcesResp.data) {
      return {
        error: resourcesResp.error,
        isLoading: resourcesResp.isLoading,
      };
    }

    const { metricsViewSpecsMap, exploreSpecsMap, exploreForMetricViewsMap } =
      resourcesResp.data;

    const mapper = (url: URL): string => {
      const queryJSON = url.searchParams.get("query");
      // If the url does not have a query arg, do not change the link.
      if (
        !queryJSON ||
        !DASHBOARD_CITATION_URL_PATHNAME_REGEX.test(url.pathname)
      ) {
        return url.href;
      }

      let query: MetricsResolverQuery;
      try {
        query = JSON.parse(queryJSON) as MetricsResolverQuery;
      } catch {
        return url.href;
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
  });
}
