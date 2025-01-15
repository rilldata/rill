<script lang="ts">
  import { DateTime } from "luxon";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let date: string;

  let clientWidth: number;

  $: format = getFormat(clientWidth);

  $: dateTime = DateTime.fromISO(date);

  $: formattedDate = dateTime.toLocaleString(format);

  $: full = dateTime.toLocaleString(DateTime.DATETIME_FULL);

  function getFormat(width: number) {
    switch (true) {
      case width < 120:
        return DateTime.DATE_SHORT;
      case width < 180:
        return DateTime.DATETIME_SHORT;
      default:
        return DateTime.DATETIME_FULL;
    }
  }
</script>

{#if dateTime.isValid}
  <Tooltip distance={8} location="top" suppress={clientWidth > 180}>
    <div bind:clientWidth class=" w-full">
      {formattedDate}
    </div>
    <TooltipContent slot="tooltip-content">
      <span class="text-xs text-gray-50 font-medium">
        {full}
      </span>
    </TooltipContent>
  </Tooltip>
{:else}
  <span class="text-gray-500">-</span>
{/if}
