<!-- 
  This page is used to redirect to either the project's first dashboard, or to the project's status page.
 -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useDashboardNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createAdminServiceGetProject } from "../../../../../client";

  $: proj = createAdminServiceGetProject(
    $page.params.organization,
    $page.params.project
  );

  // Avoid a race condition: make sure the runtime store has been updated (with the host, instanceID, and jwt).
  $: isRuntimeStoreReady =
    $proj?.data &&
    $proj.data.prodDeployment.runtimeInstanceId === $runtime.instanceId;

  let dashboardsQuery;
  $: if (isRuntimeStoreReady) {
    dashboardsQuery = useDashboardNames($runtime.instanceId);
  }
  $: if ($dashboardsQuery?.data) {
    if ($dashboardsQuery.data.length === 0) {
      goto(`/${$page.params.organization}/${$page.params.project}`);
    } else {
      goto(
        `/${$page.params.organization}/${$page.params.project}/${$dashboardsQuery.data[0]}`
      );
    }
  }
</script>
