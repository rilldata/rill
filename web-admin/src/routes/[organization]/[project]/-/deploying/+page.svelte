<script lang="ts">
  import { goto } from "$app/navigation";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";

  // If the prod runtime hasn't progressed past `preCommitSha` within this
  // window, redirect anyway and surface a warning. Covers webhook delivery
  // delays and the rare no-op-merge case where the SHA never changes.
  const FALLBACK_REDIRECT_MS = 30_000;

  export let data: PageData;
  const { organization, project, deployingDashboard, preCommitSha } = data;

  const runtimeClient = useRuntimeClient();

  // Make this reactive so that it fires once params are ready.
  // During a first deploy, runtime might not be available when deployment is still being created in the backend.
  $: deployingDashboardResp = useDeployingDashboards(
    runtimeClient,
    organization.name,
    project.name,
    deployingDashboard,
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
        message: "Failed to deploy dashboards",
      });
    }
    redirected = true;
    void goto(redirectPath);
  }

  onMount(() => {
    const fallbackUrl = `/${organization.name}/${project.name}`;
    const timeoutId = window.setTimeout(() => {
      if (redirected) return;
      eventBus.emit("notification", {
        type: "default",
        message:
          "Changes may take a moment to appear; refresh the page if needed.",
      });
      redirected = true;
      void goto(fallbackUrl);
    }, FALLBACK_REDIRECT_MS);
    return () => window.clearTimeout(timeoutId);
  });
</script>

<DashboardBuilding multipleDashboards />
