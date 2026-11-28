<script lang="ts">
  import ProjectCard from "./ProjectCard.svelte";
  import ProjectsTagFilter from "./ProjectsTagFilter.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
  import {
    filterProjectsByTags,
    getProjectTags,
  } from "@rilldata/web-admin/features/projects/project-tags";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    organization,
    createProjectsPermission,
  }: { organization: string; createProjectsPermission: boolean } = $props();

  let projectsQuery = $derived(
    createQuery(listProjectsForOrgQueryOptions(organization)),
  );
  let projects = $derived($projectsQuery.data?.projects ?? []);

  // The selected tags are kept in the URL, like the dashboards list's tag filter, so a filtered list can be shared.
  const selectedTagsStore = UrlParamsState.createStringArrayParam("tags");
  let tags = $derived(getProjectTags(projects));
  let filteredProjects = $derived(
    filterProjectsByTags(projects, selectedTagsStore.value),
  );

  let showNewProject = $derived(createProjectsPermission);
</script>

<div class="flex flex-col gap-y-4">
  <span
    class="flex flex-row items-center gap-x-2 text-fg-secondary text-base font-normal leading-normal"
  >
    <span class="grow">{m.org_check_out_projects()}</span>
    {#if tags.length > 0 || selectedTagsStore.value.length > 0}
      <ProjectsTagFilter {tags} {selectedTagsStore} />
    {/if}
    {#if showNewProject}
      <Button type="secondary" href="/{organization}/-/create-project">
        {m.org_new_project()}
      </Button>
    {/if}
  </span>

  <ol class="flex gap-6 flex-wrap">
    {#each filteredProjects as proj (proj.name)}
      <li>
        <ProjectCard {organization} project={proj.name} />
      </li>
    {:else}
      {#if projects.length > 0}
        <p class="text-fg-secondary text-xs">
          {m.org_no_projects_with_tags()}
          <button
            type="button"
            class="text-primary-600 hover:underline ml-1"
            onclick={() => selectedTagsStore.setter([])}
          >
            {m.org_projects_tag_filter_clear()}
          </button>
        </p>
      {:else}
        <p class="text-fg-secondary text-xs">
          {m.org_no_projects_yet()}
        </p>
      {/if}
    {/each}
  </ol>
</div>
