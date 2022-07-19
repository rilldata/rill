<script lang="ts">
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  export let rotated = false;
  export let suppressTooltip = false;
  export let tooltipText: string;
  export let location = "right";
  // utilize the ID for easier UI testing.
  export let id: string;
  export let width = 16;
  export let height = 16;
</script>

<Tooltip
  {location}
  alignment="middle"
  distance={16}
  suppress={suppressTooltip || tooltipText === undefined}
>
  <button
    on:click
    {id}
    style:width={`${width}px`}
    style:height={`${height}px`}
    style:grid-column="left-control"
    class="
        hover:bg-gray-300
        transition-tranform 
        text-gray-500
        duration-100
        grid
        p-0
        items-center
        justify-center
        border
        border-transparent
        hover:border-gray-400
        {rotated ? '-rotate-90' : ''}"
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    {tooltipText}
  </TooltipContent>
</Tooltip>
