<script lang="ts">
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let disabled = false;
  export let rounded = false;
  export let compact = false;
  export let bgDark = false;
  export let active = false;

  export let tooltipDistance = 8;
  export let tooltipLocation = "bottom";
  export let tooltipAlignment = "center";
  // FIXME: invert this so that margin classes have to be explicitly applied
  export let marginClasses = "ml-3";
</script>

<Tooltip
  distance={tooltipDistance}
  location={tooltipLocation}
  alignment={tooltipAlignment}
>
  <button
    on:click
    class:cursor-auto={disabled}
    class:rounded
    class="{marginClasses} grid place-items-center 
{active ? 'bg-gray-200 dark:bg-gray-800' : ''}
{disabled ? 'text-gray-400' : 'text-gray-800'}
{disabled ? '' : bgDark ? 'hover:bg-gray-600' : 'hover:bg-gray-200'}"
    style:width="{compact ? 20 : 24}px"
    style:height="{compact ? 20 : 24}px"
    style:font-size="18px"
  >
    <slot />
  </button>
  <div slot="tooltip-content">
    {#if $$slots["tooltip-content"]}
      <TooltipContent>
        <slot name="tooltip-content" />
      </TooltipContent>
    {/if}
  </div>
</Tooltip>
