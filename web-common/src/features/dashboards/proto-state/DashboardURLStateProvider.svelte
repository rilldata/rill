<script lang="ts">
  import { useDashboardUrlSync } from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
  import { onDestroy } from "svelte";
  import { getStateManagers } from "../state-managers/state-managers";

  const ctx = getStateManagers();
  const unsubscribe = useDashboardUrlSync(ctx);
  const { dashboardStore } = ctx;

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });
</script>

{#if $dashboardStore}
  <slot />
{/if}
