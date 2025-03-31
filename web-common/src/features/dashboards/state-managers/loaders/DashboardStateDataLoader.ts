import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/get-explore-state-from-session-storage";
import { getMostRecentExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertPresetToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
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

  private readonly defaultExploreStateAndErrors: CompoundQueryResult<{
    defaultExploreState: Partial<MetricsExplorerEntity> | undefined;
    errors: Error[];
  }>;
  private readonly exploreStateFromYAMLConfigAndErrors: CompoundQueryResult<{
    exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity>;
    errors: Error[];
  }>;
  // This is used to decide defaults and show/hide url params. TODO: is this the correct preset?
  public readonly explorePresetFromYAMLConfig: CompoundQueryResult<
    V1ExplorePreset | undefined
  >;
  private readonly mostRecentPartialExploreStateAndErrors: CompoundQueryResult<{
    mostRecentPartialExploreState: Partial<MetricsExplorerEntity> | undefined;
    errors: Error[];
  }>;

  private readonly exploreStateFromSessionStorage: CompoundQueryResult<
    Partial<MetricsExplorerEntity> | undefined
  >;
  private readonly partialExploreStateFromUrlForInitAndErrors: CompoundQueryResult<{
    partialExploreStateFromUrlForInit:
      | Partial<MetricsExplorerEntity>
      | undefined;
    errors: Error[];
  }>;

  /**
   * The explore state used to populate the store with initial explore.
   * 1. If state is present in the url, use it.
   * 2. If no url state, load from session storage (only persists within the tab)
   * 3. If no url state, session storage, restore user's most recent state (from local storage).
   * 4. If no url state, session storage, most recent state, apply home bookmark (cloud only).
   * 5. If no url state, session storage, most recent state, home bookmark, apply explore.yaml defaults
   * 6. If no url state, session storage, most recent state, home bookmark or defaults open as blank dashboard.
   */
  public readonly initExploreState: CompoundQueryResult<
    MetricsExplorerEntity | undefined
  >;

  public constructor(
    instanceId: string,
    metricsViewName: string,
    private readonly exploreName: string,
    private readonly storageNamespacePrefix: string | undefined,
    bookmarkOrTokenExploreState?: CompoundQueryResult<
      Partial<MetricsExplorerEntity> | undefined
    >,
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
            isError: false,
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

    this.defaultExploreStateAndErrors = getCompoundQuery(
      [this.validSpecQuery, this.fullTimeRangeQuery],
      ([validSpecResp, metricsViewTimeRangeResp]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};

        // safeguard to make sure time range summary is loaded for metrics view with time dimension
        if (
          metricsViewSpec.timeDimension &&
          !metricsViewTimeRangeResp?.timeRangeSummary
        ) {
          return {
            defaultExploreState: undefined,
            errors: [],
          };
        }

        const defaultExplorePreset = getDefaultExplorePreset(
          {
            ...exploreSpec,
            defaultPreset: {},
          },
          metricsViewSpec,
          metricsViewTimeRangeResp,
        );
        const { partialExploreState: defaultExploreState, errors } =
          convertPresetToExploreState(
            metricsViewSpec,
            exploreSpec,
            defaultExplorePreset,
          );
        return {
          defaultExploreState,
          errors,
        };
      },
    );

    this.explorePresetFromYAMLConfig = getCompoundQuery(
      [this.validSpecQuery, this.fullTimeRangeQuery],
      ([validSpecResp, metricsViewTimeRangeResp]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};

        // safeguard to make sure time range summary is loaded for metrics view with time dimension
        if (
          metricsViewSpec.timeDimension &&
          !metricsViewTimeRangeResp?.timeRangeSummary
        ) {
          return undefined;
        }

        return getDefaultExplorePreset(
          exploreSpec,
          metricsViewSpec,
          metricsViewTimeRangeResp,
        );
      },
    );

    this.exploreStateFromYAMLConfigAndErrors = getCompoundQuery(
      [this.validSpecQuery, this.explorePresetFromYAMLConfig],
      ([validSpecResp, explorePresetFromYAMLConfig]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};
        const { partialExploreState: exploreStateFromYAMLConfig, errors } =
          convertPresetToExploreState(
            metricsViewSpec,
            exploreSpec,
            explorePresetFromYAMLConfig ?? {},
          );
        return {
          exploreStateFromYAMLConfig,
          errors,
        };
      },
    );

    this.mostRecentPartialExploreStateAndErrors = getCompoundQuery(
      [this.validSpecQuery],
      ([validSpecResp]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};
        const { partialExploreState: mostRecentPartialExploreState, errors } =
          getMostRecentExploreState(
            exploreName,
            storageNamespacePrefix,
            metricsViewSpec,
            exploreSpec,
          );
        return {
          mostRecentPartialExploreState,
          errors,
        };
      },
    );

    this.exploreStateFromSessionStorage = derived(
      [this.validSpecQuery, this.explorePresetFromYAMLConfig, page],
      ([validSpecResp, explorePresetFromYAMLConfig, pageState]) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};
        const exploreStateFromSessionStorage =
          getExploreStateFromSessionStorage(
            exploreName,
            storageNamespacePrefix,
            pageState.url.searchParams,
            metricsViewSpec,
            exploreSpec,
            explorePresetFromYAMLConfig.data ?? {},
          );

        return {
          data: exploreStateFromSessionStorage,
          error: validSpecResp.error ?? explorePresetFromYAMLConfig.error,
          isLoading:
            validSpecResp.isLoading || explorePresetFromYAMLConfig.isLoading,
          isFetching:
            validSpecResp.isFetching || explorePresetFromYAMLConfig.isFetching,
        };
      },
    );

    this.partialExploreStateFromUrlForInitAndErrors = derived(
      [this.validSpecQuery, this.explorePresetFromYAMLConfig, page],
      ([validSpecResp, explorePresetFromYAMLConfig, pageState]) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};

        const { partialExploreState: partialExploreStateFromUrl, errors } =
          convertURLSearchParamsToExploreState(
            pageState.url.searchParams,
            metricsViewSpec,
            exploreSpec,
            explorePresetFromYAMLConfig.data ?? {},
          );
        const partialExploreStateFromUrlForInit =
          pageState.url.searchParams.size === 0
            ? undefined
            : partialExploreStateFromUrl;

        return {
          data: {
            partialExploreStateFromUrlForInit,
            errors,
          },
          error: validSpecResp.error ?? explorePresetFromYAMLConfig.error,
          isLoading:
            validSpecResp.isLoading || explorePresetFromYAMLConfig.isLoading,
          isFetching:
            validSpecResp.isFetching || explorePresetFromYAMLConfig.isFetching,
        };
      },
    );

    this.initExploreState = getCompoundQuery(
      [
        this.defaultExploreStateAndErrors,
        this.exploreStateFromSessionStorage,
        this.partialExploreStateFromUrlForInitAndErrors,
        this.mostRecentPartialExploreStateAndErrors,
        this.exploreStateFromYAMLConfigAndErrors,
        ...(bookmarkOrTokenExploreState ? [bookmarkOrTokenExploreState] : []),
      ],
      ([
        defaultExploreStateAndErrors,
        exploreStateFromSessionStorage,
        partialExploreStateFromUrlForInitAndErrors,
        mostRecentPartialExploreStateAndErrors,
        exploreStateFromYAMLConfigAndErrors,
        bookmarkOrTokenExploreState,
      ]) => {
        // type guards. other fields dont need it since we have chaining `??`
        if (
          !defaultExploreStateAndErrors?.defaultExploreState ||
          !exploreStateFromYAMLConfigAndErrors?.exploreStateFromYAMLConfig
        ) {
          return undefined;
        }

        const initExploreState = {
          // Since this is a complete state, we need the complete default explore state which works as a base.
          ...defaultExploreStateAndErrors.defaultExploreState,
          // 1st priority is the state from session storage.
          // TODO: since this only loads on certain params present in the url it should be merged with convertURLSearchParamsToExploreState
          ...(exploreStateFromSessionStorage ??
            // Next priority is the state loaded from url params. It will be undefined if there are no params.
            partialExploreStateFromUrlForInitAndErrors?.partialExploreStateFromUrlForInit ??
            // Next priority is the most recent state stored in local storage
            mostRecentPartialExploreStateAndErrors?.mostRecentPartialExploreState ??
            // Next priority is one of the other source defined.
            // For cloud dashboard it would be home bookmark if present.
            // For shared url it would be the saved state in token
            bookmarkOrTokenExploreState ??
            // Finally the state from yaml is used
            exploreStateFromYAMLConfigAndErrors.exploreStateFromYAMLConfig),
        } as MetricsExplorerEntity;

        return initExploreState;
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
    const explorePresetFromYAMLConfig = get(this.explorePresetFromYAMLConfig);
    if (!explorePresetFromYAMLConfig.data) return undefined;

    // Pressing back button and going back to empty url state should not restore from session store
    const backButtonUsed = type === "popstate";
    const skipSessionStorage = backButtonUsed;

    const { partialExploreState: partialExploreStateFromUrl } =
      convertURLSearchParamsToExploreState(
        urlSearchParams,
        metricsViewSpec,
        exploreSpec,
        explorePresetFromYAMLConfig.data,
      );
    // If we are skipping using state from session storage then exit early with partialExploreStateFromUrl
    // regardless if there is exploreStateFromSessionStorage for current url params or not.
    if (skipSessionStorage) return partialExploreStateFromUrl;

    const exploreStateFromSessionStorage = getExploreStateFromSessionStorage(
      this.exploreName,
      this.storageNamespacePrefix,
      urlSearchParams,
      metricsViewSpec,
      exploreSpec,
      explorePresetFromYAMLConfig.data,
    );

    return (
      // preference goes to session storage 1st
      exploreStateFromSessionStorage ??
      // else we use the partial explore state from the url params.
      partialExploreStateFromUrl
    );
  }
}
