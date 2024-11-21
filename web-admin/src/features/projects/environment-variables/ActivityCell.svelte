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

    return `Updated ${timeAgo(updatedDate)}`;
  }

  $: formattedText = formatText();
</script>

<Tooltip distance={8} location="top" alignment="start">
  <div class="text-xs text-gray-500 cursor-pointer">
    {formattedText}
  </div>
  <TooltipContent slot="tooltip-content">
    <span class="text-xs text-gray-50 font-medium"
      >{new Date(updatedOn).toLocaleString()}</span
    >
  </TooltipContent>
</Tooltip>
