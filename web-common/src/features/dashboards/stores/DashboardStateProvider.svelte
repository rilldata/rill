<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";

  const dashboardStoreReady = createDashboardStateSync(getStateManagers());

  export let exploreName: string;

  $: initLocalUserPreferenceStore(exploreName);
</script>

{#if $dashboardStoreReady.isFetching}
  <div class="grid place-items-center size-full">
    <DelayedSpinner isLoading={$dashboardStoreReady.isFetching} size="40px" />
  </div>
{:else}
  <slot />
{/if}
