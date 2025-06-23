<script lang="ts">
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import RillFilled from "@rilldata/web-common/components/icons/RillFilled.svelte";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";

  export let project: Project;
  export let selected = false;
  export let onClick: () => void = () => {};

  let hovered = false;
  $: isManaged = !project.gitRemote || !!project.managedGitId;
</script>

<button
  class="flex flex-row items-center justify-between w-full text-xs text-gray-900 text-left font-medium p-1 pl-2"
  class:hover:bg-slate-50={!selected}
  class:bg-primary-100={selected}
  on:click={onClick}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
>
  <div class="flex flex-row items-center gap-x-2 w-full">
    {#if isManaged}
      <Tooltip.Root portal="body">
        <Tooltip.Trigger>
          <RillFilled size="14" />
        </Tooltip.Trigger>
        <Tooltip.Content side="bottom">Rill-managed</Tooltip.Content>
      </Tooltip.Root>
    {:else}
      <Github size="14" />
    {/if}
    <span class="w-full">{project.orgName}/{project.name}</span>
  </div>
  {#if hovered}
    <Tooltip.Root portal="body">
      <Tooltip.Trigger>
        <a
          target="_blank"
          rel="noopener noreferrer"
          href={project.frontendUrl}
          class="justify-end"
        >
          <ExternalLink className="fill-gray-700" />
        </a>
      </Tooltip.Trigger>
      <Tooltip.Content side="bottom">Open Rill Cloud project</Tooltip.Content>
    </Tooltip.Root>
  {/if}
</button>
