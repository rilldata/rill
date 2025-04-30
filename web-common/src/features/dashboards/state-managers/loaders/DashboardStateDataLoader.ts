import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { useFullTimeRangeQuery } from "@rilldata/web-common/features/dashboards/selectors";
import { cascadingExploreStateMerge } from "@rilldata/web-common/features/dashboards/state-managers/cascading-explore-state-merge";
import { getPartialExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { type V1MetricsViewTimeRangeResponse } from "@rilldata/web-common/runtime-client";
import type { AfterNavigate } from "@sveltejs/kit";
import { derived, get } from "svelte/store";

/**
 * Loads data from explore and metrics view specs, along with all time range query.
 * Mainly outputs an initial explore state based on various conditions. Check initExploreState CompoundQuery for more info.
 * Also has a method to get a partial explore state based on url params.
 */
export class DashboardStateDataLoader {
  // These can be used to show a loading status
  public readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  public readonly fullTimeRangeQuery: CompoundQueryResult<V1MetricsViewTimeRangeResponse>;

  // Default explore state show when there is no data in session/local storage or a home bookmark.
  public readonly rillDefaultExploreState: CompoundQueryResult<MetricsExplorerEntity>;
  // Explore state from yaml config
  public readonly exploreStateFromYAMLConfig: CompoundQueryResult<
    Partial<MetricsExplorerEntity>
  >;

  private readonly partialExploreStateFromUrlForInit: CompoundQueryResult<
    Partial<MetricsExplorerEntity> | undefined
  >;

  /**
   * The explore state used to populate the store with initial explore.
   * This is a cascading merge of various states in order,
   * 1. Session storage if url params doesn't have params other than `view` and `measure` for TDD
   * 2. Params directly from the url. If sessions storage is not present then the rill defaults are merged into this for empty params.
   * 3. Bookmark or token state if provided.
   * 4. Dashboard config from yaml.
   * 5. Rill opinionated defaults.
   */
  public readonly initExploreState: CompoundQueryResult<
    MetricsExplorerEntity | undefined
  >;

  public constructor(
    instanceId: string,
    private readonly exploreName: string,
    private readonly storageNamespacePrefix: string | undefined,
    private readonly bookmarkOrTokenExploreState?: CompoundQueryResult<Partial<MetricsExplorerEntity> | null>,
  ) {
    this.validSpecQuery = useExploreValidSpec(instanceId, exploreName);
    this.fullTimeRangeQuery = useFullTimeRangeQuery(instanceId, exploreName);

    this.rillDefaultExploreState = getCompoundQuery(
      [this.validSpecQuery, this.fullTimeRangeQuery],
      ([validSpecResp, metricsViewTimeRangeResp]) => {
        const metricsViewSpec = validSpecResp?.metricsView;
        const exploreSpec = validSpecResp?.explore;

        if (
          !metricsViewSpec ||
          !exploreSpec ||
          // safeguard to make sure time range summary is loaded for metrics view with time dimension
          (metricsViewSpec.timeDimension &&
            !metricsViewTimeRangeResp?.timeRangeSummary)
        ) {
          return undefined;
        }

        return getRillDefaultExploreState(
          metricsViewSpec,
          exploreSpec,
          metricsViewTimeRangeResp?.timeRangeSummary,
        );
      },
    );

    this.exploreStateFromYAMLConfig = getCompoundQuery(
      [this.validSpecQuery, this.fullTimeRangeQuery],
      ([validSpecResp, metricsViewTimeRangeResp]) => {
        const metricsViewSpec = validSpecResp?.metricsView;
        const exploreSpec = validSpecResp?.explore;

        if (
          !metricsViewSpec ||
          !exploreSpec ||
          // safeguard to make sure time range summary is loaded for metrics view with time dimension
          (metricsViewSpec.timeDimension &&
            !metricsViewTimeRangeResp?.timeRangeSummary)
        ) {
          return undefined;
        }

        return getExploreStateFromYAMLConfig(
          exploreSpec,
          metricsViewTimeRangeResp?.timeRangeSummary,
        );
      },
    );

    this.partialExploreStateFromUrlForInit = derived(
      [this.validSpecQuery, page],
      ([validSpecResp, pageState]) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};

        let partialExploreStateFromUrlForInit:
          | Partial<MetricsExplorerEntity>
          | undefined = undefined;
        const haveSomeUrlParams = pageState.url.searchParams.size > 0;
        // Only do the conversion if there are some url params.
        // This way with an blank url state from url will be undefined,
        if (haveSomeUrlParams) {
          ({ partialExploreState: partialExploreStateFromUrlForInit } =
            convertURLSearchParamsToExploreState(
              pageState.url.searchParams,
              metricsViewSpec,
              exploreSpec,
              {},
            ));
        }

        return {
          data: partialExploreStateFromUrlForInit,
          error: validSpecResp.error,
          isLoading: validSpecResp.isLoading,
          isFetching: validSpecResp.isFetching,
        };
      },
    );

    this.initExploreState = getCompoundQuery(
      [
        this.rillDefaultExploreState,
        this.exploreStateFromYAMLConfig,
        this.partialExploreStateFromUrlForInit,
        ...(bookmarkOrTokenExploreState ? [bookmarkOrTokenExploreState] : []),
      ],
      ([
        rillDefaultExploreState,
        exploreStateFromYAMLConfig,
        partialExploreStateFromUrlForInit,
        bookmarkOrTokenExploreState,
      ]) => {
        if (!rillDefaultExploreState || !exploreStateFromYAMLConfig) {
          return undefined;
        }

        let exploreStateOrder: (
          | Partial<MetricsExplorerEntity>
          | null
          | undefined
        )[];
        if (partialExploreStateFromUrlForInit) {
          // If there are some url params then we need to fill in any missing params from rill defaults. No other state will be used.
          exploreStateOrder = [
            // 1st priority is the state loaded from url params. It will be undefined if there are no params.
            partialExploreStateFromUrlForInit,
            // Finally the fallback of rill default explore which will have the complete set of config.
            rillDefaultExploreState,
          ];
        } else {
          // Else merge other states like bookmark/token and state from yaml config
          exploreStateOrder = [
            // 1st priority is one of the other source defined.
            // For cloud dashboard it would be home bookmark if present.
            // For shared url it would be the saved state in token
            bookmarkOrTokenExploreState,
            // Next priority is the defaults from yaml config.
            exploreStateFromYAMLConfig,
            // Finally the fallback of rill default explore which will have the complete set of config.
            rillDefaultExploreState,
          ];
        }

        const nonEmptyExploreStateOrder = exploreStateOrder.filter(
          Boolean,
        ) as Partial<MetricsExplorerEntity>[];
        const initExploreState = cascadingExploreStateMerge(
          nonEmptyExploreStateOrder,
        );

        return initExploreState as MetricsExplorerEntity;
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
    const { data: rillDefaultExploreState } = get(this.rillDefaultExploreState);
    const { data: exploreStateFromYAMLConfig } = get(
      this.exploreStateFromYAMLConfig,
    );

    if (!rillDefaultExploreState || !exploreStateFromYAMLConfig) {
      return undefined;
    }

    // Pressing back button and going back to empty url state should not restore from session store
    const backButtonUsed = type === "popstate";
    const skipSessionStorage = backButtonUsed;

    const exploreStateFromSessionStorage =
      getPartialExploreStateFromSessionStorage(
        this.exploreName,
        this.storageNamespacePrefix,
        urlSearchParams,
        metricsViewSpec,
        exploreSpec,
      );

    const { partialExploreState: partialExploreStateFromUrl } =
      convertURLSearchParamsToExploreState(
        urlSearchParams,
        metricsViewSpec,
        exploreSpec,
        {},
      );

    let exploreStateOrder: (
      | Partial<MetricsExplorerEntity>
      | null
      | undefined
    )[];
    if (urlSearchParams.size > 0 || skipSessionStorage) {
      // If there are some url params then we need to fill in any missing params from rill defaults.
      exploreStateOrder = [
        // 1st priority is the state from session storage.
        skipSessionStorage ? undefined : exploreStateFromSessionStorage,
        // Next priority is the state loaded from url params. It will be undefined if there are no params.
        partialExploreStateFromUrl,
        // If there are some url params then the next state will be rill default explore state
        rillDefaultExploreState,
      ];
    } else {
      // Else merge other states like bookmark/token and state from yaml config
      // We need this explicit adding of states to reset the store when going back to an empty url.
      exploreStateOrder = [
        // 1st priority is the state from session storage.
        // We need this to make sure any state is not cleared while the user is still on the page but came back from a different dashboard.
        skipSessionStorage ? undefined : exploreStateFromSessionStorage,
        // Next priority is one of the other source defined.
        // For cloud dashboard it would be home bookmark if present.
        // For shared url it would be the saved state in token
        this.bookmarkOrTokenExploreState
          ? get(this.bookmarkOrTokenExploreState).data
          : undefined,
        // Next priority is the defaults from yaml config.
        exploreStateFromYAMLConfig,
        // Finally the fallback of rill default explore which will have the complete set of config.
        rillDefaultExploreState,
      ];
    }

    const nonEmptyExploreStateOrder = exploreStateOrder.filter(
      Boolean,
    ) as Partial<MetricsExplorerEntity>[];
    return cascadingExploreStateMerge(nonEmptyExploreStateOrder);
  }
}
