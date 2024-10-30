<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getStateManagers } from "../state-managers/state-managers";

  export let dimensionName: string | undefined;

  const {
    selectors: {
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
    actions: {
      comparison: { setComparisonDimension },
    },
  } = getStateManagers();

  $: isBeingCompared =
    dimensionName !== undefined && $isBeingComparedReadable(dimensionName);
</script>

<IconButton
  ariaLabel="Toggle breakdown for {dimensionName} dimension"
  on:click={(e) => {
    if (isBeingCompared) {
      setComparisonDimension(undefined);
    } else {
      setComparisonDimension(dimensionName);
    }
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
