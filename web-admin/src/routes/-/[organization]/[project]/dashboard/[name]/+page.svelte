<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";

  const metricViewName: string = $page.params.name;
  const stateSyncManager = new StateSyncManager(metricViewName);

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

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

<div class="p-2">
  <Dashboard {metricViewName} />
</div>
