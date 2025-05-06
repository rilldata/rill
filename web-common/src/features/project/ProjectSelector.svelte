<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import RillFilled from "@rilldata/web-common/components/icons/RillFilled.svelte";
  import { matchSorter } from "match-sorter";

  export let projects: Project[] = [];

  let searchText = "";

  $: filteredProjects = matchSorter(projects, searchText, { keys: ["name"] });

  function isRillManaged(project: Project) {
    return !project.githubUrl;
  }
</script>

<div class="flex flex-col gap-y-2 w-[500px]">
  <Search bind:value={searchText} />
  {#each filteredProjects as project (project.id)}
    <button
      class="text-xs text-gray-900 text-left flex flex-row items-center gap-x-2 w-full"
    >
      {#if isRillManaged(project)}
        <RillFilled size="12" />
      {:else}
        <Github size="12" />
      {/if}
      {project.orgName}/{project.name}
    </button>
  {/each}
</div>
