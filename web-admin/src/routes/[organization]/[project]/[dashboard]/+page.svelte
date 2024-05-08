<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import DashboardBookmarksStateProvider from "@rilldata/web-admin/features/dashboards/DashboardBookmarksStateProvider.svelte";
  import { getDashboardsForProject } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { invalidateDashboardsQueries } from "@rilldata/web-admin/features/projects/invalidations";
  import ProjectErrored from "@rilldata/web-admin/features/projects/ProjectErrored.svelte";
  import { useProjectDeploymentStatus } from "@rilldata/web-admin/features/projects/status/selectors";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardThemeProvider from "@rilldata/web-common/features/dashboards/DashboardThemeProvider.svelte";
  import DashboardURLStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardURLStateProvider.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/stores/DashboardStateProvider.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getRuntimeServiceGetResourceQueryKey } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { errorStore } from "../../../../features/errors/error-store";
  import ProjectBuilding from "../../../../features/projects/ProjectBuilding.svelte";

  const queryClient = useQueryClient();

  $: instanceId = $runtime?.instanceId;

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;
  $: dashboardName = $page.params.dashboard;

  const user = createAdminServiceGetCurrentUser();

  $: project = createAdminServiceGetProject(orgName, projectName);

  $: projectDeploymentStatus = useProjectDeploymentStatus(orgName, projectName); // polls
  $: isProjectPending =
    $projectDeploymentStatus.data ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING;
  $: isProjectErrored =
    $projectDeploymentStatus.data ===
    V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR;
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
        getRuntimeServiceGetResourceQueryKey(instanceId, {
          "name.name": dashboardName,
          "name.kind": ResourceKind.MetricsView,
        }),
      );
    }
  }

  async function getDashboardsAndInvalidate() {
    const dashboardListings = await getDashboardsForProject($project.data);
    const dashboardNames = dashboardListings.map(
      (listing) => listing.meta.name.name,
    );
    return invalidateDashboardsQueries(queryClient, dashboardNames);
  }

  $: dashboard = useDashboard(instanceId, dashboardName);
  $: isDashboardNotFound =
    !$dashboard.data &&
    $dashboard.isError &&
    $dashboard.error?.response?.status === 404;
  // We check for metricsView.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid dashboards, allowing display even when the current dashboard spec is invalid
  // and a meta.reconcileError exists.
  $: isDashboardErrored = !$dashboard.data?.metricsView?.state?.validSpec;

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

{#if isProjectPending && isDashboardNotFound}
  <ProjectBuilding organization={orgName} project={projectName} />
{:else if $dashboard.isSuccess}
  {#if isDashboardErrored}
    <ProjectErrored organization={orgName} project={projectName} />
  {:else}
    {#key dashboardName}
      <StateManagersProvider metricsViewName={dashboardName}>
        {#if $user.isSuccess && $user.data.user}
          <DashboardBookmarksStateProvider metricViewName={dashboardName}>
            <DashboardURLStateProvider metricViewName={dashboardName}>
              <DashboardThemeProvider>
                <Dashboard metricViewName={dashboardName} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardBookmarksStateProvider>
        {:else}
          <DashboardStateProvider metricViewName={dashboardName}>
            <DashboardURLStateProvider metricViewName={dashboardName}>
              <DashboardThemeProvider>
                <Dashboard metricViewName={dashboardName} />
              </DashboardThemeProvider>
            </DashboardURLStateProvider>
          </DashboardStateProvider>
        {/if}
      </StateManagersProvider>
    {/key}
  {/if}
{/if}
