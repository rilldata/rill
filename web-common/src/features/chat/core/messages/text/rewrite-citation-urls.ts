import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { createQuery } from "@tanstack/svelte-query";
import { marked } from "marked";
import { derived, type Readable } from "svelte/store";

const CITATION_URL_PATHNAME_REGEX = /\/-\/open-query\/?$/;

/**
 * Creates a rewriter that adds hooks to the ` marked ` library to rewrite citation urls to directly apply the url params.
 * A citation url will be of the form `.../-/open-query?query...`
 *
 * @param mapper
 */
export function getCitationUrlRewriter(
  mapper: MetricsResolverQueryToUrlParamsMapper | undefined,
) {
  return (text: string): string | Promise<string> => {
    marked.use({
      renderer: {
        link: (tokens) => {
          const url = new URL(tokens.href);
          // If the url is not a citation url, do not change the link.
          if (!CITATION_URL_PATHNAME_REGEX.test(url.pathname)) return false;

          // If mapper was not provided, remove citation urls.
          // This happens when something went wrong getting the mapper in an explore chat.
          // Non-explore chat will not go throught his code path at all.
          if (!mapper) return "";

          const queryJSON = url.searchParams.get("query");
          // If the url does not have a query arg, do not change the link.
          if (!queryJSON) return false;

          const [isValid, urlParams] = mapper(queryJSON);
          // Mapping failed for some reason, remove the link since it is probably invalid.
          if (!isValid) return "";

          return `<a data-sveltekit-preload-data="off" href="?${urlParams.toString()}">${tokens.text}</a>`;
        },
      },
    });

    // Run marked
    return marked.parse(text);
  };
}

type MetricsResolverQueryToUrlParamsMapper = (
  queryJSON: string,
) => [boolean, URLSearchParams];

/**
 * Creates a store that contains a mapper function that maps a metrics resolver query to url params.
 * Calls {@link mapMetricsResolverQueryToDashboard} to get partial explore state from a metrics resolver query.
 * Then calls {@link convertPartialExploreStateToUrlParams} to convert the partial explore to url params.
 */
export function getMetricsResolverQueryToUrlParamsMapperStore(
  exploreNameStore: Readable<string>,
): Readable<{
  error?: Error;
  isLoading: boolean;
  data?: MetricsResolverQueryToUrlParamsMapper;
}> {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );
  const timeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(exploreNameStore),
  );

  return derived(
    [validSpecQuery, timeRangeQuery],
    ([validSpecResp, timeRangeResp]) => {
      const error = (validSpecResp.error || timeRangeResp.error) as
        | Error
        | undefined;
      const isLoading = validSpecResp.isLoading || timeRangeResp.isLoading;
      if (error || isLoading) {
        return {
          error,
          isLoading,
        };
      }

      const metricsViewSpec = validSpecResp.data?.metricsViewSpec;
      const exploreSpec = validSpecResp.data?.exploreSpec;
      if (!metricsViewSpec || !exploreSpec) {
        return {
          error: new Error("Failed to load metrics view or explore spec"),
          isLoading: false,
        };
      }

      const timeRangeSummary = timeRangeResp.data?.timeRangeSummary;

      const mapper = (queryJSON: string): [boolean, URLSearchParams] => {
        let query: MetricsResolverQuery;
        try {
          query = JSON.parse(queryJSON) as MetricsResolverQuery;
        } catch {
          return [false, new URLSearchParams()];
        }

        const partialExploreState = mapMetricsResolverQueryToDashboard(
          metricsViewSpec,
          exploreSpec,
          timeRangeSummary,
          query,
        );
        const timeControlState = getTimeControlState(
          metricsViewSpec,
          exploreSpec,
          timeRangeSummary,
          partialExploreState,
        );

        const urlSearchParams = convertPartialExploreStateToUrlParams(
          exploreSpec,
          partialExploreState,
          timeControlState,
        );

        return [true, urlSearchParams];
      };

      return {
        isLoading: false,
        data: mapper,
      };
    },
  );
}
