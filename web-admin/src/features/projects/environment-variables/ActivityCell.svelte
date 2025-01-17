<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { timeAgo } from "../../dashboards/listing/utils";

  export let updatedOn: string;

  function formatText() {
    const updatedDate = new Date(updatedOn);

    // FIXME: `updateProjectVariables` does not update the `updatedOn` timestamp correctly
    // if (createdDate.getTime() === updatedDate.getTime()) {
    //   return `Added ${timeAgo(createdDate)}`;
    // }

    return `${timeAgo(updatedDate)}`;
  }

  $: formattedText = formatText();
</script>

<div class="flex justify-start items-center">
  <Tooltip distance={8} location="top">
    <div
      class="flex flex-row gap-x-1 text-gray-500 cursor-pointer w-fit truncate line-clamp-1"
    >
      {formattedText}
    </div>
    <TooltipContent slot="tooltip-content">
      <span class="text-xs text-gray-50 font-medium">
        {new Date(updatedOn).toLocaleString()}
      </span>
    </TooltipContent>
  </Tooltip>
</div>
