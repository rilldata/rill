<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "../../client";
  import ProjectCard from "./ProjectCard.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { projectWelcomeEnabled } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import {
    type FilterGroup,
    type SortDirection,
    TableToolbar,
  } from "@rilldata/web-common/components/table-toolbar";
  import {
    deploymentStatusFilterMatches,
    getDeploymentStatusFilterGroup,
  } from "@rilldata/web-admin/features/branches/deployment-filter-utils.ts";
  import { createQueries } from "@tanstack/svelte-query";
  import { getAdminServiceGetProjectQueryOptions } from "@rilldata/web-admin/client/index.ts";

  let {
    organization,
    createProjectsPermission,
  }: { organization: string; createProjectsPermission: boolean } = $props();

  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization, {
      pageSize: 1000,
    }),
  );
  let projectDetailsQuery = $derived(
    createQueries({
      queries:
        $projectsQuery.data?.projects.map((p) =>
          getAdminServiceGetProjectQueryOptions(organization, p.name, {}),
        ) ?? [],
      combine: (queries) => {
        return queries.map((q) => q.data);
      },
    }),
  );

  let searchText = $state("");
  let sortDirection = $state<SortDirection>("a-z");
  let statusFilter = $state<string[]>([]);

  let filterGroups = $derived([
    getDeploymentStatusFilterGroup(statusFilter, true),
  ] satisfies FilterGroup[]);

  let resolvedProjects = $derived.by(() => {
    const q = searchText.trim().toLowerCase();
    const projects =
      $projectsQuery.data?.projects?.filter((p) => {
        if (!p.name?.toLowerCase().includes(q)) return false;
        const deployment = $projectDetailsQuery.find(
          (d) => d?.project?.name === p.name,
        )?.deployment;
        return deploymentStatusFilterMatches(statusFilter, deployment);
      }) ?? [];
    return projects.sort((a, b) => {
      switch (sortDirection) {
        case "a-z":
          return (a.name ?? "").localeCompare(b.name ?? "");
        case "z-a":
          return (b.name ?? "").localeCompare(a.name ?? "");
        case "newest":
          return a.createdOn > b.createdOn ? -1 : 1;
        case "oldest":
          return a.createdOn > b.createdOn ? 1 : -1;
      }
    });
  });

  let showNewProject = $derived(
    projectWelcomeEnabled && createProjectsPermission,
  );

  function onFilterChange(key: string, selected: string | string[]) {
    if (key === "status" && Array.isArray(selected)) statusFilter = selected;
  }
</script>

<div class="flex flex-col gap-y-4">
  <TableToolbar
    bind:searchText
    searchDisabled={!$projectsQuery.data?.projects?.length}
    showSort
    bind:sortDirection
    showViewToggle
    {filterGroups}
    onFilterChange={onFilterChange}
    onClearAllFilters={() => {
      statusFilter = [];
      searchText = "";
    }}
  >
    {#if showNewProject}
      <Button type="secondary" href="/{organization}/-/create-project">
        + New project
      </Button>
    {/if}
  </TableToolbar>

  <ol class="flex gap-6 flex-wrap">
    {#each resolvedProjects as proj (proj.name)}
      <li>
        <ProjectCard {organization} project={proj.name} />
      </li>
    {:else}
      <p class="text-fg-secondary text-xs">
        This organization has no projects yet.
      </p>
    {/each}
  </ol>
</div>
