<script lang="ts">
  import { useDashboardUrlSync } from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
  import { onDestroy } from "svelte";
  import type { Unsubscriber } from "svelte/store";
  import { getStateManagers } from "../state-managers/state-managers";

  export let metricViewName: string;

  const ctx = getStateManagers();
  let unsubscribe: Unsubscriber;
  const { dashboardStore, metricsViewName: ctxName } = ctx;

  $: if (metricViewName === $ctxName) {
    // Make sure we use the correct sync instance for the current metrics view
    unsubscribe?.();
    unsubscribe = useDashboardUrlSync(ctx);
  }

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });
</script>

{#if $dashboardStore}
  <slot />
{/if}
