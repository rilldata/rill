<script lang="ts">
  import { page } from "$app/stores";
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import {
    DashboardListItem,
    getDashboardListItemsFromFilesAndCatalogEntries,
    getDashboardsForProject,
  } from "@rilldata/web-admin/components/projects/dashboards";
  import { invalidateDashboardsQueries } from "@rilldata/web-admin/components/projects/invalidations";
  import { useProject } from "@rilldata/web-admin/components/projects/use-project";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import {
    createRuntimeServiceListCatalogEntries,
    createRuntimeServiceListFiles,
    getRuntimeServiceListCatalogEntriesQueryKey,
    getRuntimeServiceListFilesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ProjectBuilding from "../../../../components/projects/ProjectBuilding.svelte";
  import ProjectErrored from "../../../../components/projects/ProjectErrored.svelte";

  const queryClient = useQueryClient();

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;
  $: dashboardName = $page.params.dashboard;

  // Poll for project status
  $: project = useProject(orgName, projectName);

  let isProjectBuilding: boolean;
  let isProjectOK: boolean;

  $: if ($project.data?.prodDeployment?.status) {
    const projectWasNotOk = !isProjectOK;

    isProjectBuilding =
      $project.data?.prodDeployment?.status ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
      $project.data?.prodDeployment?.status ===
        V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING;
    isProjectOK =
      $project.data?.prodDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_OK;

    if (projectWasNotOk && isProjectOK) {
      getDashboardsAndInvalidate();

      // Invalidate the queries used to assess dashboard validity
      queryClient.invalidateQueries(
        getRuntimeServiceListFilesQueryKey($runtime?.instanceId, {
          glob: "dashboards/*.yaml",
        })
      );
      queryClient.invalidateQueries(
        getRuntimeServiceListCatalogEntriesQueryKey($runtime?.instanceId, {
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

  // Here we check to see if a dashboard is valid by looking at `DashboardListItem.isValid`
  // As is the case in `Breadcrumbs.svelte`, there are two queries we compose to get the dashboard list items,
  // and we should hide this complexity in a custom hook.
  // We avoid calling `GetCatalogEntry` to check for dashboard validity because that would trigger a 404 page.
  $: dashboardFiles = createRuntimeServiceListFiles(
    $runtime?.instanceId,
    {
      glob: "dashboards/*.yaml",
    },
    {
      query: {
        placeholderData: undefined,
        enabled: !!project && !!$runtime?.instanceId,
      },
    }
  );
  $: dashboardCatalogEntries = createRuntimeServiceListCatalogEntries(
    $runtime?.instanceId,
    {
      type: "OBJECT_TYPE_METRICS_VIEW",
    },
    {
      query: {
        placeholderData: undefined,
        enabled: !!project && !!$runtime?.instanceId,
      },
    }
  );
  let dashboardListItems: DashboardListItem[];
  let currentDashboard: DashboardListItem;
  $: if ($dashboardFiles.isSuccess && $dashboardCatalogEntries.isSuccess) {
    dashboardListItems = getDashboardListItemsFromFilesAndCatalogEntries(
      $dashboardFiles.data?.paths,
      $dashboardCatalogEntries.data?.entries
    );

    currentDashboard = dashboardListItems?.find(
      (listing) => listing.name === $page.params.dashboard
    );
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
