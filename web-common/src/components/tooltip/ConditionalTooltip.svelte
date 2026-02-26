<!--
  @component
  ConditionalTooltip - A wrapper that optionally applies a Tooltip.

  Use this when you need to conditionally show a tooltip on content without
  duplicating the inner markup. When `showTooltip` is false, the content
  renders directly without any wrapper.

  @example
  <ConditionalTooltip showTooltip={hasError} location="top">
    <div>Content that may or may not have a tooltip</div>
    <TooltipContent slot="tooltip-content">
      Tooltip message here
    </TooltipContent>
  </ConditionalTooltip>
-->
<script lang="ts">
  import Tooltip from "./Tooltip.svelte";
  import type {
    Alignment,
    Location,
  } from "@rilldata/web-common/lib/place-element";

  /** Whether to show the tooltip. When false, content renders without a wrapper. */
  export let showTooltip: boolean;

  // Tooltip props (pass-through)
  export let location: Location = "bottom";
  export let alignment: Alignment = "middle";
  export let distance = 0;
  export let pad = 8;
  export let suppress = false;
  export let activeDelay = 200;
  export let hideDelay = 0;
</script>

{#if showTooltip}
  <Tooltip
    {location}
    {alignment}
    {distance}
    {pad}
    {suppress}
    {activeDelay}
    {hideDelay}
  >
    <slot />
    <slot name="tooltip-content" slot="tooltip-content" />
  </Tooltip>
{:else}
  <slot />
{/if}
