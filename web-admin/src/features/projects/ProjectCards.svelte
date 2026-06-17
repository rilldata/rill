<script lang="ts">
  import ProjectCard from "./ProjectCard.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
    <span class="grow">{m.projects_cards_subtitle()}</span>
    {#if showNewProject}
      <Button type="secondary" href="/{organization}/-/create-project">
        {m.projects_cards_new_project()}
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
        {m.projects_cards_empty()}
      </p>
    {/each}
  </ol>
</div>
