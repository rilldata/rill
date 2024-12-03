<script lang="ts">
  import { page } from "$app/stores";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { convertURLToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToExploreState";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
  import { getLocalUserPreferencesState } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  /**
   * Temporary wrapper component that mimics the parsing and loading of url into metrics.
   * This is ideally done in the loader function but for embed it needs to be done here.
   * TODO: Fix embed to update the URL and get rid of this.
   */

  const { exploreName, metricsViewName, validSpecStore } = getStateManagers();

  $: exploreSpec = $validSpecStore.data?.explore ?? {};
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};
  $: metricsViewTimeRange = useMetricsViewTimeRange(
    $runtime.instanceId,
    $metricsViewName,
  );
  $: defaultExplorePreset = getDefaultExplorePreset(
    exploreSpec,
    getLocalUserPreferencesState($exploreName),
    $metricsViewTimeRange.data,
  );

  // Get Explore state from URL params
  let partialExploreState: Partial<MetricsExplorerEntity> = {};
  function parseUrl(url: URL) {
    const { partialExploreState: partialExploreStateFromUrl } =
      convertURLToExploreState(
        url.searchParams,
        metricsViewSpec,
        exploreSpec,
        defaultExplorePreset,
      );
    partialExploreState = partialExploreStateFromUrl;
  }

  // only reactive to url
  $: parseUrl($page.url);
</script>

<DashboardURLStateSync {defaultExplorePreset} {partialExploreState}>
  <slot />
</DashboardURLStateSync>
