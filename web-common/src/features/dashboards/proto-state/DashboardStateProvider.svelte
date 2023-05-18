<script lang="ts">
  import { page } from "$app/stores";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";

  export let metricViewName: string;

  $: metricsExplorer = useDashboardStore(metricViewName);
  $: stateSyncManager = new StateSyncManager(metricViewName);
  $: if ($metricsExplorer) {
    stateSyncManager.handleStateChange($metricsExplorer);
  }
  $: if ($page) {
    stateSyncManager.handleUrlChange();
  }
</script>

<slot />
