import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateDataLoader";
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
import { saveMostRecentExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import type { AfterNavigate } from "@sveltejs/kit";
import { derived, get, type Readable } from "svelte/store";
import { updateExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-active-page-store";

export class DashboardStateSync {
  private readonly exploreStore: Readable<MetricsExplorerEntity | undefined>;
  private readonly validSpecQuery: ReturnType<typeof useExploreValidSpec>;
  private readonly timeControlStore: TimeControlStore;

  private readonly unsubInit: (() => void) | undefined;
  private readonly unsubExploreState: (() => void) | undefined;

  private initialized = false;
  private prevUrl: URL | undefined;

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
      [dataLoader.initExploreState],
      (states) => states,
    ).subscribe(([initExploreState]) => {
      // initExploreState is not ready yet. Either data is still loading or dashboard is not reconciled yet.
      if (initExploreState?.activePage === undefined) return;
      void this.handleExploreInit(initExploreState);
    });

    this.unsubExploreState = this.exploreStore.subscribe((exploreState) => {
      if (!exploreState || !this.initialized) return;
      void this.gotoNewState(exploreState);
    });
  }

  public teardown() {
    this.unsubInit?.();
    this.unsubExploreState?.();
  }

  private handleExploreInit(initExploreState: MetricsExplorerEntity) {
    if (this.initialized) return;
    this.initialized = true;

    const { data: validSpecData } = get(this.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
    const pageState = get(page);
    const { data: exploreStatesFromSpecData } = get(
      this.dataLoader.exploreStatesFromSpecQuery,
    );
    const explorePresetFromYAMLConfig =
      exploreStatesFromSpecData?.explorePresetFromYAMLConfig ?? {};

    metricsExplorerStore.init(this.exploreName, initExploreState);
    // Get time controls state after explore state is initialized.
    const timeControlsState = get(this.timeControlStore);
    const redirectUrl = new URL(pageState.url);
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      explorePresetFromYAMLConfig,
      initExploreState,
      pageState.url,
    );
    this.prevUrl = redirectUrl;

    if (redirectUrl.search === pageState.url.search) {
      return;
    }

    console.log("REPLACE:INIT", pageState.url.search, redirectUrl.search);
    const updatedExploreState =
      get(metricsExplorerStore).entities[this.exploreName];
    // update session store to make sure updated to url or the initial state is propagated to the session store
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      updatedExploreState,
    );
    saveMostRecentExploreState(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      updatedExploreState,
    );
    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    return goto(redirectUrl, {
      replaceState: true,
      state: pageState.state,
    });
  }

  // The decision to get the exploreState from url params depends on the navigation type.
  // This will be called from an afterNavigation callback.
  public handleURLChange(
    urlSearchParams: URLSearchParams,
    type: AfterNavigate["type"],
  ) {
    if (!get(metricsExplorerStore).entities[this.exploreName]) return;

    const partialExplore = this.dataLoader.getExploreStateFromURLParams(
      urlSearchParams,
      type,
    );
    if (!partialExplore) return;

    const { data: validSpecData } = get(this.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
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
    // Get time controls state after explore state is updated.
    const timeControlsState = get(this.timeControlStore);
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
      (this.prevUrl && this.prevUrl.search === redirectUrl.search)
    ) {
      this.prevUrl = redirectUrl;
      return;
    }

    const updatedExploreState =
      get(metricsExplorerStore).entities[this.exploreName];
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      updatedExploreState,
    );
    saveMostRecentExploreState(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      updatedExploreState,
    );
    console.log("REPLACE:URL", this.prevUrl?.search, redirectUrl.search);
    this.prevUrl = redirectUrl;
    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    return goto(redirectUrl, {
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

    const newUrl = new URL(pageState.url);
    const exploreStateParams = convertExploreStateToURLSearchParams(
      exploreState,
      exploreSpec,
      timeControlsState,
      explorePresetFromYAMLConfig,
      newUrl,
    );
    newUrl.search = exploreStateParams.toString();
    if (!this.prevUrl || this.prevUrl.search === newUrl.search) {
      console.log("NOGO", this.prevUrl?.search, newUrl.search);
      return;
    }

    // also update the session store
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      exploreState,
    );
    saveMostRecentExploreState(
      this.exploreName,
      this.extraPrefix,
      metricsViewSpec,
      exploreSpec,
      timeControlsState,
      exploreState,
    );
    console.log("GOTO:STATE", this.prevUrl?.search, newUrl.search);
    this.prevUrl = newUrl;
    // dashboard changed so we should update the url
    return goto(newUrl);
  }
}
