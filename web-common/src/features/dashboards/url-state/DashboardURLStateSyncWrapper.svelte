<script lang="ts">
  import { page } from "$app/stores";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
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

  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  let loaded = false;
  function parseUrl(url: URL, defaultExplorePreset: V1ExplorePreset) {
    // Get Explore state from URL params
    const {
      partialExploreState: partialExploreStateFromUrl,
      loaded: loadedFromStorage,
    } = convertURLToExploreState(
      $exploreName,
      storeKeyPrefix,
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      defaultExplorePreset,
    );
    partialExploreState = partialExploreStateFromUrl;
    loaded = loadedFromStorage;
  }

  // only reactive to url and defaultExplorePreset
  $: parseUrl($page.url, defaultExplorePreset);
</script>

{#if !$validSpecStore.isLoading && !$metricsViewTimeRange.isLoading}
  <DashboardURLStateSync {defaultExplorePreset} {partialExploreState} {loaded}>
    <slot />
  </DashboardURLStateSync>
{/if}
