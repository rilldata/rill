import { page } from "$app/stores";
import {
  type CompoundQueryResult,
  getCompoundQuery,
} from "@rilldata/web-common/features/compound-query-result";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getExploreStatesFromSpecs } from "@rilldata/web-common/features/dashboards/url-state/get-explore-states-from-specs";
import {
  getExploreStatesFromURLParams,
  useExploreValidSpec,
} from "@rilldata/web-common/features/explores/selectors";
import { derived, type Readable } from "svelte/store";

export class DashboardStateDataLoader {
  // These can be used to show a loading status
  public readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  public readonly fullTimeRangeQuery: ReturnType<
    typeof useMetricsViewTimeRange
  >;

  public readonly initExploreState: Readable<MetricsExplorerEntity | undefined>;
  public readonly partialExploreState: Readable<
    Partial<MetricsExplorerEntity> | undefined
  >;

  public readonly exploreStatesFromSpecQuery: CompoundQueryResult<
    ReturnType<typeof getExploreStatesFromSpecs>
  >;

  private readonly exploreStatesFromURLParamsQuery: Readable<
    ReturnType<typeof getExploreStatesFromURLParams> | undefined
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
        return getExploreStatesFromSpecs(
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

        return getExploreStatesFromURLParams(
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
        const { exploreStateFromSessionStorage, partialExploreStateFromUrl } =
          exploreStatesFromURLParams;

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
            partialExploreStateFromUrl ??
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

    this.partialExploreState = derived(
      this.exploreStatesFromURLParamsQuery,
      (exploreStatesFromURLParams) => {
        return (
          // For partial state from url, 1st priority is the session storage.
          exploreStatesFromURLParams?.exploreStateFromSessionStorage ??
          // Next priority is the state from the url params
          exploreStatesFromURLParams?.partialExploreStateFromUrl
        );
      },
    );
  }
}
