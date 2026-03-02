<script lang="ts">
  import { goto } from "$app/navigation";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import type { PageData } from "./$types";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let data: PageData;
  const { organization, project, deployingDashboard } = data;

  const runtimeClient = useRuntimeClient();

  // Make this reactive so that it fires once params are ready.
  // During a first deploy, runtime might not be available when deployment is still being created in the backend.
  $: deployingDashboardResp = useDeployingDashboards(
    runtimeClient,
    organization.name,
    project.name,
    deployingDashboard,
  );

  $: ({ data: deployingDashboardsData } = $deployingDashboardResp ?? {
    data: null,
  });
  $: ({ redirectPath, dashboardsErrored } = deployingDashboardsData ?? {
    redirectPath: null,
    dashboardsErrored: false,
  });

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
