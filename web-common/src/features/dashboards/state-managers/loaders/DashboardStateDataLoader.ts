import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getPartialExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  createTimeControlStoreCompoundQuery,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import {
  ExploreWebViewSpecificURLParams,
  GlobalExploreURLParams,
} from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createQueryServiceMetricsViewTimeRange,
  type V1ExplorePreset,
} from "@rilldata/web-common/runtime-client";
import type { AfterNavigate } from "@sveltejs/kit";
import { derived, get } from "svelte/store";

/**
 * Loads data from explore and metrics view specs, along with all time range query.
 * Mainly outputs a initial explore state based on various conditions. Check initExploreState CompoundQuery for more info.
 * Also has a method to get a partial explore state based on url params.
 */
export class DashboardStateDataLoader {
  // These can be used to show a loading status
  public readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  public readonly fullTimeRangeQuery: ReturnType<
    typeof useMetricsViewTimeRange
  >;
  private readonly timeControlStore: CompoundQueryResult<TimeControlState>;

  // This is used to decide defaults and show/hide url params. TODO: is this the correct preset?
  public readonly explorePresetFromYAMLConfig: CompoundQueryResult<
    V1ExplorePreset | undefined
  >;

  private readonly blankDashboardUrlParams: CompoundQueryResult<URLSearchParams>;
  private readonly sessionStorageUrlParams: CompoundQueryResult<
    URLSearchParams | undefined
  >;
  private readonly yamlConfigUrlParams: CompoundQueryResult<URLSearchParams>;

  public readonly initUrlParams: CompoundQueryResult<URLSearchParams>;

  public constructor(
    instanceId: string,
    metricsViewName: string,
    private readonly exploreName: string,
    private readonly storageNamespacePrefix: string | undefined,
    private readonly bookmarkOrTokenUrlParams?: CompoundQueryResult<URLSearchParams | null>,
  ) {
    this.validSpecQuery = useExploreValidSpec(instanceId, exploreName);
    this.fullTimeRangeQuery = derived(
      [this.validSpecQuery],
      ([validSpecResp], set) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        if (!metricsViewSpec.timeDimension) {
          // We return early to avoid having isLoading=true when time dimension is not present.
          // This allows us to check isLoading further down without any issues of it getting stuck.
          set({
            data: undefined,
            error: null,
            isLoading: false,
            isFetching: false,
          } as any);
          return;
        }

        createQueryServiceMetricsViewTimeRange(
          instanceId,
          metricsViewName,
          {},
          {
            query: {
              queryClient,
            },
          },
        ).subscribe(set);
      },
    );

    this.timeControlStore = createTimeControlStoreCompoundQuery(
      instanceId,
      metricsViewName,
      exploreName,
    );

    this.blankDashboardUrlParams = getCompoundQuery(
      [this.validSpecQuery, this.fullTimeRangeQuery, this.timeControlStore],
      ([validSpecResp, metricsViewTimeRangeResp, timeControlState]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};

        // safeguard to make sure time range summary is loaded for metrics view with time dimension
        if (
          metricsViewSpec.timeDimension &&
          (!metricsViewTimeRangeResp?.timeRangeSummary ||
            !timeControlState?.ready)
        ) {
          return undefined;
        }

        // TODO: get rid of this step by step conversion
        const defaultExplorePreset = getDefaultExplorePreset(
          {
            ...exploreSpec,
            defaultPreset: {},
          },
          metricsViewSpec,
          metricsViewTimeRangeResp,
        );
        const { partialExploreState: defaultExploreState } =
          convertPresetToExploreState(
            metricsViewSpec,
            exploreSpec,
            defaultExplorePreset,
          );
        return convertExploreStateToURLSearchParams(
          defaultExploreState as MetricsExplorerEntity,
          exploreSpec,
          timeControlState,
          {},
        );
      },
    );

    this.sessionStorageUrlParams = derived(
      [this.validSpecQuery, page],
      ([validSpecResp, pageState]) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};
        const sessionStorageUrlParams =
          getPartialExploreStateFromSessionStorage(
            exploreName,
            storageNamespacePrefix,
            pageState.url.searchParams,
            metricsViewSpec,
            exploreSpec,
          );

        return {
          data: sessionStorageUrlParams,
          error: validSpecResp.error,
          isLoading: validSpecResp.isLoading,
          isFetching: validSpecResp.isFetching,
        };
      },
    );

    this.yamlConfigUrlParams = getCompoundQuery(
      [
        this.validSpecQuery,
        this.explorePresetFromYAMLConfig,
        this.timeControlStore,
      ],
      ([validSpecResp, explorePresetFromYAMLConfig, timeControlState]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};
        const { partialExploreState: exploreStateFromYAMLConfig } =
          convertPresetToExploreState(
            metricsViewSpec,
            exploreSpec,
            explorePresetFromYAMLConfig ?? {},
          );
        return convertExploreStateToURLSearchParams(
          exploreStateFromYAMLConfig as MetricsExplorerEntity,
          exploreSpec,
          timeControlState,
          {},
        );
      },
    );

    this.initUrlParams = getCompoundQuery(
      [
        // TODO: find a better way
        derived(page, (pageState) => ({
          data: pageState,
          error: undefined,
          isLoading: false,
          isFetching: false,
        })),
        this.sessionStorageUrlParams,
        this.yamlConfigUrlParams,
        this.blankDashboardUrlParams,
      ],
      ([
        pageState,
        sessionStorageUrlParams,
        yamlConfigUrlParams,
        blankDashboardUrlParams,
      ]) => {
        // guards against data not being loaded
        if (!blankDashboardUrlParams || !yamlConfigUrlParams) {
          return undefined;
        }

        return this.getCascadingUrlParams(
          pageState!.url.searchParams,
          sessionStorageUrlParams,
          yamlConfigUrlParams,
          blankDashboardUrlParams,
        );
      },
    );
  }

  // The decision to get the exploreState from url params depends on the navigation type.
  // So we cannot go the derived store route.
  public getExploreStateFromURLParams(
    urlSearchParams: URLSearchParams,
    type: AfterNavigate["type"],
  ) {
    const validSpecResp = get(this.validSpecQuery);
    if (!validSpecResp?.data?.metricsView || !validSpecResp?.data?.explore)
      return undefined;
    const metricsViewSpec = validSpecResp.data.metricsView;
    const exploreSpec = validSpecResp.data.explore;

    const yamlConfigUrlParams = get(this.yamlConfigUrlParams);
    const blankDashboardUrlParams = get(this.blankDashboardUrlParams);
    if (!yamlConfigUrlParams.data || !blankDashboardUrlParams.data)
      return undefined;

    // Pressing back button and going back to empty url state should not restore from session store
    const backButtonUsed = type === "popstate";
    const skipSessionStorage = backButtonUsed;

    // If we are skipping using state from session storage then exit early with partialExploreStateFromUrl
    // regardless if there is exploreStateFromSessionStorage for current url params or not.
    if (skipSessionStorage) {
      return this.getCascadingUrlParams(
        urlSearchParams,
        undefined,
        yamlConfigUrlParams.data,
        blankDashboardUrlParams.data,
      );
    }

    const sessionStorageUrlParams = getPartialExploreStateFromSessionStorage(
      this.exploreName,
      this.storageNamespacePrefix,
      urlSearchParams,
      metricsViewSpec,
      exploreSpec,
    );

    return this.getCascadingUrlParams(
      urlSearchParams,
      sessionStorageUrlParams,
      yamlConfigUrlParams.data,
      blankDashboardUrlParams.data,
    );
  }

  private getCascadingUrlParams(
    pageUrlParams: URLSearchParams,
    sessionStorageUrlParams: URLSearchParams | undefined,
    yamlConfigUrlParams: URLSearchParams,
    blankDashboardUrlParams: URLSearchParams,
  ) {
    const initUrlParams = new URLSearchParams();

    const urlParamsOrder = [
      pageUrlParams,
      sessionStorageUrlParams,
      this.bookmarkOrTokenUrlParams,
      yamlConfigUrlParams,
      blankDashboardUrlParams,
    ].filter(Boolean) as URLSearchParams[];

    const curView =
      pageUrlParams.get(ExploreStateURLParams.WebView) ??
      ExploreUrlWebView.Explore;
    [
      ...GlobalExploreURLParams,
      ...ExploreWebViewSpecificURLParams[curView],
    ].forEach((k) => {
      const val = urlParamsOrder.find((p) => p.has(k));
      if (!val) return;
      initUrlParams.set(k, val.get(k) ?? "");
    });

    return initUrlParams;
  }
}
