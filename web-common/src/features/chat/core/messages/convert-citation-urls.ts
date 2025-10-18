import { useMetricsViewTimeRangeFromExplore } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { mapMetricsResolverQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-metrics-resolver-query-to-dashboard.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import { marked } from "marked";
import { derived, type Readable } from "svelte/store";

export function getCitationUrlConverter(
  mapper: MetricsResolverQueryToUrlParamsMapper,
) {
  return (text: string): string | Promise<string> => {
    marked.use({
      renderer: {
        link: (tokens) => {
          const url = new URL(tokens.href);
          const queryJSON = url.searchParams.get("query");
          if (!queryJSON) return false;
          const [isValid, urlParams] = mapper(queryJSON);
          if (!isValid) return false;

          return `<a href="?${urlParams.toString()}">${tokens.text}</a>`;
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
 * Creates a store that contains a metrics resolver query to url params mapper.
 * Calls {@link mapMetricsResolverQueryToDashboard} to get partial explore state from a metrics resolver query.
 * Then calls {@link convertPartialExploreStateToUrlParams} to convert the partial explore to url params.
 *
 * @param instanceId
 * @param exploreName
 */
export function getMetricsResolverQueryToUrlParamsMapperStore(
  instanceId: string,
  exploreName: string,
): Readable<{
  error?: Error;
  isLoading: boolean;
  data?: MetricsResolverQueryToUrlParamsMapper;
}> {
  return derived(
    [
      useExploreValidSpec(instanceId, exploreName, undefined, queryClient),
      useMetricsViewTimeRangeFromExplore(
        instanceId,
        exploreName,
        {},
        queryClient,
      ),
    ],
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

      const metricsViewSpec = validSpecResp.data?.metricsView;
      const exploreSpec = validSpecResp.data?.explore;
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
