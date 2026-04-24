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

  let { organization }: { organization: string } = $props();

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
  let sortDirection = $state<SortDirection>("newest");
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
          (q) => q?.project?.name === p.name,
        )?.deployment;
        return deploymentStatusFilterMatches(statusFilter, deployment);
      }) ?? [];
    return projects.sort((a, b) => {
      const aAfterB = a.createdOn > b.createdOn;
      if (sortDirection === "newest") {
        return aAfterB ? -1 : 1;
      } else {
        return aAfterB ? 1 : -1;
      }
    });
  });

  function onFilterChange(key: string, selected: string[]) {
    if (key === "status") statusFilter = selected;
  }
</script>

<div class="flex flex-col gap-y-4">
  <div class="flex flex-row items-center justify-between">
    <h2 class="text-2xl font-semibold">Projects</h2>
    {#if projectWelcomeEnabled}
      <Button type="secondary" href="/{organization}/-/create-project">
        + New project
      </Button>
    {/if}
  </div>

  <TableToolbar
    bind:searchText
    searchDisabled={!$projectsQuery.data?.projects?.length}
    showSort
    bind:sortDirection
    {filterGroups}
    {onFilterChange}
    onClearAllFilters={() => {
      statusFilter = [];
      searchText = "";
    }}
  />

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
