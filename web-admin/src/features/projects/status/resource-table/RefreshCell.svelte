<script lang="ts">
  import { DateTime } from "luxon";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let date: string;

  $: dateTime = DateTime.fromISO(date);

  $: formattedDate = dateTime.toLocaleString(DateTime.DATETIME_SHORT);

  $: full = dateTime.toLocaleString(DateTime.DATETIME_FULL);
</script>

{#if dateTime.isValid}
  <Tooltip distance={8} location="top">
    <div class="whitespace-nowrap">
      {formattedDate}
    </div>
    <TooltipContent slot="tooltip-content">
      <span class="text-xs text-gray-50 font-medium">
        {full}
      </span>
    </TooltipContent>
  </Tooltip>
{:else}
  <span class="text-fg-secondary">-</span>
{/if}
