<script lang="ts">
  import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { readable } from "svelte/store";

  export let exploreName: string;
  export let token: V1MagicAuthToken;

  $: initLocalUserPreferenceStore(exploreName);

  $: dashboardStoreReady = createDashboardStateSync(
    getStateManagers(),
    readable({
      isFetching: false,
      data: token.state,
      error: null,
    }),
  );
</script>

{#if $dashboardStoreReady}
  <slot />
{:else}
  <div class="grid place-items-center mt-40">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{/if}
