<script lang="ts">
  import { page } from "$app/stores";
  import { useQueryClient } from "@rilldata/svelte-query";
  import { getHomeBookmark } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { createDashboardStateSync } from "@rilldata/web-common/features/dashboards/stores/syncDashboardState";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let metricViewName: string;

  $: initLocalUserPreferenceStore(metricViewName);
  const queryClient = useQueryClient();
  $: bookmarks = getHomeBookmark(
    queryClient,
    $runtime?.instanceId,
    $page.params.organization,
    $page.params.project,
    metricViewName,
  );

  const dashboardStateSync = createDashboardStateSync(
    getStateManagers(),
    bookmarks,
  );
</script>

{#if $dashboardStateSync}
  <slot />
{:else}
  <div class="grid place-items-center mt-40">
    <Spinner status={EntityStatus.Running} size="40px" />
  </div>
{/if}
