<script lang="ts">
  import type {
    Alignment,
    Location,
  } from "@rilldata/web-common/lib/place-element";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let disabled = false;
  export let rounded = false;
  export let size = 24;
  export let bgGray = false;
  export let active = false;
  export let disableTooltip = false;
  export let disableHover = false;

  export let tooltipDistance = 8;
  export let tooltipLocation: Location = "bottom";
  export let tooltipAlignment: Alignment = "middle";
  export let marginClasses = "";
  export let ariaLabel = "";
</script>

<Tooltip
  distance={tooltipDistance}
  location={tooltipLocation}
  alignment={tooltipAlignment}
>
  <button
    type="button"
    on:click
    {disabled}
    aria-label={ariaLabel}
    class:cursor-auto={disabled}
    class:rounded
    class="{marginClasses} grid place-items-center {disabled
      ? 'text-gray-400'
      : 'text-gray-800'}
{disableHover || disabled
      ? ''
      : bgGray
        ? 'hover:bg-gray-200'
        : 'hover:bg-gray-50'}"
    class:bg-gray-100={active}
    style:width="{size}px"
    style:height="{size}px"
    style:font-size="18px"
  >
    <slot />
  </button>
  <div slot="tooltip-content">
    {#if $$slots["tooltip-content"] && !disableTooltip}
      <TooltipContent>
        <slot name="tooltip-content" />
      </TooltipContent>
    {/if}
  </div>
</Tooltip>
