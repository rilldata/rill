<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { getUrlFromMetricsExplorer } from "@rilldata/web-common/features/dashboards/url-state/toUrl";

  export let partialMetrics: Partial<MetricsExplorerEntity>;

  const ctx = getStateManagers();
  const { metricsViewName, dashboardStore, validSpecStore } = ctx;
  $: exploreSpec = $validSpecStore.data?.explore;
  $: metricsSpec = $validSpecStore.data?.metricsView;

  let prevUrl = "";
  function gotoNewState() {
    if (!exploreSpec) return;

    const u = new URL(
      `${$page.url.protocol}//${$page.url.host}${$page.url.pathname}`,
    );
    getUrlFromMetricsExplorer(
      $dashboardStore,
      u.searchParams,
      exploreSpec,
      exploreSpec.defaultPreset ?? {},
    );
    const newUrl = u.toString();
    if (window.location.href !== newUrl) {
      void goto(newUrl);
    }
  }

  $: if ($dashboardStore && exploreSpec) {
    gotoNewState();
  }

  $: if (partialMetrics && metricsSpec && prevUrl !== window.location.href) {
    metricsExplorerStore.syncFromUrlParams(
      $metricsViewName,
      partialMetrics,
      metricsSpec,
    );
    prevUrl = window.location.href;
  }
</script>

<slot />
