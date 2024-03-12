<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { Builder } from "bits-ui";
  import { getAttrs, builderActions } from "bits-ui";

  export let rotated = false;
  export let suppressTooltip = false;
  export let tooltipText: string;
  export let location = "right";
  // utilize the ID for easier UI testing.
  export let id: string;

  export let rounded = false;
  export let label: string | undefined = undefined;
  export let builders: Builder[] = [];
</script>

<Tooltip
  {location}
  alignment="middle"
  distance={16}
  suppress={suppressTooltip || tooltipText === undefined}
>
  <button
    use:builderActions={{ builders }}
    {...getAttrs(builders)}
    on:click|preventDefault
    {id}
    class:rounded
    class:-rotate-90={rotated}
    class:opacity-100={suppressTooltip}
    class="
    group-hover:opacity-100
    opacity-0
    grid

    h-full aspect-square
        focus:outline-none
        focus:bg-gray-300
        hover:bg-gray-300
        text-gray-500
      
        items-center
        justify-center
        border-transparent
        hover:border-gray-400"
    aria-label={label}
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    {tooltipText}
  </TooltipContent>
</Tooltip>
