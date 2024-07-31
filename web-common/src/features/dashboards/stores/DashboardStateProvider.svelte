<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import { isLoadingWithTimeout } from "../../entity-management/spinner-utils";

  const dashboardStoreReady = createDashboardStateSync(getStateManagers());

  export let metricViewName: string;

  $: initLocalUserPreferenceStore(metricViewName);

  const loadingWithTimeout = isLoadingWithTimeout();
  $: loadingWithTimeout.setLoading($dashboardStoreReady.isFetching);
  $: isFetchingWithTimeout = $loadingWithTimeout;
</script>

{#if isFetchingWithTimeout}
  <div class="grid place-items-center size-full">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{:else}
  <slot />
{/if}
