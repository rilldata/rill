<script lang="ts">
  import { DateTime } from "luxon";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let date: string;

  $: dateTime = DateTime.fromISO(date);

  // Use compact format: "Jan 15, 2:30 PM" - shorter than DATETIME_SHORT
  $: formattedDate = dateTime.toLocaleString({
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });

  $: full = dateTime.toLocaleString(DateTime.DATETIME_FULL);
</script>

{#if dateTime.isValid}
  <Tooltip distance={8} location="top">
    <div class="whitespace-nowrap">
      {formattedDate}
    </div>
    <TooltipContent slot="tooltip-content">
      <span class="text-xs font-medium">
        {full}
      </span>
    </TooltipContent>
  </Tooltip>
{:else}
  <span class="text-fg-secondary">-</span>
{/if}
