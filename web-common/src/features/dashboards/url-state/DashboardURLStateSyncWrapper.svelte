<script lang="ts">
  import { page } from "$app/stores";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { restorePersistedDashboardState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import {
    convertPresetToExploreState,
    convertURLToExploreState,
  } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";
  import type { V1ExplorePreset } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  /**
   * Temporary wrapper component that mimics the parsing and loading of url into metrics.
   * This is ideally done in the loader function but for embed it needs to be done here.
   * TODO: Fix embed to update the URL and get rid of this.
   */

  const { exploreName, metricsViewName, validSpecStore } = getStateManagers();

  const orgName = $page.params.organization;
  const projectName = $page.params.project;
  const storeKeyPrefix =
    orgName && projectName ? `__${orgName}__${projectName}` : "";

  $: exploreSpec = $validSpecStore.data?.explore ?? {};
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};
  $: metricsViewTimeRange = useMetricsViewTimeRange(
    $runtime.instanceId,
    $metricsViewName,
  );
  $: defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    $metricsViewTimeRange.data,
  );
  let initExploreState: Partial<MetricsExplorerEntity> = {};
  let initUrlSearch = "";
  $: {
    ({ partialExploreState: initExploreState } = convertPresetToExploreState(
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    ));

    let initLoadedOutsideOfURL = false;
    const stateFromLocalStorage = restorePersistedDashboardState(
      exploreSpec,
      storeKeyPrefix + $exploreName,
    );
    if (stateFromLocalStorage) {
      initLoadedOutsideOfURL = true;
      Object.assign(initExploreState, stateFromLocalStorage);
    }

    initUrlSearch = initLoadedOutsideOfURL
      ? getUpdatedUrlForExploreState(
          exploreSpec,
          defaultExplorePreset,
          initExploreState,
          new URLSearchParams(),
        )
      : "";
  }

  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  let urlSearchForPartial = "";
  function parseUrl(url: URL, defaultExplorePreset: V1ExplorePreset) {
    // Get Explore state from URL params
    const {
      partialExploreState: partialExploreStateFromUrl,
      urlSearchForPartial: _urlSearchForPartial,
    } = convertURLToExploreState(
      $exploreName,
      storeKeyPrefix,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    urlSearchForPartial = _urlSearchForPartial;
  }

  // only reactive to url and defaultExplorePreset
  $: parseUrl($page.url, defaultExplorePreset);
</script>

{#if !$validSpecStore.isLoading && !$metricsViewTimeRange.isLoading}
  <DashboardURLStateSync
    metricsViewName={$metricsViewName}
    exploreName={$exploreName}
    {initExploreState}
    {defaultExplorePreset}
    {initUrlSearch}
    {partialExploreState}
    {urlSearchForPartial}
  >
    <slot />
  </DashboardURLStateSync>
{/if}
