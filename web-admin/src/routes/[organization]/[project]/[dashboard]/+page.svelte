<script lang="ts">
  import { page } from "$app/stores";
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { getDashboardsForProject } from "@rilldata/web-admin/components/projects/dashboards";
  import { invalidateProjectQueries } from "@rilldata/web-admin/components/projects/invalidations";
  import { useProject } from "@rilldata/web-admin/components/projects/use-project";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ProjectBuilding from "../../../../components/projects/ProjectBuilding.svelte";
  import ProjectErrored from "../../../../components/projects/ProjectErrored.svelte";

  const queryClient = useQueryClient();

  $: org = $page.params.organization;
  $: proj = $page.params.project;
  $: dash = $page.params.dashboard;
  // Poll for project status
  $: project = useProject(org, proj);

  let isProjectBuilding: boolean;
  let isProjectErrored: boolean;
  let isProjectOK: boolean;

  $: if ($project.data?.prodDeployment?.status) {
    const projectWasNotOk = !isProjectOK;

    isProjectBuilding =
      $project.data?.prodDeployment?.status ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      $project.data?.prodDeployment?.status ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING;
    isProjectErrored =
      $project.data?.prodDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR;
    isProjectOK =
      $project.data?.prodDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_OK;

    if (projectWasNotOk && isProjectOK) {
      getDashboardsAndInvalidate();
    }
  }

  async function getDashboardsAndInvalidate() {
    return invalidateProjectQueries(
      queryClient,
      await getDashboardsForProject($project.data)
    );
  }

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
  <title>{dash} - Rill</title>
</svelte:head>

{#if isProjectBuilding}
  <ProjectBuilding organization={org} project={proj} />
{:else if isProjectErrored}
  <ProjectErrored organization={org} project={proj} />
{:else if isProjectOK}
  <Dashboard leftMargin={"48px"} hasTitle={false} metricViewName={dash} />
{/if}
