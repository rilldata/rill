<script lang="ts">
  import { page } from "$app/stores";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";

  export let metricViewName: string;

  $: metricsViewQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: metricsExplorer = useDashboardStore(metricViewName);
  $: stateSyncManager = new StateSyncManager(metricViewName);
  $: if ($metricsExplorer) {
    stateSyncManager.handleStateChange($metricsExplorer);
  }
  $: if ($page && $metricsViewQuery.data) {
    stateSyncManager.handleUrlChange(
      get(metricsExplorer),
      $metricsViewQuery.data
    );
  }
</script>

{#if $metricsViewQuery.data}
  <slot />
{/if}
