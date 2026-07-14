<script lang="ts">
  import { goto } from "$app/navigation";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let data: PageData;
  const { organization, project, targetDashboard, preCommitSha } = data;

  const runtimeClient = useRuntimeClient();

  // Make this reactive so that it fires once params are ready.
  // During a first deploy, runtime might not be available when deployment is still being created in the backend.
  $: deployingDashboardResp = useDeployingDashboards(
    runtimeClient,
    organization.name,
    project.name,
    targetDashboard,
    preCommitSha,
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
        message: m.error_deploy_dashboards(),
      });
    }
    redirected = true;
    void goto(redirectPath);
  }
</script>

<DashboardBuilding multipleDashboards />
