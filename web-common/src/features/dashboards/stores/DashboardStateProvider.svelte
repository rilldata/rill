<script lang="ts">
  import { page } from "$app/stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { writable } from "svelte/store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";

  export let metricViewName: string;

  $: initLocalUserPreferenceStore(metricViewName);

  const dashboardStateSync = createDashboardStateSync(
    getStateManagers(),
    writable({
      isFetching: false,
      error: "",
      data: $page.url.searchParams.get("state") ?? "",
    }),
  );
</script>

{#if $dashboardStateSync}
  <slot />
{:else}
  <div class="grid place-items-center mt-40">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{/if}
