import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { useFullTimeRangeQuery } from "@rilldata/web-common/features/dashboards/selectors";
import { cascadingExploreStateMerge } from "@rilldata/web-common/features/dashboards/state-managers/cascading-explore-state-merge";
import { getPartialExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { getMostRecentPartialExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { AfterNavigate } from "@sveltejs/kit";
import { get } from "svelte/store";

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

    this.initExploreState = getCompoundQuery(
      [
        this.validSpecQuery,
        this.rillDefaultExploreState,
        this.exploreStateFromYAMLConfig,
        ...(bookmarkOrTokenExploreState ? [bookmarkOrTokenExploreState] : []),
      ],
      ([
        validSpecResp,
        rillDefaultExploreState,
        exploreStateFromYAMLConfig,
        bookmarkOrTokenExploreState,
      ]) => {
        const metricsViewSpec = validSpecResp?.metricsView;
        const exploreSpec = validSpecResp?.explore;
        if (
          !metricsViewSpec ||
          !exploreSpec ||
          !rillDefaultExploreState ||
          !exploreStateFromYAMLConfig
        ) {
          return undefined;
        }

        return this.getMergedExploreState({
          metricsViewSpec,
          exploreSpec,
          urlSearchParams: get(page).url.searchParams,
          bookmarkOrTokenExploreState,
          exploreStateFromYAMLConfig,
          rillDefaultExploreState,
          backButtonUsed: false,
        });
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

    return this.getMergedExploreState({
      metricsViewSpec,
      exploreSpec,
      urlSearchParams,
      bookmarkOrTokenExploreState: this.bookmarkOrTokenExploreState
        ? get(this.bookmarkOrTokenExploreState).data
        : null,
      exploreStateFromYAMLConfig,
      rillDefaultExploreState,
      backButtonUsed,
    });
  }

  /**
   * Decides the order of merging of various explore state source.
   * Returns a cascading merged state of the sources.
   */
  private getMergedExploreState({
    metricsViewSpec,
    exploreSpec,
    urlSearchParams,
    bookmarkOrTokenExploreState,
    exploreStateFromYAMLConfig,
    rillDefaultExploreState,
    backButtonUsed,
  }: {
    metricsViewSpec: V1MetricsViewSpec;
    exploreSpec: V1ExploreSpec;
    urlSearchParams: URLSearchParams;
    bookmarkOrTokenExploreState:
      | Partial<MetricsExplorerEntity>
      | null
      | undefined;
    exploreStateFromYAMLConfig: Partial<MetricsExplorerEntity>;
    rillDefaultExploreState: MetricsExplorerEntity;
    backButtonUsed: boolean;
  }) {
    const skipSessionStorage = backButtonUsed;
    const exploreStateFromSessionStorage = skipSessionStorage
      ? null
      : getPartialExploreStateFromSessionStorage(
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

    const { mostRecentPartialExploreState } = getMostRecentPartialExploreState(
      this.exploreName,
      this.storageNamespacePrefix,
      metricsViewSpec,
      exploreSpec,
    );

    const shouldSkipOtherSources =
      // If the url has some params that do not map to session storage then we need to only use state from url back-filled with rill defaults.
      (urlSearchParams.size > 0 && !exploreStateFromSessionStorage) ||
      // The exception to this is when back button is pressed and the user landed on empty url.
      backButtonUsed;

    const exploreStateOrder = [
      // 1st priority is the state from url params. For certain params the state is from session storage.
      // We need the state from session storage to make sure any state is not cleared while the user is still on the page but came back from a different dashboard.
      // TODO: move all this logic based on url params to a "fromURL" method. Will replace convertURLSearchParamsToExploreState
      exploreStateFromSessionStorage ??
        (urlSearchParams.size > 0 ? partialExploreStateFromUrl : null),
      // Next priority is the most recent state user had visited. This is a small subset of the full state.
      shouldSkipOtherSources ? null : mostRecentPartialExploreState,
      // Next priority is one of the other source defined.
      // For cloud dashboard it would be home bookmark if present.
      // For shared url it would be the saved state in token
      shouldSkipOtherSources ? null : bookmarkOrTokenExploreState,
      // Next priority is the defaults from yaml config.
      shouldSkipOtherSources ? null : exploreStateFromYAMLConfig,
      // Finally the fallback of rill default explore which will have the complete set of config.
      rillDefaultExploreState,
    ];

    const nonEmptyExploreStateOrder = exploreStateOrder.filter(
      Boolean,
    ) as Partial<MetricsExplorerEntity>[];
    const finalExploreState = cascadingExploreStateMerge(
      nonEmptyExploreStateOrder,
    ) as MetricsExplorerEntity;

    return finalExploreState;
  }
}
