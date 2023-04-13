<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "../../../../client";
  import ProjectBuilding from "../../../../components/deployments/ProjectBuilding.svelte";
  import ProjectErrored from "../../../../components/deployments/ProjectErrored.svelte";

  // TODO: add 404 logic as in `web-local`'s `dashboard/[name]/+page.svelte`

  $: org = $page.params.organization;
  $: proj = $page.params.project;
  $: dash = $page.params.dashboard;

  // Poll for project status
  $: project = createAdminServiceGetProject(org, proj, {
    query: {
      refetchInterval: 1000,
    },
  });
  $: isProjectBuilding =
    $project.data?.productionDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    $project.data?.productionDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING;
  $: isProjectErrored =
    $project.data?.productionDeployment?.status ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR;
  $: isProjectOK =
    $project.data?.productionDeployment?.status ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_OK;

  $: metricsExplorer = useDashboardStore(dash);
  const stateSyncManager = new StateSyncManager(dash);
  $: if ($metricsExplorer) {
    stateSyncManager.handleStateChange($metricsExplorer);
  }
  $: if ($page) {
    stateSyncManager.handleUrlChange();
  }
</script>

<svelte:head>
  <title>Rill | {dash}</title>
</svelte:head>

{#if isProjectBuilding}
  <ProjectBuilding organization={org} project={proj} />
{:else if isProjectErrored}
  <ProjectErrored organization={org} project={proj} />
{:else if isProjectOK}
  <Dashboard leftMargin={"44px"} hasTitle={false} metricViewName={dash} />
{/if}
