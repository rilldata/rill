<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import {
    getTimeControlState,
    type TimeControlState,
  } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    convertExploreStateToURLSearchParams,
    getUpdatedUrlForExploreState,
  } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
  import { updateExploreSessionStore } from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-store";
  import { type V1ExplorePreset } from "@rilldata/web-common/runtime-client";
  import { get } from "svelte/store";

  export let exploreName: string;
  export let extraKeyPrefix: string | undefined = undefined;
  export let defaultExplorePreset: V1ExplorePreset;
  export let initExploreState: Partial<MetricsExplorerEntity>;
  export let partialExploreState: Partial<MetricsExplorerEntity>;

  const { dashboardStore, validSpecStore, timeRangeSummaryStore } =
    getStateManagers();
  $: exploreSpec = $validSpecStore.data?.explore;
  $: metricsSpec = $validSpecStore.data?.metricsView;

  $: ({ data: timeRangeSummaryResp } = $timeRangeSummaryStore);

  let timeControlsState: TimeControlState | undefined = undefined;
  $: if (metricsSpec && exploreSpec && $dashboardStore) {
    timeControlsState = getTimeControlState(
      metricsSpec,
      exploreSpec,
      timeRangeSummaryResp?.timeRangeSummary,
      $dashboardStore,
    );
  }

  let initializing = false;
  let prevUrl = "";

  $: if (!$dashboardStore && initExploreState.activePage !== undefined) {
    void handleExploreInit();
  }

  $: if (partialExploreState) {
    void handleURLChange(partialExploreState);
  }

  // reactive to only dashboardStore
  // but gotoNewState checks other fields
  $: if ($dashboardStore) {
    void gotoNewState();
  }

  async function handleExploreInit() {
    if (initializing || !exploreSpec || !metricsSpec) return;
    initializing = true;

    metricsExplorerStore.init(exploreName, initExploreState);
    timeControlsState ??= getTimeControlState(
      metricsSpec,
      exploreSpec,
      timeRangeSummaryResp?.timeRangeSummary,
      get(metricsExplorerStore).entities[exploreName],
    );
    const redirectUrl = new URL($page.url);
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      defaultExplorePreset,
      initExploreState,
      $page.url,
    );
    // update session store to make sure updated to url or the initial state is propagated to the session store
    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      get(metricsExplorerStore).entities[exploreName],
      exploreSpec,
      timeControlsState,
    );
    prevUrl = redirectUrl.toString();

    if (redirectUrl.search === $page.url.search) {
      return;
    }

    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    return goto(redirectUrl, {
      replaceState: true,
      state: $page.state,
    });
  }

  function handleURLChange(partialExplore: Partial<MetricsExplorerEntity>) {
    if (!metricsSpec || !exploreSpec) return;

    const redirectUrl = new URL($page.url);
    metricsExplorerStore.mergePartialExplorerEntity(
      exploreName,
      partialExplore,
      metricsSpec,
    );
    // if we added extra url params from sessionStorage then update the url
    redirectUrl.search = getUpdatedUrlForExploreState(
      exploreSpec,
      timeControlsState,
      defaultExplorePreset,
      partialExplore,
      $page.url,
    );

    if (
      redirectUrl.search === $page.url.search ||
      // redirect loop breaker
      (prevUrl && prevUrl === redirectUrl.toString())
    ) {
      prevUrl = redirectUrl.toString();
      return;
    }

    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      $dashboardStore,
      exploreSpec,
      timeControlsState,
    );
    prevUrl = redirectUrl.toString();
    // using `replaceState` directly messes up the navigation entries,
    // `from` and `to` have the old url before being replaced in `afterNavigate` calls leading to incorrect handling.
    void goto(redirectUrl, {
      replaceState: true,
      state: $page.state,
    });
  }

  function gotoNewState() {
    if (!exploreSpec) return;

    const u = new URL($page.url);
    const exploreStateParams = convertExploreStateToURLSearchParams(
      $dashboardStore,
      exploreSpec,
      timeControlsState,
      defaultExplorePreset,
      u,
    );
    u.search = exploreStateParams.toString();
    const newUrl = u.toString();
    if (!prevUrl || prevUrl === newUrl) return;

    prevUrl = newUrl;
    // dashboard changed so we should update the url
    void goto(newUrl);
    // also update the session store
    updateExploreSessionStore(
      exploreName,
      extraKeyPrefix,
      $dashboardStore,
      exploreSpec,
      timeControlsState,
    );
  }
</script>

<slot />
