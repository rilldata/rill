<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { formatRunDate } from "../tableUtils";

  export let reportTime: string;
  export let timeZone: string;
  export let adhoc: boolean;
  export let errorMessage: string;
</script>

<div class="flex gap-x-2 items-center px-4 py-[10px]">
  <div class="text-gray-700 text-sm">
    {formatRunDate(reportTime, timeZone)}
  </div>
  {#if errorMessage === ""}
    <Tag color="blue">Email sent</Tag>
  {:else}
    <Tooltip distance={8}>
      <Tag color="red">Failed</Tag>
      <TooltipContent slot="tooltip-content">
        {errorMessage}
      </TooltipContent>
    </Tooltip>
  {/if}
  {#if adhoc}
    <Tooltip distance={8}>
      <Tag>Ad-hoc</Tag>
      <TooltipContent slot="tooltip-content">
        This report was run manually off-schedule.
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
