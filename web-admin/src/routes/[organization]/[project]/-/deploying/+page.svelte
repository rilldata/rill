<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { project, runtime, deployingDashboard } = data;

  // Get organization name from params (string), not from data (object)
  // Ensure it's a string to prevent [Object object] in URLs
  // In tests, $page.params might not be immediately available, so we guard against undefined
  $: organizationName =
    typeof $page.params.organization === "string"
      ? $page.params.organization
      : undefined;

  // Make this reactive so it only runs when organizationName is available
  // This prevents race conditions where the query might be created before params are ready
  $: deployingDashboardResp = organizationName
    ? useDeployingDashboards(
        runtime.instanceId,
        organizationName,
        project.name,
        deployingDashboard,
      )
    : null;

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
