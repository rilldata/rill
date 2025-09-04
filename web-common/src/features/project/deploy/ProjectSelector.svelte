<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import ProjectSelectorItem from "@rilldata/web-common/features/project/deploy/ProjectSelectorItem.svelte";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import { matchSorter } from "match-sorter";

  export let projects: Project[] = [];
  export let selectedProject: Project | undefined = undefined;
  export let enableSearch = false;

  let searchText = "";

  $: filteredProjects = matchSorter(projects, searchText, { keys: ["name"] });
</script>

<div class="flex flex-col gap-y-2">
  {#if enableSearch}
    <Search
      bind:value={searchText}
      background={false}
      forcedInputStyle="bg-slate-100"
    />
  {/if}
  <div class="flex flex-col gap-y-0.5 w-full">
    {#each filteredProjects as project (project.id)}
      <ProjectSelectorItem
        {project}
        selected={project.id === selectedProject?.id}
        onClick={() => (selectedProject = project)}
      />
    {/each}
  </div>
</div>
