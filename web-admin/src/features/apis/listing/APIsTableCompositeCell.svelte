<script lang="ts">
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { timeAgo } from "../../dashboards/listing/utils";

  export let id: string;
  export let title: string;
  export let description: string | undefined;
  export let resolver: string | undefined;
  export let reconcileError: string | undefined;
  export let lastUpdated: string | undefined;

  $: lastUpdatedDate = lastUpdated ? new Date(lastUpdated) : null;
</script>

<a
  href={`apis/${id}`}
  class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <APIIcon size="14px" />
    <span
      class="text-fg-primary text-sm font-semibold group-hover:text-accent-primary-action truncate"
    >
      {title}
    </span>
    {#if resolver}
      <span
        class="shrink-0 text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-surface-secondary text-fg-secondary border border-border"
      >
        {resolver}
      </span>
    {/if}
    {#if reconcileError}
      <span
        class="text-red-500 text-xs font-normal shrink-0"
        title={reconcileError}>Error</span
      >
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-fg-secondary text-xs font-normal min-h-[16px] overflow-hidden"
  >
    {#if description}
      <span class="truncate">{description}</span>
      <span class="shrink-0">•</span>
    {/if}
    {#if lastUpdatedDate}
      <Tooltip distance={8}>
        <span class="shrink-0">Updated {timeAgo(lastUpdatedDate)}</span>
        <TooltipContent slot="tooltip-content">
          {lastUpdatedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {:else}
      <span class="shrink-0">Not yet reconciled</span>
    {/if}
  </div>
</a>
