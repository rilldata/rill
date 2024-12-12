<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Compare from "@rilldata/web-common/components/icons/Compare.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let dimensionName: string | undefined;
  export let isBeingCompared: boolean;
  export let toggleComparisonDimension: (dimensionName: string) => void;
</script>

<IconButton
  ariaLabel="Toggle breakdown for {dimensionName} dimension"
  on:click={(e) => {
    e.stopPropagation();
    if (dimensionName) toggleComparisonDimension(dimensionName);
  }}
>
  <Tooltip location="left" distance={8}>
    <Compare isColored={isBeingCompared} />
    <TooltipContent slot="tooltip-content">
      {isBeingCompared ? "Remove comparison" : "Compare"}
    </TooltipContent>
  </Tooltip>
</IconButton>
