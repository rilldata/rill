<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";

  // svelte-ignore custom_element_props_identifier
  let {
    suppressTooltip = false,
    tooltipText,
    label = undefined,
    children,
    ref = $bindable(null),
    ...restProps
  }: HTMLButtonAttributes & {
    suppressTooltip?: boolean;
    tooltipText: string;
    label?: string;
    children?: Snippet;
    ref?: HTMLButtonElement | null;
  } = $props();
</script>

<!-- Opening the ContextMenu causes this tooltip to flash in another location, likely due
  to a race condition. Disabling the tooltip for now.   -->
<Tooltip distance={16} location="right" suppress={true}>
  <button
    {...restProps}
    bind:this={ref}
    aria-label={label}
    class="group-hover:w-fit"
    class:!w-fit={suppressTooltip}
  >
    {@render children?.()}
  </button>
  <TooltipContent slot="tooltip-content">
    {tooltipText}
  </TooltipContent>
</Tooltip>

<style lang="postcss">
  button {
    @apply h-full aspect-square;
    @apply flex justify-center items-center;
    @apply text-fg-secondary;
    @apply transition-transform duration-100;
    width: 0px;
  }

  button:hover,
  button:focus {
    @apply outline-none bg-surface-active;
  }
</style>
