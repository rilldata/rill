<script lang="ts">
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;
  export let description: string;
  export let isPublic: boolean;
  export let updatedOn: string | undefined;

  $: updatedDate = updatedOn ? new Date(updatedOn) : null;
</script>

<a
  href={`/${organization}/${project}`}
  class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <Tooltip distance={8}>
      <span class="text-fg-secondary shrink-0 flex">
        {#if isPublic}
          <Globe size="14px" />
        {:else}
          <Lock size="14px" />
        {/if}
      </span>
      <TooltipContent slot="tooltip-content">
        <span class="text-xs">
          This project is
          <span class="font-medium">{isPublic ? "public" : "private"}</span>
        </span>
      </TooltipContent>
    </Tooltip>
    <span
      class="text-fg-primary text-sm font-semibold group-hover:text-accent-primary-action truncate"
    >
      {project}
    </span>
    <Tag>
      <ProjectAccessControls {organization} {project}>
        <svelte:fragment slot="read-project">Viewer</svelte:fragment>
        <svelte:fragment slot="manage-project">Admin</svelte:fragment>
      </ProjectAccessControls>
    </Tag>
  </div>
  <div
    class="flex gap-x-1 text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden"
  >
    {#if updatedDate}
      <Tooltip distance={8}>
        <span class="shrink-0">Updated {timeAgo(updatedDate)}</span>
        <TooltipContent slot="tooltip-content">
          {updatedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {/if}
    {#if description}
      {#if updatedDate}
        <span class="shrink-0">•</span>
      {/if}
      <span class="truncate">{description}</span>
    {/if}
  </div>
</a>
