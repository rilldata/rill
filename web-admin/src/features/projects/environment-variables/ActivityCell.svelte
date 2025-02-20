<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { DateTime } from "luxon";
  import TimeAgo from "./TimeAgo.svelte";

  export let updatedOn: string;

  $: fullDate = DateTime.fromISO(updatedOn).toLocaleString(
    DateTime.DATETIME_FULL,
  );
</script>

<div class="flex justify-start items-center">
  <Tooltip distance={8} location="top">
    <div
      class="flex flex-row gap-x-1 text-gray-500 cursor-pointer w-fit truncate line-clamp-1"
    >
      <TimeAgo datetime={updatedOn} />
    </div>
    <TooltipContent slot="tooltip-content">
      <span class="text-xs text-gray-50 font-medium">
        {fullDate}
      </span>
    </TooltipContent>
  </Tooltip>
</div>
