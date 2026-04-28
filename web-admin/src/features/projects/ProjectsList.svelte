<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { projectWelcomeEnabled } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { LayoutGrid, List } from "lucide-svelte";
  import ProjectCards from "./ProjectCards.svelte";
  import ProjectsTable from "./ProjectsTable.svelte";

  export let organization: string;

  type View = "grid" | "list";
  let view: View = "grid";

  $: projs = createAdminServiceListProjectsForOrganization(organization, {
    pageSize: 1000,
  });
  $: projects = $projs.data?.projects ?? [];
</script>

<div class="flex flex-col gap-y-4">
  <div
    class="flex flex-row items-center text-fg-secondary text-base font-normal leading-normal"
  >
    <span class="grow">Check out your projects below.</span>
    <div class="flex items-center gap-x-2">
      <div
        class="flex items-center rounded-md border border-gray-200 dark:border-gray-700 p-0.5"
        role="group"
        aria-label="View toggle"
      >
        <Tooltip distance={8}>
          <button
            type="button"
            class="view-toggle-btn"
            class:active={view === "grid"}
            aria-pressed={view === "grid"}
            aria-label="Grid view"
            on:click={() => (view = "grid")}
          >
            <LayoutGrid size="14" />
          </button>
          <TooltipContent slot="tooltip-content">Grid view</TooltipContent>
        </Tooltip>
        <Tooltip distance={8}>
          <button
            type="button"
            class="view-toggle-btn"
            class:active={view === "list"}
            aria-pressed={view === "list"}
            aria-label="List view"
            on:click={() => (view = "list")}
          >
            <List size="14" />
          </button>
          <TooltipContent slot="tooltip-content">List view</TooltipContent>
        </Tooltip>
      </div>
      {#if projectWelcomeEnabled}
        <Button type="primary" href="/{organization}/-/create-project">
          Create new
        </Button>
      {/if}
    </div>
  </div>

  {#if projects.length === 0}
    <p class="text-fg-secondary text-xs">
      This organization has no projects yet.
    </p>
  {:else if view === "grid"}
    <ProjectCards {organization} {projects} />
  {:else}
    <ProjectsTable {organization} {projects} />
  {/if}
</div>

<style lang="postcss">
  .view-toggle-btn {
    @apply flex items-center justify-center px-2 py-1 rounded text-fg-secondary;
    @apply transition-colors;
  }

  .view-toggle-btn:hover {
    @apply bg-surface-hover text-fg-primary;
  }

  .view-toggle-btn.active {
    @apply bg-surface-active text-fg-primary;
  }
</style>
