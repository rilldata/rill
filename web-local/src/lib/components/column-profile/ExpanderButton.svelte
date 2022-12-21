<script>
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createHoverStateActionFactory } from "@rilldata/web-common/lib/actions/hover-state-action-factory";

  export let rotated = false;
  export let suppressTooltip = false;
  export let isHovered = false;

  const { hovered, captureHoverState } = createHoverStateActionFactory();
  hovered.subscribe((trueOrFalse) => {
    isHovered = trueOrFalse;
  });
</script>

<Tooltip
  location="top"
  alignment="start"
  distance={8}
  pad={8}
  suppress={suppressTooltip}
>
  <button
    on:click
    use:captureHoverState
    style:width="20px"
    style:height="20px"
    style:grid-column="left-control"
    class="
    hover:bg-gray-300
    transition-tranform 
    text-gray-500
    duration-100
    grid
    items-center
    justify-center
    rounded
    {!rotated ? '-rotate-90' : ''}"
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    <!-- {!rotated ? "show columns" : "hide columne"} -->
    <SlidingWords active={rotated}>columns</SlidingWords>
  </TooltipContent>
</Tooltip>
