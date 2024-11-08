<script lang="ts">
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { convertURLToMetricsExplore } from "@rilldata/web-common/features/dashboards/url-state/convertPresetToMetricsExplore";
  import DashboardURLStateSync from "@rilldata/web-common/features/dashboards/url-state/DashboardURLStateSync.svelte";
  import { getBasePreset } from "@rilldata/web-common/features/dashboards/url-state/getBasePreset";
  import { getLocalUserPreferencesState } from "@rilldata/web-common/features/dashboards/user-preferences";

  /**
   * Temporary wrapper component that mimics the parsing and loading of url into metrics.
   * This is ideally done in the loader function but for embed it needs to be done here.
   * TODO: Fix embed to update the URL and get rid of this.
   */

  const { exploreName, validSpecStore } = getStateManagers();

  $: exploreSpec = $validSpecStore.data?.explore ?? {};
  $: metricsViewSpec = $validSpecStore.data?.metricsView ?? {};
  $: basePreset = getBasePreset(
    exploreSpec,
    getLocalUserPreferencesState($exploreName),
  );

  let partialMetrics: Partial<MetricsExplorerEntity> = {};
  function parseUrl(url: URL) {
    const { entity } = convertURLToMetricsExplore(
      url.searchParams,
      metricsViewSpec,
      exploreSpec,
      basePreset,
    );
    partialMetrics = entity;
  }

  // only reactive to url
  $: parseUrl($page.url);
</script>

<DashboardURLStateSync {basePreset} {partialMetrics}>
  <slot />
</DashboardURLStateSync>
