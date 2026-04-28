<script lang="ts">
  import type { V1Project } from "@rilldata/web-admin/client";
  import ProjectsTableRow from "./ProjectsTableRow.svelte";

  export let organization: string;
  export let projects: V1Project[];
</script>

<div class="w-full">
  <div class="row row-head">
    <div class="cell text-fg-muted text-sm font-medium">Project name</div>
    <div class="cell text-fg-muted text-sm font-medium">Status</div>
    <div class="cell text-fg-muted text-sm font-medium">Permission</div>
    <div class="cell text-fg-muted text-sm font-medium">Your role</div>
    <div class="cell" aria-hidden="true"></div>
  </div>
  {#each projects as proj (proj.name)}
    <ProjectsTableRow
      {organization}
      project={proj.name ?? ""}
      isPublic={!!proj.public}
    />
  {/each}
</div>

<style lang="postcss">
  .row {
    @apply grid items-center w-full border-b border-border;
    grid-template-columns: minmax(0, 1fr) 200px 200px 200px 60px;
  }

  .row-head {
    @apply h-10;
  }

  .cell {
    @apply px-2 min-w-0;
  }
</style>
