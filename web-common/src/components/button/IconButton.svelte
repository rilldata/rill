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
  export let ariaPressed: boolean | undefined = undefined;
</script>

<Tooltip
  distance={tooltipDistance}
  location={tooltipLocation}
  alignment={tooltipAlignment}
>
  <button
    type="button"
    on:click
    aria-label={ariaLabel}
    aria-pressed={ariaPressed}
    class:cursor-auto={disabled}
    class:rounded
    class:text-fg-disabled={disabled}
    class="{marginClasses} grid place-items-center text-fg-muted hover:text-fg-secondary
{disableHover || disabled
      ? ''
      : bgGray
        ? 'hover:bg-surface-hover'
        : 'hover:bg-surface-background'}"
    class:bg-surface-active={active}
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
