import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getExploreStateFromSessionStorage } from "@rilldata/web-common/features/dashboards/state-managers/loaders/get-explore-state-from-session-storage";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getExploreStatesFromYaml } from "@rilldata/web-common/features/dashboards/state-managers/loaders/get-explore-states-from-yaml";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { AfterNavigate } from "@sveltejs/kit";
import { derived, get, type Readable } from "svelte/store";

export class DashboardStateDataLoader {
  // These can be used to show a loading status
  public readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  public readonly fullTimeRangeQuery: ReturnType<
    typeof useMetricsViewTimeRange
  >;

  public readonly initExploreState: Readable<MetricsExplorerEntity | undefined>;
  public readonly exploreStatesFromSpecQuery: CompoundQueryResult<
    ReturnType<typeof getExploreStatesFromYaml>
  >;

  private readonly exploreStatesFromURLParamsQuery: Readable<
    | {
        partialExploreStateFromUrl: Partial<MetricsExplorerEntity>;
        partialExploreStateFromUrlForInit:
          | Partial<MetricsExplorerEntity>
          | undefined;
        exploreStateFromSessionStorage:
          | Partial<MetricsExplorerEntity>
          | undefined;
        errors: Error[];
      }
    | undefined
  >;

  public constructor(
    instanceId: string,
    metricsViewName: string,
    private readonly exploreName: string,
    private readonly extraPrefix: string | undefined,
    otherSourcesOfState: Readable<Partial<MetricsExplorerEntity> | undefined>[],
  ) {
    this.validSpecQuery = useExploreValidSpec(instanceId, exploreName);
    this.fullTimeRangeQuery = useMetricsViewTimeRange(
      instanceId,
      metricsViewName,
    );

    this.exploreStatesFromSpecQuery = getCompoundQuery(
      [this.validSpecQuery, this.fullTimeRangeQuery],
      ([validSpecResp, metricsViewTimeRangeResp]) => {
        const metricsViewSpec = validSpecResp?.metricsView ?? {};
        const exploreSpec = validSpecResp?.explore ?? {};

        // Safeguard to make sure time range summary is loaded when time dimension is present.
        if (
          metricsViewSpec.timeDimension &&
          !metricsViewTimeRangeResp?.timeRangeSummary
        ) {
          return undefined;
        }

        return getExploreStatesFromYaml(
          metricsViewSpec,
          exploreSpec,
          metricsViewTimeRangeResp ?? {},
          exploreName,
          extraPrefix,
        );
      },
    );

    this.exploreStatesFromURLParamsQuery = derived(
      [this.validSpecQuery, this.exploreStatesFromSpecQuery, page],
      ([validSpecResp, exploreStatesFromSpecs, page]) => {
        if (!validSpecResp.data || !exploreStatesFromSpecs.data)
          return undefined;

        const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
        const exploreSpec = validSpecResp.data?.explore ?? {};
        const explorePresetFromYAMLConfig =
          exploreStatesFromSpecs.data?.explorePresetFromYAMLConfig ?? {};

        return this.getExploreStatesFromURLParams(
          this.exploreName,
          this.extraPrefix,
          page.url.searchParams,
          metricsViewSpec,
          exploreSpec,
          explorePresetFromYAMLConfig,
        );
      },
    );

    this.initExploreState = derived(
      [
        this.exploreStatesFromSpecQuery,
        this.exploreStatesFromURLParamsQuery,
        ...otherSourcesOfState,
      ],
      ([
        exploreStatesFromSpecs,
        exploreStatesFromURLParams,
        ...otherSourcesOfState
      ]) => {
        if (!exploreStatesFromSpecs.data || !exploreStatesFromURLParams)
          return undefined;

        const {
          defaultExploreState,
          exploreStateFromYAMLConfig,
          mostRecentPartialExploreState,
        } = exploreStatesFromSpecs.data;
        const {
          exploreStateFromSessionStorage,
          partialExploreStateFromUrlForInit,
        } = exploreStatesFromURLParams;

        // Select the 1st available exploreState from "otherSourcesOfState"
        const firstStateFromOtherSources = otherSourcesOfState.find(
          (state) => state !== undefined,
        );

        const initExploreState = {
          // Since this is a complete state, we need the complete default explore state which works as a base.
          ...defaultExploreState,
          // 1st priority is the state from session storage.
          // TODO: since this only loads on certain params present in the url it should be merged with convertURLSearchParamsToExploreState
          ...(exploreStateFromSessionStorage ??
            // Next priority is the state loaded from url params. It will be undefined if there are no params.
            partialExploreStateFromUrlForInit ??
            // Next priority is the most recent state stored in local storage
            mostRecentPartialExploreState ??
            // Next priority is one of the other source defined.
            // For cloud dashboard it would be home bookmark if present.
            // For shared url it would be the saved state in token
            firstStateFromOtherSources ??
            // Finally the state from yaml is used
            exploreStateFromYAMLConfig),
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
    const exploreStatesFromSpec = get(this.exploreStatesFromSpecQuery).data;
    if (!exploreStatesFromSpec) return undefined;

    // Pressing back button and going back to empty url state should not restore from session store
    const backButtonUsed = type === "popstate";
    const skipSessionStorage = backButtonUsed;

    const { exploreStateFromSessionStorage, partialExploreStateFromUrl } =
      this.getExploreStatesFromURLParams(
        this.exploreName,
        this.extraPrefix,
        urlSearchParams,
        metricsViewSpec,
        exploreSpec,
        exploreStatesFromSpec.explorePresetFromYAMLConfig,
      );

    if (skipSessionStorage) return partialExploreStateFromUrl;
    return exploreStateFromSessionStorage ?? partialExploreStateFromUrl;
  }

  private getExploreStatesFromURLParams(
    exploreName: string,
    prefix: string | undefined,
    searchParams: URLSearchParams,
    metricsViewSpec: V1MetricsViewSpec,
    exploreSpec: V1ExploreSpec,
    defaultExplorePreset: V1ExplorePreset,
  ) {
    const { partialExploreState: partialExploreStateFromUrl, errors } =
      convertURLSearchParamsToExploreState(
        searchParams,
        metricsViewSpec,
        exploreSpec,
        defaultExplorePreset,
      );
    const partialExploreStateFromUrlForInit =
      searchParams.size === 0 ? undefined : partialExploreStateFromUrl;

    const exploreStateFromSessionStorage = getExploreStateFromSessionStorage(
      exploreName,
      prefix,
      searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );

    return {
      partialExploreStateFromUrl,
      partialExploreStateFromUrlForInit,
      exploreStateFromSessionStorage,
      errors,
    };
  }
}
