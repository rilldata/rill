<script lang="ts">
  import { useDashboardUrlSync } from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy } from "svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import { metricsExplorerStore } from "../dashboard-stores";

  export let metricViewName: string;

  $: metricsViewQuery = useMetaQuery($runtime.instanceId, metricViewName);
  let unsubscribe;
  $: if ($metricsViewQuery?.data) {
    // unsubscribe any previous subscription. this can happen when metricViewName changes and hence the metricsViewQuery
    if (unsubscribe) unsubscribe();
    unsubscribe = useDashboardUrlSync(metricViewName, metricsViewQuery);
  }

  $: if ($metricsViewQuery.data) {
    metricsExplorerStore.sync(metricViewName, $metricsViewQuery.data);
  }

  const { dashboardStore } = getStateManagers();

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });
</script>

{#if $dashboardStore}
  <slot />
{/if}
