<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceListProjectsForOrganization,
    V1DeploymentStatus,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import { derived, type Readable } from "svelte/store";
  import HealthSummaryCards from "@rilldata/web-admin/features/projects/admin-console/HealthSummaryCards.svelte";
  import ProjectHealthTable from "@rilldata/web-admin/features/projects/admin-console/ProjectHealthTable.svelte";

  $: organization = $page.params.organization;

  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    { pageSize: 50 },
  );

  $: projects = $projectsQuery.data?.projects ?? [];

  // TODO: Replace with createAdminServiceListOrganizationProjectsWithHealth
  // once the generated client is available. That endpoint provides parseErrorCount
  // and reconcileErrorCount per project, enabling accurate health determination:
  // a project is "healthy" only if deployment is RUNNING and error counts are zero.
  $: projectDataStore = deriveProjectData(organization, projects);

  function deriveProjectData(
    org: string,
    projs: V1Project[],
  ): Readable<
    Array<{
      name: string;
      status: V1DeploymentStatus;
      updatedOn: string | undefined;
    }>
  > {
    if (projs.length === 0) {
      return derived([], () => []);
    }

    const queries = projs.map((p) =>
      createAdminServiceGetProject(org, p.name!),
    );

    return derived(queries, (results) =>
      results.map((result, i) => ({
        name: projs[i].name!,
        status:
          result.data?.deployment?.status ??
          V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
        updatedOn: projs[i].updatedOn,
      })),
    );
  }

  $: projectData = $projectDataStore;

  $: totalProjects = projectData.length;
  $: healthyCount = projectData.filter(
    (p) => p.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING,
  ).length;
  $: errorCount = projectData.filter(
    (p) => p.status === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED,
  ).length;
</script>

{#if $projectsQuery.isLoading}
  <p class="text-fg-secondary text-sm">Loading projects...</p>
{:else if $projectsQuery.isError}
  <p class="text-red-500 text-sm">Failed to load projects</p>
{:else}
  <div class="flex flex-col gap-y-6">
    <HealthSummaryCards {totalProjects} {healthyCount} {errorCount} />
    <ProjectHealthTable {organization} projects={projectData} />
  </div>
{/if}
