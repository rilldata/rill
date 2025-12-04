<script lang="ts">
  import { goto } from "$app/navigation";
  import { navigating, page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardBuilding from "@rilldata/web-admin/features/dashboards/DashboardBuilding.svelte";
  import DashboardErrored from "@rilldata/web-admin/features/dashboards/DashboardErrored.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import { useDeployingDashboards } from "@rilldata/web-admin/features/dashboards/listing/deploying-dashboards.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { deploying, deployingDashboard } = data;

  $: ({
    params: { organization, project },
  } = $page);
  $: ({ instanceId } = $runtime);

  $: deployingDashboards = useDeployingDashboards(
    instanceId,
    organization,
    project,
    deploying,
    deployingDashboard,
  );
  $: ({ isPending, data: deployingDashboardsData } = $deployingDashboards);
  $: ({ redirectToDashboardPath, dashboardsReconciling, dashboardsErrored } =
    deployingDashboardsData ?? {
      redirectToDashboardPath: null,
      dashboardsReconciling: false,
      dashboardsErrored: false,
    });

  $: if (redirectToDashboardPath) {
    void goto(redirectToDashboardPath);
  }
  // Continue showing a spinner when redirecting to target dashboard after deploy.
  // This prevents a flash just after dashboard has loaded and before the dashboard components get mounted.
  $: redirecting = $navigating?.to?.url.pathname === redirectToDashboardPath;

  $: showSpinner =
    dashboardsReconciling || (deploying && isPending) || redirecting;
  $: showError = dashboardsErrored && !redirecting;
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

{#if showError}
  <DashboardErrored {organization} {project} />
{:else if showSpinner}
  <DashboardBuilding multipleDashboards />
{:else}
  <ContentContainer maxWidth={800} title="Project dashboards">
    <div class="flex flex-col items-center gap-y-4">
      <DashboardsTable />
    </div>
  </ContentContainer>
{/if}
