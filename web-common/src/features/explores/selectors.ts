import type { CreateQueryOptions, QueryFunction } from "@rilldata/svelte-query";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import {
  convertPresetToExploreState,
  convertURLToExploreState,
} from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { getExplorePresetForWebView } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { FromURLParamViewMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createRuntimeServiceGetExplore,
  getQueryServiceMetricsViewTimeRangeQueryKey,
  getRuntimeServiceGetExploreQueryKey,
  queryServiceMetricsViewTimeRange,
  runtimeServiceGetExplore,
  type RpcStatus,
  type V1ExploreSpec,
  type V1GetExploreResponse,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1ExplorePreset,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";
import { error, redirect } from "@sveltejs/kit";

export function useExplore(
  instanceId: string,
  exploreName: string,
  queryOptions?: CreateQueryOptions<
    V1GetExploreResponse,
    ErrorType<RpcStatus>,
    V1GetExploreResponse
  >,
) {
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: queryOptions,
    },
  );
}

export type ExploreValidSpecResponse = {
  explore: V1ExploreSpec | undefined;
  metricsView: V1MetricsViewSpec | undefined;
};
export function useExploreValidSpec(
  instanceId: string,
  exploreName: string,
  queryOptions?: CreateQueryOptions<
    V1GetExploreResponse,
    ErrorType<RpcStatus>,
    ExploreValidSpecResponse
  >,
) {
  const defaultQueryOptions: CreateQueryOptions<
    V1GetExploreResponse,
    ErrorType<RpcStatus>,
    ExploreValidSpecResponse
  > = {
    select: (data) =>
      <ExploreValidSpecResponse>{
        explore: data.explore?.explore?.state?.validSpec,
        metricsView: data.metricsView?.metricsView?.state?.validSpec,
      },
    queryClient,
    enabled: !!exploreName,
  };
  return createRuntimeServiceGetExplore(
    instanceId,
    { name: exploreName },
    {
      query: {
        ...defaultQueryOptions,
        ...queryOptions,
      },
    },
  );
}

export async function fetchExploreSpec(
  instanceId: string,
  exploreName: string,
) {
  const queryParams = {
    name: exploreName,
  };
  const queryKey = getRuntimeServiceGetExploreQueryKey(instanceId, queryParams);
  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetExplore>>
  > = ({ signal }) => runtimeServiceGetExplore(instanceId, queryParams, signal);

  const response = await queryClient.fetchQuery({
    queryFn: queryFunction,
    queryKey,
    staleTime: Infinity,
  });

  const exploreResource = response.explore;
  const metricsViewResource = response.metricsView;

  if (!exploreResource?.explore) {
    throw error(404, "Explore not found");
  }
  if (!metricsViewResource?.metricsView) {
    throw error(404, "Metrics view not found");
  }

  let fullTimeRange: V1MetricsViewTimeRangeResponse | undefined = undefined;
  const metricsViewName = exploreResource.explore.state?.validSpec?.metricsView;
  if (
    metricsViewResource.metricsView.state?.validSpec?.timeDimension &&
    metricsViewName
  ) {
    fullTimeRange = await queryClient.fetchQuery({
      queryFn: () =>
        queryServiceMetricsViewTimeRange(instanceId, metricsViewName, {}),
      queryKey: getQueryServiceMetricsViewTimeRangeQueryKey(
        instanceId,
        metricsViewName,
        {},
      ),
      staleTime: Infinity,
      cacheTime: Infinity,
    });
  }

  const defaultExplorePreset = getDefaultExplorePreset(
    exploreResource.explore.state?.validSpec ?? {},
    fullTimeRange,
  );

  return {
    explore: exploreResource,
    metricsView: metricsViewResource,
    defaultExplorePreset,
  };
}

// converts the url search params to a partial explore state.
// if only the `view` param is set then it redirects to a new url with params loaded from sessionStorage for the view
export function getPartialExploreStateOrRedirect(
  exploreName: string,
  metricsViewSpec: V1MetricsViewSpec | undefined,
  exploreSpec: V1ExploreSpec | undefined,
  defaultExplorePreset: V1ExplorePreset,
  prefix: string | undefined,
  url: URL,
) {
  if (!metricsViewSpec || !exploreSpec) {
    return {
      partialExploreState: {},
      errors: [],
    };
  }

  const redirectUrl = shouldRedirectToViewWithParams(
    exploreName,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
    prefix,
    url,
  );
  if (redirectUrl) {
    throw redirect(307, redirectUrl);
  }

  // we didnt redirect so get partial state for current url
  return convertURLToExploreState(
    url.searchParams,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
  );
}

/**
 * Redirects to a view with params loaded from session storage.
 * 1. If only view param is set then load the params from session storage and build a new url.
 * 2. If no param is set then load the params for the default view from session storage and build a new url.
 * If the url from above either of the above is not the same as the current one then redirect.
 *
 * Since there could be some defaults defined, the new url even with no params could end up being the same as the current url.
 * So to avoid a redirect loop we need to not redirect in this case.
 */
export function shouldRedirectToViewWithParams(
  exploreName: string,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
  prefix: string | undefined,
  url: URL,
) {
  if (
    // exactly one param is set, but it is not `view`
    (url.searchParams.size === 1 &&
      !url.searchParams.has(ExploreStateURLParams.WebView)) ||
    // exactly 2 params are set and both `view` and `measure` are not set
    (url.searchParams.size === 2 &&
      !url.searchParams.has(ExploreStateURLParams.WebView) &&
      !url.searchParams.has(ExploreStateURLParams.ExpandedMeasure)) ||
    // more than 2 params are set
    url.searchParams.size > 2
  ) {
    return;
  }

  const viewFromUrl = url.searchParams.get(ExploreStateURLParams.WebView);
  const view = viewFromUrl
    ? FromURLParamViewMap[viewFromUrl]
    : (defaultExplorePreset.view ??
      V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED);
  const explorePresetFromSessionStorage = getExplorePresetForWebView(
    exploreName,
    prefix,
    view,
  );
  if (!explorePresetFromSessionStorage) {
    return;
  }

  const { partialExploreState } = convertPresetToExploreState(
    metricsViewSpec,
    exploreSpec,
    explorePresetFromSessionStorage,
  );
  const newUrl = new URL(url);
  newUrl.search = convertExploreStateToURLSearchParams(
    partialExploreState as MetricsExplorerEntity,
    exploreSpec,
    defaultExplorePreset,
  );
  // copy over any partial params. this will include the view and measure param
  url.searchParams.forEach((value, key) => newUrl.searchParams.set(key, value));
  if (newUrl.toString() === url.toString()) {
    // url hasn't changed, avoid redirect loop
    return;
  }

  return `${newUrl.pathname}${newUrl.search}`;
}
