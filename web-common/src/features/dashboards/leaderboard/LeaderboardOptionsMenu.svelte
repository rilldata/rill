<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createEventDispatcher } from "svelte";

  export let isBeingCompared = false;

  const dispatch = createEventDispatcher();
</script>

<IconButton
  on:click={(e) => {
    dispatch("toggle-dimension-comparison");
    e.stopPropagation();
  }}
>
  <Tooltip location="left" distance={8}>
    <Compare isColored={isBeingCompared} />
    <TooltipContent slot="tooltip-content">
      {isBeingCompared ? "Remove comparison" : "Compare"}
    </TooltipContent>
  </Tooltip>
</IconButton>
