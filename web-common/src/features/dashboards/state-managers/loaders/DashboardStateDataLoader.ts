import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { cascadingExploreStateMerge } from "@rilldata/web-common/features/dashboards/state-managers/cascading-explore-state-merge";
import { getPartialExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { getMostRecentPartialExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { normalizeWeekday } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import { cleanEmbedUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  getQueryServiceMetricsViewTimeRangeQueryOptions,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
} from "@rilldata/web-common/runtime-client";
import type { AfterNavigate } from "@sveltejs/kit";
import { createQuery, type QueryClient } from "@tanstack/svelte-query";
import { Settings } from "luxon";
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
  public readonly rillDefaultExploreState: CompoundQueryResult<ExploreState>;
  // Explore state from yaml config
  public readonly exploreStateFromYAMLConfig: CompoundQueryResult<
    Partial<ExploreState>
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
    ExploreState | undefined
  >;

  public constructor(
    instanceId: string,
    private readonly exploreName: string,
    private readonly storageNamespacePrefix: string | undefined,
    private readonly bookmarkOrTokenExploreState:
      | CompoundQueryResult<Partial<ExploreState> | null>
      | undefined,
    public readonly disableMostRecentDashboardState: boolean,
  ) {
    this.validSpecQuery = useExploreValidSpec(instanceId, exploreName);
    this.fullTimeRangeQuery = this.useFullTimeRangeQuery(
      instanceId,
      this.validSpecQuery,
    );

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
          metricsViewSpec.smallestTimeGrain,
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
   * Wrapper function that fetches full time range.
   * Uses useExploreValidSpec unlike the useMetricsViewTimeRange since it is more widely used.
   *
   * Does an additional validation where null min and max returned throws an error instead.
   */
  private useFullTimeRangeQuery(
    instanceId: string,
    validSpecQuery: ReturnType<typeof useExploreValidSpec>,
    queryClient?: QueryClient,
  ): CompoundQueryResult<V1MetricsViewTimeRangeResponse> {
    const fullTimeRangeQueryOptionsStore = derived(
      validSpecQuery,
      (validSpecResp) => {
        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};
        const firstDayOfWeek = metricsViewSpec.firstDayOfWeek;
        const metricsViewName = exploreSpec.metricsView ?? "";

        Settings.defaultWeekSettings = {
          firstDay: normalizeWeekday(firstDayOfWeek),
          weekend: [6, 7],
          minimalDays: 4,
        };

        return getQueryServiceMetricsViewTimeRangeQueryOptions(
          instanceId,
          metricsViewName,
          {},
          {
            query: {
              enabled: Boolean(metricsViewSpec.timeDimension),
            },
          },
        );
      },
    );
    const fullTimeRangeQuery = createQuery(
      fullTimeRangeQueryOptionsStore,
      queryClient,
    );

    return derived(
      [fullTimeRangeQueryOptionsStore, fullTimeRangeQuery],
      ([fullTimeRangeQueryOptions, fullTimeRange]) => {
        // TODO: update the fields once we move away from getCompoundQuery

        if (!fullTimeRangeQueryOptions.enabled) {
          // We return early to avoid having isLoading=true when the time range query is not enabled.
          // This allows us to check isLoading further down without any issues of it getting stuck.
          // TODO: revisit once we move away from getCompoundQuery
          return {
            data: undefined,
            error: null,
            isFetching: false,
            isLoading: false,
          };
        }

        if (
          fullTimeRange.data?.timeRangeSummary?.min === null &&
          fullTimeRange.data?.timeRangeSummary?.max === null
        ) {
          // The timeRangeSummary is null when there are 0 rows of data.
          // Notably, this happens when a security policy fully restricts a user from reading any data.
          // Show a different error in this case.
          return {
            data: undefined,
            error: new Error(
              "This dashboard currently has no data to display. This may be due to access permissions.",
            ),
            isFetching: false,
            isLoading: false,
          };
        }

        return {
          data: fullTimeRange.data,
          error: fullTimeRange.error,
          isFetching: fullTimeRange.isFetching,
          isLoading: fullTimeRange.isLoading,
        };
      },
    );
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
    bookmarkOrTokenExploreState: Partial<ExploreState> | null | undefined;
    exploreStateFromYAMLConfig: Partial<ExploreState>;
    rillDefaultExploreState: ExploreState;
    backButtonUsed: boolean;
  }) {
    urlSearchParams = cleanEmbedUrlParams(urlSearchParams);

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
      shouldSkipOtherSources || this.disableMostRecentDashboardState
        ? null
        : mostRecentPartialExploreState,
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
    ) as Partial<ExploreState>[];
    const finalExploreState = cascadingExploreStateMerge(
      nonEmptyExploreStateOrder,
    ) as ExploreState;

    return finalExploreState;
  }
}
