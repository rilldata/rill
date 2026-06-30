<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";

  export let value: string | undefined;
  export let fallback: string = "—";

  $: date = value ? new Date(value) : null;
</script>

{#if date}
  <Tooltip distance={8}>
    <span class="truncate">{timeAgo(date)}</span>
    <TooltipContent slot="tooltip-content">
      {date.toLocaleString()}
    </TooltipContent>
  </Tooltip>
{:else}
  <span class="text-fg-tertiary">{fallback}</span>
{/if}
