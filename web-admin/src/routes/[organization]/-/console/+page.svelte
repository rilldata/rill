<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListOrganizationProjectsWithHealth,
    V1DeploymentStatus,
    type V1ProjectHealth,
  } from "@rilldata/web-admin/client";
  import HealthSummaryCards from "@rilldata/web-admin/features/projects/admin-console/HealthSummaryCards.svelte";
  import ProjectHealthTable from "@rilldata/web-admin/features/projects/admin-console/ProjectHealthTable.svelte";

  $: organization = $page.params.organization;

  $: healthQuery = createAdminServiceListOrganizationProjectsWithHealth(
    organization,
    { pageSize: 50 },
  );

  $: projects = $healthQuery.data?.projects ?? [];

  function isHealthy(p: V1ProjectHealth): boolean {
    return (
      p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
      (p.parseErrorCount ?? 0) === 0 &&
      (p.reconcileErrorCount ?? 0) === 0
    );
  }

  function hasErrors(p: V1ProjectHealth): boolean {
    return (
      p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED ||
      (p.parseErrorCount ?? 0) > 0 ||
      (p.reconcileErrorCount ?? 0) > 0
    );
  }

  $: totalProjects = projects.length;
  $: healthyCount = projects.filter(isHealthy).length;
  $: errorCount = projects.filter(hasErrors).length;

  $: projectData = projects.map((p) => ({
    name: p.projectName!,
    status:
      p.deploymentStatus ??
      V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
    updatedOn: p.updatedOn,
    parseErrorCount: p.parseErrorCount ?? 0,
    reconcileErrorCount: p.reconcileErrorCount ?? 0,
  }));
</script>

{#if $healthQuery.isLoading}
  <p class="text-fg-secondary text-sm">Loading projects...</p>
{:else if $healthQuery.isError}
  <p class="text-red-500 text-sm">Failed to load projects</p>
{:else}
  <div class="flex flex-col gap-y-6">
    <HealthSummaryCards {totalProjects} {healthyCount} {errorCount} />
    <ProjectHealthTable {organization} projects={projectData} />
  </div>
{/if}
