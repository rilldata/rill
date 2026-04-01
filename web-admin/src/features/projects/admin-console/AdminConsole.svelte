<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceListProjectsForOrganization,
    V1DeploymentStatus,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { derived, type Readable } from "svelte/store";
  import HealthSummaryCards from "./HealthSummaryCards.svelte";
  import ProjectHealthTable from "./ProjectHealthTable.svelte";

  $: organization = $page.params.organization;

  $: projectsQuery = createAdminServiceListProjectsForOrganization(
    organization,
    { pageSize: 50 },
  );

  $: projects = $projectsQuery.data?.projects ?? [];

  // Create a derived store that combines all per-project deployment queries
  // into a single array of { name, status, updatedOn }
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

    // TODO: Replace with createAdminServiceListOrganizationProjectsWithHealth once generated client is available
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

<ContentContainer title="Admin Console" maxWidth={1100}>
  {#if $projectsQuery.isLoading}
    <p class="text-gray-500 text-sm">Loading projects...</p>
  {:else if $projectsQuery.isError}
    <p class="text-red-500 text-sm">Failed to load projects</p>
  {:else}
    <div class="flex flex-col gap-y-6 pt-4">
      <HealthSummaryCards {totalProjects} {healthyCount} {errorCount} />
      <ProjectHealthTable {organization} projects={projectData} />

      <!-- Analytics canvas dashboard placeholder -->
      <div
        class="border border-dashed border-gray-300 rounded-lg p-8 flex items-center justify-center min-h-[400px]"
      >
        <p class="text-gray-400 text-sm">Analytics dashboard coming soon</p>
      </div>
    </div>
  {/if}
</ContentContainer>
