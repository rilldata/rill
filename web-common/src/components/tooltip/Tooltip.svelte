<!-- @component 
The Tooltip component is a workhorse generalized-render component wrapper.
By default, it will take the first DOM element in the <slot /> and 
listen to mouseover changes to it to trigger the tooltip.

You can set the following parameters:
- location: left, right, top, bottom.
- alignment: start, middle, end. This will align using flexbox-like language.
  - start on left or right means that the top of the tooltip aligns with the top of the target.
  - start on top or  bottom means that the left side of the tooltip aligns with the left side of the target.
- distance: the distance in pixels between the tooltip and the target.
- pad: the padding in pixels between the tooltip and the target.
- suppress: a boolean to suppress the tooltip.
- activeDelay: the delay in miliseconds before activating the tooltip once hover intent is detected.
- hideDelay: the delay in miliseconds before hiding the tooltip once mouse has left.
- hoverIntentThreshold: the threshold in pixels for hover intent detection.
- hoverIntentTimeout: the time in milliseconds to wait for hover intent.
- active: a boolean to control the active state of the tooltip.

The suppress prop gives programmatic access to suppressing the tooltip, e.g. if a menu is open or you just need
to stop all tooltips.

This component also passes down a boolean writable store to its children, so a child can suppress the tooltip
without the consumer needing to figure out the logic to suppress in the parent component.

FIXME: In the future, we should also be listening to focus events from the child.
-->
<script lang="ts">
  import { setContext } from "svelte";
  import FloatingElement from "../floating-element/FloatingElement.svelte";
  import type {
    Alignment,
    Location,
  } from "@rilldata/web-common/lib/place-element";
  import { hoverIntent } from "./hover-intent";
  import {
    CHILD_REQUESTED_TOOLTIP_SUPPRESSION_CONTEXT_KEY,
    childRequestedTooltipSuppression,
  } from "./store";

  export let location: Location = "bottom";
  export let alignment: Alignment = "middle";
  export let distance = 0;
  export let pad = 8;
  export let suppress = false;
  export let activeDelay = 200;
  export let hideDelay = 0;
  export let hoverIntentThreshold = 5;
  export let hoverIntentTimeout = 100;
  export let active = false;

  let parent: HTMLDivElement;

  setContext(
    CHILD_REQUESTED_TOOLTIP_SUPPRESSION_CONTEXT_KEY,
    childRequestedTooltipSuppression,
  );
</script>

<div
  role="tooltip"
  class="contents"
  bind:this={parent}
  use:hoverIntent={{
    threshold: hoverIntentThreshold,
    timeout: hoverIntentTimeout,
    activeDelay,
    hideDelay,
    onActiveChange: (value) => (active = value),
  }}
>
  <slot />
</div>
{#if active && !suppress && !$childRequestedTooltipSuppression}
  <FloatingElement target={parent} {location} {alignment} {distance} {pad}>
    <slot name="tooltip-content" />
  </FloatingElement>
{/if}

<style>
  .contents {
    display: contents;
  }
</style>
