<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { Builder } from "bits-ui";
  import { builderActions, getAttrs } from "bits-ui";

  // utilize the ID for easier UI testing.
  export let id: string;
  export let testId: string = "";
  export let suppressTooltip = false;
  export let tooltipText: string;
  export let label: string | undefined = undefined;
  export let builders: Builder[] = [];
</script>

<!-- Opening the ContextMenu causes this tooltip to flash in another location, likely due 
  to a race condition. Disabling the tooltip for now.   -->
<Tooltip distance={16} location="right" suppress={true}>
  <button
    {...getAttrs(builders)}
    aria-label={label}
    class="group-hover:w-fit"
    class:!w-fit={suppressTooltip}
    {id}
    data-testid={testId}
    on:click|preventDefault|stopPropagation
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
    @apply flex justify-center items-center;
    @apply text-gray-500;
    @apply transition-transform duration-100;
    width: 0px;
  }

  button:hover,
  button:focus {
    @apply outline-none bg-slate-300;
  }
</style>
