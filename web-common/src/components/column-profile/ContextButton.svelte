<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { Builder } from "bits-ui";
  import { getAttrs, builderActions } from "bits-ui";

  // utilize the ID for easier UI testing.
  export let id: string;
  export let suppressTooltip = false;
  export let tooltipText: string;
  export let label: string | undefined = undefined;
  export let builders: Builder[] = [];
</script>

<Tooltip location="right" distance={16} suppress={suppressTooltip}>
  <button
    {id}
    class="group-hover:opacity-100"
    class:!opacity-100={suppressTooltip}
    aria-label={label}
    {...getAttrs(builders)}
    on:click|preventDefault
    use:builderActions={{ builders }}
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    {tooltipText}
  </TooltipContent>
</Tooltip>

<style lang="postcss">
  button {
    @apply h-full aspect-square;
    @apply grid place-content-center;
    @apply text-gray-500 opacity-0;
    @apply transition-transform duration-100;
  }

  button:hover,
  button:focus {
    @apply outline-none bg-gray-300;
  }
</style>
