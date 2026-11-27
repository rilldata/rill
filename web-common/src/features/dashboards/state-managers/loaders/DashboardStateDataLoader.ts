import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { loadEphemeralMeasureLibrary } from "@rilldata/web-common/features/dashboards/ephemeral-measures/library";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
import { getExploreUrlParamsFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { getMostRecentExploreUrlParams } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getExploreUrlParamsFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { validateAndCleanExploreState } from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import { normalizeWeekday } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import { cleanEmbedUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import {
  addEphemeralMeasureLibraryToUrlParams,
  mergeExploreUrlParams,
  normalizeExploreUrlParams,
} from "@rilldata/web-common/features/dashboards/url-state/merge-explore-url-params";
import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  getQueryServiceMetricsViewTimeRangeQueryOptions,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { AfterNavigate } from "@sveltejs/kit";
import { createQuery, type QueryClient } from "@tanstack/svelte-query";
import { Settings } from "luxon";
import { derived, get } from "svelte/store";

/**
 * Loads data from explore and metrics view specs, along with all time range query.
 * Mainly outputs the initial explore url params based on various conditions. Check initExploreUrlParams for more info.
 * Also has a method to get the explore url params for a url.
 * The params are converted to an explore state by the consumers.
 */
export class DashboardStateDataLoader {
  // These can be used to show a loading status
  public readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  public readonly fullTimeRangeQuery: CompoundQueryResult<V1MetricsViewTimeRangeResponse>;

  // Default explore state show when there is no data in session/local storage or a home bookmark.
  // Used to back-fill fields that have no param on the current view.
  public readonly rillDefaultExploreState: CompoundQueryResult<ExploreState>;
  // Url params for the rill default explore state. Used to fill in params and to remove default params from the url.
  public readonly rillDefaultExploreUrlParams: CompoundQueryResult<URLSearchParams>;
  // Url params for the fields set in the yaml config.
  public readonly exploreUrlParamsFromYAMLConfig: CompoundQueryResult<URLSearchParams>;

  /**
   * The url params used to populate the store with the initial explore.
   * One source provides the params and the others fill in what it leaves out, see mergeExploreUrlParams:
   * 1. Session storage if url params doesn't have params other than `view` and `measure` for TDD
   * 2. Params directly from the url, filled with rill defaults.
   * 3. Otherwise the home bookmark with the most recent params over it.
   * The yaml config and rill opinionated defaults fill in the rest.
   */
  public readonly initExploreUrlParams: CompoundQueryResult<
    URLSearchParams | undefined
  >;

  public constructor(
    private readonly client: RuntimeClient,
    private readonly exploreName: string,
    private readonly storageNamespacePrefix: string | undefined,
    private readonly bookmarkUrlParams:
      | CompoundQueryResult<URLSearchParams | null>
      | undefined,
    public readonly disableMostRecentDashboardState: boolean,
    public readonly disableInitSessionDashboardState: boolean,
  ) {
    this.validSpecQuery = useExploreValidSpec(client, exploreName);
    this.fullTimeRangeQuery = this.useFullTimeRangeQuery(
      client,
      this.validSpecQuery,
    );

    this.rillDefaultExploreState = this.getFromSpecs(
      getRillDefaultExploreState,
    );
    this.rillDefaultExploreUrlParams = this.getFromSpecs(
      getRillDefaultExploreUrlParams,
    );
    this.exploreUrlParamsFromYAMLConfig = this.getFromSpecs(
      getExploreUrlParamsFromYAMLConfig,
    );

    this.initExploreUrlParams = getCompoundQuery(
      [
        this.validSpecQuery,
        this.rillDefaultExploreUrlParams,
        this.exploreUrlParamsFromYAMLConfig,
        ...(bookmarkUrlParams ? [bookmarkUrlParams] : []),
      ],
      ([
        validSpecResp,
        rillDefaultExploreUrlParams,
        exploreUrlParamsFromYAMLConfig,
        bookmarkUrlParams,
      ]) => {
        const metricsViewSpec = validSpecResp?.metricsView;
        const exploreSpec = validSpecResp?.explore;
        if (
          !metricsViewSpec ||
          !exploreSpec ||
          !rillDefaultExploreUrlParams ||
          !exploreUrlParamsFromYAMLConfig
        ) {
          return undefined;
        }

        return this.getMergedExploreUrlParams({
          metricsViewSpec,
          exploreSpec,
          urlSearchParams: get(page).url.searchParams,
          bookmarkUrlParams,
          exploreUrlParamsFromYAMLConfig,
          rillDefaultExploreUrlParams,
          backButtonUsed: false,
          skipSessionStorage: this.disableInitSessionDashboardState,
        });
      },
    );
  }

  // The decision to get the params from the url depends on the navigation type.
  // So we cannot go the derived store route.
  public getExploreUrlParamsFromURL(
    urlSearchParams: URLSearchParams,
    type: AfterNavigate["type"],
  ) {
    const validSpecResp = get(this.validSpecQuery);
    if (!validSpecResp?.data?.metricsView || !validSpecResp?.data?.explore)
      return undefined;
    const metricsViewSpec = validSpecResp.data.metricsView;
    const exploreSpec = validSpecResp.data.explore;
    const { data: rillDefaultExploreUrlParams } = get(
      this.rillDefaultExploreUrlParams,
    );
    const { data: exploreUrlParamsFromYAMLConfig } = get(
      this.exploreUrlParamsFromYAMLConfig,
    );

    if (!rillDefaultExploreUrlParams || !exploreUrlParamsFromYAMLConfig) {
      return undefined;
    }

    // Pressing back button and going back to empty url state should not restore from session store
    const backButtonUsed = type === "popstate";

    return this.getMergedExploreUrlParams({
      metricsViewSpec,
      exploreSpec,
      urlSearchParams,
      bookmarkUrlParams: this.bookmarkUrlParams
        ? get(this.bookmarkUrlParams).data
        : null,
      exploreUrlParamsFromYAMLConfig,
      rillDefaultExploreUrlParams,
      backButtonUsed,
      // This should not be disabled when disableInitSessionDashboardState is set.
      // While going between views, we still need session storage to save the state.
      skipSessionStorage: false,
    });
  }

  /**
   * Derives a value from the specs and the full time range,
   * once the time range summary is loaded for a metrics view with a time dimension.
   */
  private getFromSpecs<T>(
    getter: (
      metricsViewSpec: V1MetricsViewSpec,
      exploreSpec: V1ExploreSpec,
      timeRangeSummary: V1TimeRangeSummary | undefined,
    ) => T,
  ): CompoundQueryResult<T> {
    return getCompoundQuery(
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

        return getter(
          metricsViewSpec,
          exploreSpec,
          metricsViewTimeRangeResp?.timeRangeSummary,
        );
      },
    );
  }

  /**
   * Wrapper function that fetches full time range.
   * Uses useExploreValidSpec unlike the useMetricsViewTimeRange since it is more widely used.
   *
   * Does an additional validation where null min and max returned throws an error instead.
   */
  private useFullTimeRangeQuery(
    client: RuntimeClient,
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
          client,
          { metricsViewName },
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
          !fullTimeRange.isLoading &&
          !fullTimeRange.isFetching &&
          fullTimeRange.data?.timeRangeSummary?.min == null &&
          fullTimeRange.data?.timeRangeSummary?.max == null
        ) {
          // The timeRangeSummary is null when there are 0 rows of data or all timestamp values are NULL.
          // In Cloud (or Developer with a mock user), access policies may be restricting data.
          // In Developer without a mock user, the issue is more likely data configuration.
          const isCloudOrMockUser =
            get(featureFlags.adminServer) ||
            get(selectedMockUserStore) !== null;
          const message = isCloudOrMockUser
            ? "This dashboard currently has no data to display. This may be due to the data source configuration or access permissions."
            : "This dashboard currently has no data to display. Check that your data source has rows and the time dimension column contains non-NULL values.";
          return {
            data: undefined,
            error: new Error(message),
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
   * Gathers the sources of explore url params and merges them, see mergeExploreUrlParams.
   */
  private getMergedExploreUrlParams({
    metricsViewSpec,
    exploreSpec,
    urlSearchParams,
    bookmarkUrlParams,
    exploreUrlParamsFromYAMLConfig,
    rillDefaultExploreUrlParams,
    backButtonUsed,
    skipSessionStorage,
  }: {
    metricsViewSpec: V1MetricsViewSpec;
    exploreSpec: V1ExploreSpec;
    urlSearchParams: URLSearchParams;
    bookmarkUrlParams: URLSearchParams | null | undefined;
    exploreUrlParamsFromYAMLConfig: URLSearchParams;
    rillDefaultExploreUrlParams: URLSearchParams;
    backButtonUsed: boolean;
    skipSessionStorage: boolean;
  }) {
    urlSearchParams = cleanEmbedUrlParams(urlSearchParams);

    const sessionParams =
      skipSessionStorage || backButtonUsed
        ? undefined
        : getExploreUrlParamsFromSessionStorage(
            this.exploreName,
            this.storageNamespacePrefix,
            urlSearchParams,
          );

    // Loaded up front: the most recent params' ad-hoc names are validated
    // against these definitions, and they are added to the merged params below.
    const libraryEphemeralMeasures = this.loadEphemeralMeasureLibrary(
      metricsViewSpec,
      exploreSpec,
    );

    const mostRecentParams = this.disableMostRecentDashboardState
      ? undefined
      : getMostRecentExploreUrlParams(
          this.exploreName,
          this.storageNamespacePrefix,
          metricsViewSpec,
          exploreSpec,
          libraryEphemeralMeasures,
        );

    const urlParams = mergeExploreUrlParams({
      urlParams: normalizeExploreUrlParams(urlSearchParams),
      sessionParams,
      mostRecentParams,
      bookmarkParams: bookmarkUrlParams,
      yamlParams: exploreUrlParamsFromYAMLConfig,
      rillDefaultParams: rillDefaultExploreUrlParams,
      backButtonUsed,
    });
    addEphemeralMeasureLibraryToUrlParams(urlParams, libraryEphemeralMeasures);
    return urlParams;
  }

  /**
   * Loads the per-metrics-view ad-hoc measure library, validated against the
   * current spec so definitions referencing measures this explore does not
   * expose are dropped. Empty when persisted state is disabled.
   */
  private loadEphemeralMeasureLibrary(
    metricsViewSpec: V1MetricsViewSpec,
    exploreSpec: V1ExploreSpec,
  ): EphemeralMeasureDef[] {
    if (this.disableMostRecentDashboardState || !exploreSpec.metricsView) {
      return [];
    }
    const library = loadEphemeralMeasureLibrary(
      exploreSpec.metricsView,
      this.storageNamespacePrefix,
    );
    if (!library.length) return [];
    const libraryState: Partial<ExploreState> = { ephemeralMeasures: library };
    validateAndCleanExploreState(metricsViewSpec, exploreSpec, libraryState);
    return libraryState.ephemeralMeasures ?? [];
  }
}
