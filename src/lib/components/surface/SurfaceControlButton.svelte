<script>
  /**
   * contains the placement and styling of the surface view buttons.
   */
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  export let show = false;
  export let left = undefined;
  export let right = undefined;
  export let top = "calc(var(--header-height) / 2 - 24px / 2)";
</script>

<Tooltip location="bottom" alignment="start" distance={12}>
  <button
    class="
        fixed 
        z-40 
        grid place-items-center
        hover:text-gray-500 
        rounded
        bg-transparent
        hover:bg-gray-300
        {show ? 'opacity-100 text-gray-500' : 'opacity-0 text-white'} 
        focus:opacity-100 
        hover:opacity-100
        transition-opacity
        "
    style:width="24px"
    style:height="24px"
    style:left
    style:right
    style:top
    on:click
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    <slot name="tooltip-content" />
  </TooltipContent>
</Tooltip>
