<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { getDashboardsForProject } from "@rilldata/web-admin/components/projects/dashboards";
  import { invalidateDashboardsQueries } from "@rilldata/web-admin/components/projects/invalidations";
  import { useProjectDeploymentStatus } from "@rilldata/web-admin/components/projects/selectors";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/DashboardStateProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import {
    createRuntimeServiceGetCatalogEntry,
    getRuntimeServiceGetCatalogEntryQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import type { QueryError } from "@rilldata/web-common/runtime-client/error";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { errorStore } from "../../../../components/errors/error-store";
  import ProjectBuilding from "../../../../components/projects/ProjectBuilding.svelte";

  const queryClient = useQueryClient();

  $: instanceId = $runtime?.instanceId;

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;
  $: dashboardName = $page.params.dashboard;

  $: project = createAdminServiceGetProject(orgName, projectName);

  $: projectDeploymentStatus = useProjectDeploymentStatus(orgName, projectName); // polls
  $: isProjectPending =
    $projectDeploymentStatus.data ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING;
  $: isProjectReconciling =
    $projectDeploymentStatus.data ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING;
  $: isProjectErrored =
    $projectDeploymentStatus.data ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR;
  $: isProjectBuilding = isProjectPending || isProjectReconciling;
  $: isProjectBuilt = isProjectOK || isProjectErrored;

  let isProjectOK: boolean;

  $: if ($projectDeploymentStatus.data) {
    const projectWasNotOk = !isProjectOK;

    isProjectOK =
      $projectDeploymentStatus.data === V1DeploymentStatus.DEPLOYMENT_STATUS_OK;

    if (projectWasNotOk && isProjectOK) {
      getDashboardsAndInvalidate();

      // Invalidate the query used to assess dashboard validity
      queryClient.invalidateQueries(
        getRuntimeServiceGetCatalogEntryQueryKey(instanceId, dashboardName)
      );
    }
  }

  async function getDashboardsAndInvalidate() {
    const dashboardListings = await getDashboardsForProject($project.data);
    const dashboardNames = dashboardListings.map((listing) => listing.name);
    return invalidateDashboardsQueries(queryClient, dashboardNames);
  }

  $: dashboard = createRuntimeServiceGetCatalogEntry(instanceId, dashboardName);
  $: isDashboardOK = $dashboard.isSuccess;
  $: isDashboardNotFound =
    $dashboard.isError &&
    ($dashboard.error as QueryError)?.response?.status === 404;
  // isDashboardErrored // We'll reinstate this case once we integrate the new Reconcile

  // If no dashboard is found, show a 404 page
  $: if (isProjectBuilt && isDashboardNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Dashboard not found",
      body: `The dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }
</script>

<svelte:head>
  <title>{dashboardName} - Rill</title>
</svelte:head>

<!-- Note: Project and dashboard states might appear to diverge. A project could be errored 
  because dashboard #1 is errored, but dashboard #2 could be OK.  -->

{#if isProjectBuilding && isDashboardNotFound}
  <ProjectBuilding organization={orgName} project={projectName} />
{:else if isDashboardOK}
  <StateManagersProvider metricsViewName={dashboardName}>
    {#key dashboardName}
      <DashboardStateProvider metricViewName={dashboardName}>
        <DashboardURLStateProvider metricViewName={dashboardName}>
          <Dashboard metricViewName={dashboardName} leftMargin={"48px"} />
        </DashboardURLStateProvider>
      </DashboardStateProvider>
    {/key}
  </StateManagersProvider>
{/if}
<!-- We'll reinstate this case once we integrate the new Reconcile -->
<!-- {:else if isDashboardErrored}
  <ProjectErrored organization={orgName} project={projectName} />
{/if} -->
