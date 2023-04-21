<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import {
    getProjectStatusStore,
    ProjectStatusStore,
  } from "@rilldata/web-admin/components/projects/project-status-store";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { StateSyncManager } from "@rilldata/web-common/features/dashboards/proto-state/StateSyncManager";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ProjectBuilding from "../../../../components/projects/ProjectBuilding.svelte";
  import ProjectErrored from "../../../../components/projects/ProjectErrored.svelte";

  $: org = $page.params.organization;
  $: proj = $page.params.project;
  $: dash = $page.params.dashboard;

  const queryClient = useQueryClient();

  $: projectStatusQuery = createAdminServiceGetProject(org, proj);
  let projectStatusStore: ProjectStatusStore;
  $: projectStatusStore = getProjectStatusStore(
    org,
    proj,
    queryClient,
    projectStatusQuery
  );

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

{#if $projectStatusStore.pending || $projectStatusStore.reconciling}
  <ProjectBuilding organization={org} project={proj} />
{:else if $projectStatusStore.errored}
  <ProjectErrored organization={org} project={proj} />
{:else if $projectStatusStore.ok}
  <Dashboard leftMargin={"48px"} hasTitle={false} metricViewName={dash} />
{/if}
