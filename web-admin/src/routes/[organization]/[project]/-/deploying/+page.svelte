<script lang="ts">
  import { goto } from "$app/navigation";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { organization, project, runtime, deployingDashboard } = data;

  const deployingDashboardResp = useDeployingDashboards(
    runtime.instanceId,
    organization,
    project.name,
    deployingDashboard,
  );

  $: ({ data: deployingDashboardsData } = $deployingDashboardResp);
  $: ({ redirectPath, dashboardsErrored } = deployingDashboardsData ?? {
    redirectPath: null,
    dashboardsErrored: false,
  });

  $: console.log(
    "deployingDashboardsData",
    organization,
    project,
    runtime,
    deployingDashboard,
    redirectPath,
    dashboardsErrored,
  );

  let redirected = false;
  $: if (redirectPath && !redirected) {
    if (dashboardsErrored) {
      eventBus.emit("notification", {
        type: "error",
        message: "Failed to deploy dashboards",
      });
    }
    redirected = true;
    void goto(redirectPath);
  }
</script>

<DashboardBuilding multipleDashboards />
