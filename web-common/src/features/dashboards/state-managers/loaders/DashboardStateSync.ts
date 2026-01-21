import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateDataLoader";
import { saveMostRecentPartialExploreState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/most-recent-explore-state";
import {
  metricsExplorerStore,
  useExploreState,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { resolveTimeRanges } from "@rilldata/web-common/features/dashboards/time-controls/rill-time-ranges";
import {
  createTimeControlStoreFromName,
  type TimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { updateExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store";
import { getCleanedUrlParamsForGoto } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { createRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import type { AfterNavigate } from "@sveltejs/kit";
import { getContext, setContext } from "svelte";
import { derived, get, type Readable } from "svelte/store";
import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";

export const DASHBOARD_STATE_SYNC_KEY = Symbol("state-sync");

/**
 * Keeps explore state and url in sync.
 * If there is no explore state present initialised the explore state using DashboardStateDataLoader.initExploreState.
 *
 * prevUrl is used to make sure there is no redirect loop.
 */
export class DashboardStateSync {
  private readonly exploreStore: Readable<ExploreState | undefined>;
  private readonly timeControlStore: TimeControlStore;
  // Cached url params for a rill opinionated dashboard defaults. Used to remove params from url.
  // To avoid converting the default explore state to url evey time it is needed we maintain a cached version here.
  private readonly rillDefaultExploreURLParams: CompoundQueryResult<URLSearchParams>;

  private readonly unsubInit: (() => void) | undefined;
  private readonly unsubExploreState: (() => void) | undefined;

  private initialized = false;
  // There can be cases when updating either the url or the state can impact the code handling the other part.
  // So we need a lock to make sure an update doesn't trigger the counterpart code.
  private updating = false;

  public static getFromContext() {
    return getContext<DashboardStateSync>(DASHBOARD_STATE_SYNC_KEY);
  }

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

    this.rillDefaultExploreURLParams = createRillDefaultExploreUrlParams(
      dataLoader.validSpecQuery,
      dataLoader.fullTimeRangeQuery,
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

    setContext(DASHBOARD_STATE_SYNC_KEY, this);
  }

  public teardown() {
    this.unsubInit?.();
    this.unsubExploreState?.();
  }

  public getUrlForExploreState(exploreState: ExploreState) {
    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const exploreSpec = validSpecData?.explore ?? {};
    const pageState = get(page);
    const { data: rillDefaultExploreURLParams } = get(
      this.rillDefaultExploreURLParams,
    );
    // Type-safety
    if (!rillDefaultExploreURLParams) return pageState.url;

    const timeControlsState = get(this.timeControlStore);

    const redirectUrl = new URL(pageState.url);
    const exploreStateParams = getCleanedUrlParamsForGoto(
      exploreSpec,
      exploreState,
      timeControlsState,
      rillDefaultExploreURLParams,
      pageState.url,
    );

    redirectUrl.search = exploreStateParams.toString();

    return redirectUrl;
  }

  /**
   * Initializes the dashboard store.
   * If the url needs to change to match the init then we replace the current url with the new url.
   */
  private async handleExploreInit(initExploreState: ExploreState) {
    // If this is re-triggered any of the dependant query was refetched, then we need to make sure this is not run again.
    if (this.initialized) return;

    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
    const { data: rillDefaultExploreURLParams } = get(
      this.rillDefaultExploreURLParams,
    );
    // Ensure dashboard data is loaded before we proceed.
    if (!rillDefaultExploreURLParams) return;

    const pageState = get(page);

    if (metricsViewSpec.timeDimension && !import.meta.env.VITEST) {
      // Resolve start/end by making a network call.
      [
        initExploreState.selectedTimeRange,
        // initExploreState.selectedComparisonTimeRange,
      ] = await resolveTimeRanges(
        exploreSpec,
        [
          initExploreState.selectedTimeRange,
          // initExploreState.selectedComparisonTimeRange,
        ],
        initExploreState.selectedTimezone,
      );
    }

    // Init the store with state we got from dataLoader
    metricsExplorerStore.init(this.exploreName, initExploreState);
    // Get time controls state after explore state is initialized.
    const timeControlsState = get(this.timeControlStore);
    // Get the updated url params. If we merged state other than the url we would need to navigate to it.
    const redirectUrl = this.getUrlForExploreState(initExploreState);

    // Update session storage with the initial state
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      initExploreState,
      exploreSpec,
      timeControlsState,
    );
    if (!this.dataLoader.disableMostRecentDashboardState) {
      // Update "most recent explore state" with the initial state
      saveMostRecentPartialExploreState(
        this.exploreName,
        this.extraPrefix,
        initExploreState,
      );
    }

    // If the current url same as the new url then there is no need to do anything
    if (redirectUrl.search === pageState.url.search) {
      this.initialized = true;
      return;
    }

    // Else navigate to the new url.
    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    await goto(redirectUrl, {
      replaceState: true,
      state: pageState.state,
    });
    this.initialized = true;
  }

  /**
   * The decision to get the exploreState from url params depends on the navigation type.
   * This will be called from an afterNavigation callback.
   */
  public async handleURLChange(
    urlSearchParams: URLSearchParams,
    type: AfterNavigate["type"],
  ) {
    // Since we call this in afterNavigation, there could be a scenario where navigation completes but data for init isnt loaded yet.
    // Init already incorporates the url into the state so we can skip this processing.
    if (this.updating || !this.initialized) return;
    this.updating = true;

    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const metricsViewSpec = validSpecData?.metricsView ?? {};
    const exploreSpec = validSpecData?.explore ?? {};
    const { data: rillDefaultExploreURLParams } = get(
      this.rillDefaultExploreURLParams,
    );
    // Type-safety
    if (!rillDefaultExploreURLParams) return;

    const partialExplore = this.dataLoader.getExploreStateFromURLParams(
      urlSearchParams,
      type,
    );
    // This can be undefined when one of the queries has not loaded yet.
    // Rest of the code can be indeterminate when queries have not loaded.
    // This shouldn't ideally happen.
    if (!partialExplore) return;

    const pageState = get(page);

    if (metricsViewSpec.timeDimension && !import.meta.env.VITEST) {
      // Resolve start/end by making a network call.
      [
        partialExplore.selectedTimeRange,
        // partialExplore.selectedComparisonTimeRange,
      ] = await resolveTimeRanges(
        exploreSpec,
        [
          partialExplore.selectedTimeRange,
          // partialExplore.selectedComparisonTimeRange,
        ],
        partialExplore.selectedTimezone,
      );
    }

    // Merge the partial state from url into the store
    metricsExplorerStore.mergePartialExplorerEntity(
      this.exploreName,
      partialExplore,
      metricsViewSpec,
    );
    // Get time controls state after explore state is updated.
    const timeControlsState = get(this.timeControlStore);
    // Get the updated URL, this could be different from the page url if we added extra state.
    // The extra state could come from session storage, home bookmark or yaml defaults
    const redirectUrl = this.getUrlForExploreState(partialExplore);

    // Get the full updated state and save to session storage
    const updatedExploreState =
      get(metricsExplorerStore).entities[this.exploreName];
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      updatedExploreState,
      exploreSpec,
      timeControlsState,
    );
    if (!this.dataLoader.disableMostRecentDashboardState) {
      // Update "most recent explore state" with updated state from url
      saveMostRecentPartialExploreState(
        this.exploreName,
        this.extraPrefix,
        updatedExploreState,
      );
    }

    this.updating = false;
    // If the url doesn't need to be changed further then we can skip the goto
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

  /**
   * Called when the state is updated outside of this class, through an action for example.
   *
   * This will check if the url needs to be changed and will navigate to the new url.
   */
  private async gotoNewState(exploreState: ExploreState) {
    // Updating state either in handleExploreInit or handleURLChange will synchronously update the state triggering this function.
    // Since those methods handle redirect themselves we need to skip this logic.
    // Those methods need to replace the current URL while this does a direct navigation.
    if (this.updating) return;
    this.updating = true;

    const { data: validSpecData } = get(this.dataLoader.validSpecQuery);
    const exploreSpec = validSpecData?.explore ?? {};
    const timeControlsState = get(this.timeControlStore);

    const pageState = get(page);

    // Get the new url params for the updated state
    const newUrl = this.getUrlForExploreState(exploreState);

    // Update the session storage with the new explore state.
    updateExploreSessionStore(
      this.exploreName,
      this.extraPrefix,
      exploreState,
      exploreSpec,
      timeControlsState,
    );
    if (!this.dataLoader.disableMostRecentDashboardState) {
      // Update "most recent explore state" with updated state.
      // Since we do not update the state per action we do it here as blanket update.
      saveMostRecentPartialExploreState(
        this.exploreName,
        this.extraPrefix,
        exploreState,
      );
    }

    // If the state didnt result in a new url then skip goto.
    // This avoids adding redundant urls to the history.
    if (newUrl.search === pageState.url.search) {
      this.updating = false;
      return;
    }

    // dashboard changed so we should update the url
    await goto(newUrl);
    this.updating = false;
  }
}
