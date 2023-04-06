<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";

  const metricViewName: string = $page.params.dashboard;
  const stateSyncManager = new StateSyncManager(metricViewName);

  $: metricsExplorer = useDashboardStore(metricViewName);

  $: if (metricsExplorer) {
    stateSyncManager.handleStateChange(metricsExplorer);
  }
  $: if ($page) {
    stateSyncManager.handleUrlChange();
  }

  // TODO: add 404 logic as in `web-local`'s `dashboard/[name]/+page.svelte`
</script>

<svelte:head>
  <title>Rill | {metricViewName}</title>
</svelte:head>

<Dashboard {metricViewName} hasTitle={false} />
