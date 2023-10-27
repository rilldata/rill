<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import {
    createDashboardSaveStateInit,
    createDashboardSaveStateMutation,
    useDashboardSaveState,
  } from "@rilldata/web-admin/features/dashboards/dashboard-bookmark";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { get } from "svelte/store";

  export let dashboardName: string;
  export let orgName: string;
  export let projectName: string;

  const queryClient = useQueryClient();

  $: proj = createAdminServiceGetProject(orgName, projectName);
  $: bookmark = useDashboardSaveState($proj?.data?.project.id, dashboardName);

  const ctx = getStateManagers();
  const dashboardStore = ctx.dashboardStore;
  let loadedFromBookmark = false;
  const metaQuery = useMetaQuery(ctx);

  $: bookmarkInit = createDashboardSaveStateInit(
    queryClient,
    $proj?.data?.project?.id,
    dashboardName
  );
  $: if ($proj?.data && !$bookmark.isFetching) {
    if (!$bookmark.data) {
      console.log("Init");
      bookmarkInit($dashboardStore.proto);
    } else if (!loadedFromBookmark) {
      console.log("Sync");
      loadedFromBookmark = true;
      metricsExplorerStore.syncFromUrl(
        dashboardName,
        $bookmark.data.data,
        $metaQuery.data
      );
    }
  }
  $: if ($proj?.data && !$bookmark.isFetching && !$bookmark.data) {
    bookmarkInit($dashboardStore.proto);
  }

  $: bookmarkUpdater = createDashboardSaveStateMutation(
    queryClient,
    $proj?.data?.project.id,
    dashboardName
  );

  function saveState() {
    bookmarkUpdater(get(dashboardStore).proto);
  }
</script>

<button on:click={saveState}>Save</button>
<slot />
