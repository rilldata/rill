<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    DashboardListItem,
    getDashboardsForProject,
    useDashboardListItems,
  } from "@rilldata/web-admin/components/projects/dashboards";
  import { invalidateDashboardsQueries } from "@rilldata/web-admin/components/projects/invalidations";
  import { useProjectDeploymentStatus } from "@rilldata/web-admin/components/projects/selectors";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import {
    getRuntimeServiceListCatalogEntriesQueryKey,
    getRuntimeServiceListFilesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { errorStore } from "../../../../components/errors/error-store";
  import ProjectBuilding from "../../../../components/projects/ProjectBuilding.svelte";
  import ProjectErrored from "../../../../components/projects/ProjectErrored.svelte";

  const queryClient = useQueryClient();

  $: instanceId = $runtime?.instanceId;

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;
  $: dashboardName = $page.params.dashboard;

  $: project = createAdminServiceGetProject(orgName, projectName);
  // Poll specifically for the project's deployment status
  $: projectDeploymentStatus = useProjectDeploymentStatus(orgName, projectName); // polls

  let isProjectBuilding: boolean;
  let isProjectOK: boolean;

  $: if ($projectDeploymentStatus.data) {
    const projectWasNotOk = !isProjectOK;

    isProjectBuilding =
      $projectDeploymentStatus.data ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      $projectDeploymentStatus.data ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING;
    isProjectOK =
      $projectDeploymentStatus.data === V1DeploymentStatus.DEPLOYMENT_STATUS_OK;

    if (projectWasNotOk && isProjectOK) {
      getDashboardsAndInvalidate();

      // Invalidate the queries used to assess dashboard validity
      queryClient.invalidateQueries(
        getRuntimeServiceListFilesQueryKey(instanceId, {
          glob: "dashboards/*.yaml",
        })
      );
      queryClient.invalidateQueries(
        getRuntimeServiceListCatalogEntriesQueryKey(instanceId, {
          type: "OBJECT_TYPE_METRICS_VIEW",
        })
      );
    }
  }

  async function getDashboardsAndInvalidate() {
    const dashboardListings = await getDashboardsForProject($project.data);
    const dashboardNames = dashboardListings.map((listing) => listing.name);
    return invalidateDashboardsQueries(queryClient, dashboardNames);
  }

  // We avoid calling `GetCatalogEntry` to check for dashboard validity because that would trigger a 404 page.
  $: dashboardListItems = useDashboardListItems(
    instanceId,
    $projectDeploymentStatus.data
  );
  let currentDashboard: DashboardListItem;
  $: if ($dashboardListItems.isSuccess) {
    currentDashboard = $dashboardListItems?.items?.find(
      (listing) => listing.name === $page.params.dashboard
    );

    // If no dashboard is found, show a 404 page
    if (!currentDashboard) {
      errorStore.set({
        statusCode: 404,
        header: "Dashboard not found",
        body: `The dashboard you requested could not be found. Please check that you have provided a valid dashboard name.`,
      });
    }
  }
</script>

<svelte:head>
  <title>{dashboardName} - Rill</title>
</svelte:head>

{#if isProjectBuilding}
  <ProjectBuilding organization={orgName} project={projectName} />
{:else if !currentDashboard}
  <!-- show nothing -->
{:else if currentDashboard && !currentDashboard.isValid}
  <ProjectErrored organization={orgName} project={projectName} />
{:else}
  <Dashboard
    leftMargin={"48px"}
    hasTitle={false}
    metricViewName={dashboardName}
  />
{/if}
