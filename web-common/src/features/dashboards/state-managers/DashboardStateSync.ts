import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/DashboardStateDataLoader";
import {
  metricsExplorerStore,
  useExploreState,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  createTimeControlStoreFromName,
  type TimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  convertExploreStateToURLSearchParams,
  getUpdatedUrlForExploreState,
} from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { updateExploreSessionStore } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
import { saveMostRecentExploreState } from "@rilldata/web-common/features/dashboards/url-state/most-recent-explore-state";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import { derived, get, type Readable } from "svelte/store";

export class DashboardStateSync {
  private readonly exploreStore: Readable<MetricsExplorerEntity | undefined>;
  private readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  private readonly timeControlStore: TimeControlStore;

  private readonly unsubInit: (() => void) | undefined;
  private readonly unsubUrlChange: (() => void) | undefined;
  private readonly unsubExploreState: (() => void) | undefined;

  private initializing = false;
  private prevUrl = "";

  public constructor(
    instanceId: string,
    metricsViewName: string,
    private readonly exploreName: string,
    private readonly extraPrefix: string | undefined,
    private readonly dataLoader: DashboardStateDataLoader,
  ) {
    this.validSpecQuery = useExploreValidSpec(instanceId, exploreName);
    this.exploreStore = useExploreState(exploreName);
    this.timeControlStore = createTimeControlStoreFromName(
      instanceId,
      metricsViewName,
      exploreName,
    );

    this.unsubInit = derived(
      [dataLoader.initExploreState, this.exploreStore],
      (states) => states,
    ).subscribe(([initExploreState, exploreStore]) => {
      if (exploreStore || initExploreState?.activePage === undefined) return;
      void this.handleExploreInit(initExploreState);
    });

    this.unsubUrlChange = dataLoader.partialExploreState.subscribe(
      (partialExploreState) => {
        if (!partialExploreState) return;
        void this.handleURLChange(partialExploreState);
      },
    );

    this.unsubExploreState = this.exploreStore.subscribe((exploreState) => {
      if (!exploreState) return;
      void this.gotoNewState(exploreState);
    });
  }

  public teardown() {
    this.unsubInit?.();
    this.unsubUrlChange?.();
    this.unsubExploreState?.();
  }

  private handleExploreInit(initExploreState: MetricsExplorerEntity) {
    if (this.initializing) return;
    this.initializing = true;

    const { data: validSpecData } = get(this.validSpecQuery);
    const exploreSpec = validSpecData?.explore ?? {};
    const timeControlsState = get(this.timeControlStore);
    const pageState = get(page);
    const { data: exploreStatesFromSpecData } = get(
      this.dataLoader.exploreStatesFromSpecQuery,
    );
    const explorePresetFromYAMLConfig =
      exploreStatesFromSpecData?.explorePresetFromYAMLConfig ?? {};

    metricsExplorerStore.init(this.exploreName, initExploreState);
    const redirectUrl = new URL(pageState.url);
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      explorePresetFromYAMLConfig,
      initExploreState,
      pageState.url,
    );
    // update session store to make sure updated to url or the initial state is propagated to the session store
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      get(metricsExplorerStore).entities[this.exploreName],
      exploreSpec,
      timeControlsState,
    );
    this.prevUrl = redirectUrl.toString();

    if (redirectUrl.search === pageState.url.search) {
      return;
    }

    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    return goto(redirectUrl, {
      replaceState: true,
      state: pageState.state,
    });
  }

  private handleURLChange(partialExplore: Partial<MetricsExplorerEntity>) {
    const { data: validSpecData } = get(this.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
    const timeControlsState = get(this.timeControlStore);
    const pageState = get(page);
    const { data: exploreStatesFromSpecData } = get(
      this.dataLoader.exploreStatesFromSpecQuery,
    );
    const explorePresetFromYAMLConfig =
      exploreStatesFromSpecData?.explorePresetFromYAMLConfig ?? {};

    const redirectUrl = new URL(pageState.url);
    metricsExplorerStore.mergePartialExplorerEntity(
      this.exploreName,
      partialExplore,
      metricsViewSpec,
    );
    // if we added extra url params from sessionStorage then update the url
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      explorePresetFromYAMLConfig,
      partialExplore,
      pageState.url,
    );

    if (
      redirectUrl.search === pageState.url.search ||
      // redirect loop breaker
      (this.prevUrl && this.prevUrl === redirectUrl.toString())
    ) {
      this.prevUrl = redirectUrl.toString();
      return;
    }

    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      get(this.exploreStore)!,
      exploreSpec,
      timeControlsState,
    );
    this.prevUrl = redirectUrl.toString();
    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    void goto(redirectUrl, {
      replaceState: true,
      state: pageState.state,
    });
  }

  private gotoNewState(exploreState: MetricsExplorerEntity) {
    const { data: validSpecData } = get(this.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
    const timeControlsState = get(this.timeControlStore);
    const pageState = get(page);
    const { data: exploreStatesFromSpecData } = get(
      this.dataLoader.exploreStatesFromSpecQuery,
    );
    const explorePresetFromYAMLConfig =
      exploreStatesFromSpecData?.explorePresetFromYAMLConfig ?? {};

    const u = new URL(pageState.url);
    const exploreStateParams = convertExploreStateToURLSearchParams(
      exploreState,
      exploreSpec,
      timeControlsState,
      explorePresetFromYAMLConfig,
      u,
    );
    u.search = exploreStateParams.toString();
    const newUrl = u.toString();
    if (!this.prevUrl || this.prevUrl === newUrl) return;

    this.prevUrl = newUrl;
    // dashboard changed so we should update the url
    void goto(newUrl);
    // also update the session store
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      exploreState,
      exploreSpec,
      timeControlsState,
    );
    saveMostRecentExploreState(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      exploreState,
    );
  }
}
