<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import RillFilled from "@rilldata/web-common/components/icons/RillFilled.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { matchSorter } from "match-sorter";

  export let projects: Project[] = [];
  export let selectedProject: Project | undefined = undefined;

  let searchText = "";

  $: filteredProjects = matchSorter(projects, searchText, { keys: ["name"] });

  function isRillManaged(project: Project) {
    return !project.githubUrl;
  }
</script>

<div class="flex flex-col gap-y-2">
  <Search bind:value={searchText} />
  <div class="flex flex-col gap-y-0.5 w-[500px]">
    {#each filteredProjects as project (project.id)}
      {@const selected = project.id === selectedProject?.id}
      <button
        class="flex flex-row items-center gap-x-2 w-full hover:bg-slate-100 text-xs text-gray-900 text-left font-medium p-1 pl-2"
        class:bg-blue-50={selected}
        on:click={() => (selectedProject = project)}
      >
        {#if isRillManaged(project)}
          <Tooltip.Root portal="body">
            <Tooltip.Trigger>
              <RillFilled size="12" />
            </Tooltip.Trigger>
            <Tooltip.Content side="bottom">Rill-managed</Tooltip.Content>
          </Tooltip.Root>
        {:else}
          <Github size="12" />
        {/if}
        {project.orgName}/{project.name}
      </button>
    {/each}
  </div>
</div>
