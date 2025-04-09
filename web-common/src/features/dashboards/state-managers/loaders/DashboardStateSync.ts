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
  type TimeControlState,
  type TimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { updateExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import type { AfterNavigate } from "@sveltejs/kit";
import { derived, get, type Readable } from "svelte/store";

/**
 * Keeps explore state and url in sync.
 * If there is no explore state present initialised the explore state using DashboardStateDataLoader.initExploreState.
 *
 * prevUrl is used to make sure there is no redirect loop.
 */
export class DashboardStateSync {
  private readonly exploreStore: Readable<MetricsExplorerEntity | undefined>;
  private readonly timeControlStore: TimeControlStore;

  private readonly unsubInit: (() => void) | undefined;
  private readonly unsubExploreState: (() => void) | undefined;

  private initialized = false;
  // There can be cases when updating either the url or the state can impact the code handling the other part.
  // So we need a lock to make sure an update doesn't trigger the counterpart code.
  private updating = false;

  public constructor(
    instanceId: string,
    metricsViewName: string,
    private readonly exploreName: string,
    private readonly extraPrefix: string | undefined,
    private readonly dataLoader: DashboardStateDataLoader,
  ) {
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
      if (
        initExploreState.isLoading ||
        initExploreState.data?.activePage === undefined
      )
        return;
      void this.handleExploreInit(initExploreState.data);
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

    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const exploreSpec = validSpecData?.explore ?? {};
    const pageState = get(page);
    const { data: defaultExploreUrlParams } = get(
      this.dataLoader.defaultExploreUrlParams,
    );
    if (!defaultExploreUrlParams) return;

    metricsExplorerStore.init(this.exploreName, initExploreState);
    // Get time controls state after explore state is initialized.
    const timeControlsState = get(this.timeControlStore);
    const redirectUrl = new URL(pageState.url);
    redirectUrl.search = this.getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      defaultExploreUrlParams,
      initExploreState,
      pageState.url,
    );

    if (redirectUrl.search === pageState.url.search) {
      return;
    }

    const updatedExploreState =
      get(metricsExplorerStore).entities[this.exploreName];
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      updatedExploreState,
      exploreSpec,
      timeControlsState,
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
    if (!get(metricsExplorerStore).entities[this.exploreName] || this.updating)
      return;

    const partialExplore = this.dataLoader.getExploreStateFromURLParams(
      urlSearchParams,
      type,
    );
    // This can be undefined when one of the queries has not loaded yet.
    // Rest of the code can be indeterminate when queries have not loaded.
    // This shouldn't ideally happen.
    if (!partialExplore) return;

    this.updating = true;
    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
    const pageState = get(page);
    const { data: defaultExploreUrlParams } = get(
      this.dataLoader.defaultExploreUrlParams,
    );
    if (!defaultExploreUrlParams) return;

    const redirectUrl = new URL(pageState.url);
    metricsExplorerStore.mergePartialExplorerEntity(
      this.exploreName,
      partialExplore,
      metricsViewSpec,
    );
    // Get time controls state after explore state is updated.
    const timeControlsState = get(this.timeControlStore);
    // if we added extra url params from session storage then update the url
    redirectUrl.search = this.getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      defaultExploreUrlParams,
      partialExplore,
      pageState.url,
    );

    const updatedExploreState =
      get(metricsExplorerStore).entities[this.exploreName];
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      updatedExploreState,
      exploreSpec,
      timeControlsState,
    );

    this.updating = false;
    // redirect loop breaker
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

  private gotoNewState(exploreState: MetricsExplorerEntity) {
    if (this.updating) return;
    this.updating = true;

    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const exploreSpec = validSpecData?.explore ?? {};
    const timeControlsState = get(this.timeControlStore);
    const pageState = get(page);
    const { data: defaultExploreUrlParams } = get(
      this.dataLoader.defaultExploreUrlParams,
    );
    if (!defaultExploreUrlParams) return;

    const newUrl = new URL(pageState.url);
    const exploreStateParams = convertPartialExploreStateToUrlParams(
      exploreState,
      exploreSpec,
      timeControlsState,
      defaultExploreUrlParams,
      newUrl,
    );
    newUrl.search = exploreStateParams.toString();

    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      exploreState,
      exploreSpec,
      timeControlsState,
    );

    this.updating = false;
    // redirect loop breaker
    if (newUrl.search === pageState.url.search) {
      return;
    }

    // dashboard changed so we should update the url
    return goto(newUrl);
  }

  /**
   * Sometimes data is loaded from sources other than the url.
   * In that case update the URL to make sure the state matches the current url.
   */
  private getUpdatedUrlForExploreState(
    exploreSpec: V1ExploreSpec,
    timeControlsState: TimeControlState | undefined,
    defaultExploreUrlParams: URLSearchParams,
    partialExploreState: Partial<MetricsExplorerEntity>,
    url: URL,
  ): string {
    // Create params from the explore state
    const stateParams = convertPartialExploreStateToUrlParams(
      partialExploreState,
      exploreSpec,
      timeControlsState,
      defaultExploreUrlParams,
      url,
    );

    // Filter out the default view parameter if needed
    url.searchParams.forEach((value, key) => {
      if (
        key === ExploreStateURLParams.WebView &&
        value === defaultExploreUrlParams.get(ExploreStateURLParams.WebView)
      ) {
        return; // Skip this parameter
      }
      stateParams.set(key, value);
    });

    return stateParams.toString();
  }
}
