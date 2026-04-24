<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "../../client";
  import ProjectCard from "./ProjectCard.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { projectWelcomeEnabled } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import {
    type SortDirection,
    TableToolbar,
  } from "@rilldata/web-common/components/table-toolbar";

  let { organization }: { organization: string } = $props();

  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(organization, {
      pageSize: 1000,
    }),
  );

  let searchText = $state("");
  let sortDirection = $state<SortDirection>("newest");

  let resolvedProjects = $derived.by(() => {
    const q = searchText.trim().toLowerCase();
    const projects =
      $projectsQuery.data?.projects?.filter((p) =>
        p.name?.toLowerCase().includes(q),
      ) ?? [];
    return projects.sort((a, b) => {
      const aAfterB = a.createdOn > b.createdOn;
      if (sortDirection === "newest") {
        return aAfterB ? -1 : 1;
      } else {
        return aAfterB ? 1 : -1;
      }
    });
  });
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
  />

  <ol class="flex gap-6 flex-wrap">
    {#each resolvedProjects as proj}
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
