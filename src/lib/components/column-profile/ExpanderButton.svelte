<script>
  import SlidingWords from "$lib/components/tooltip/SlidingWords.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  export let rotated = false;
  export let suppressTooltip = false;
  export let isHovered = false;
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
    on:mouseover={() => {
      isHovered = true;
    }}
    on:focus={() => {
      isHovered = true;
    }}
    on:mouseleave={() => {
      isHovered = false;
    }}
    on:blur={() => {
      isHovered = false;
    }}
    style:width="16px"
    style:height="16px"
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
