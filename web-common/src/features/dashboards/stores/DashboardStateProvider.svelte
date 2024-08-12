<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";

  export let metricViewName: string;

  const dashboardStoreReady = createDashboardStateSync(getStateManagers());

  $: initLocalUserPreferenceStore(metricViewName);
</script>

{#if $dashboardStoreReady.isFetching}
  <div class="grid place-items-center size-full">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{:else}
  <slot />
{/if}
