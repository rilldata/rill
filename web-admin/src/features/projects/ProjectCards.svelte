<script lang="ts">
  import ProjectCard from "./ProjectCard.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
  import { createQuery } from "@tanstack/svelte-query";

  let {
    organization,
    createProjectsPermission,
  }: { organization: string; createProjectsPermission: boolean } = $props();

  let projectsQuery = $derived(
    createQuery(listProjectsForOrgQueryOptions(organization)),
  );
  let projects = $derived($projectsQuery.data?.projects ?? []);

  let showNewProject = $derived(createProjectsPermission);
</script>

<div class="flex flex-col gap-y-4">
  <span
    class="flex flex-row items-center text-fg-secondary text-base font-normal leading-normal"
  >
    <span class="grow">Check out your projects below.</span>
    {#if showNewProject}
      <Button type="secondary" href="/{organization}/-/create-project">
        + New project
      </Button>
    {/if}
  </span>

  <ol class="flex gap-6 flex-wrap">
    {#each projects as proj (proj.name)}
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
