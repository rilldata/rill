<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getUrlFromMetricsExplorer } from "@rilldata/web-common/features/dashboards/url-state/toUrl";

  export let searchParams: URLSearchParams;

  const ctx = getStateManagers();
  const { metricsViewName, dashboardStore } = ctx;
  const metricsView = useMetricsView(ctx);

  $: if ($dashboardStore && $metricsView.data) {
    const u = new URL(
      `${$page.url.protocol}//${$page.url.host}${$page.url.pathname}`,
    );
    getUrlFromMetricsExplorer(
      $dashboardStore,
      u.searchParams,
      $metricsView.data,
    );
    void goto(u.toString());
  }

  $: if (searchParams && $metricsView.data) {
    metricsExplorerStore.syncFromUrlParams(
      $metricsViewName,
      searchParams,
      $metricsView.data,
    );
  }
</script>

<slot />
