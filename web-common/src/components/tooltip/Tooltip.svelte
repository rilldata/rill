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
- activeDelay: the delay in miliseconds before rendering the tooltip once mouse has entered.
- nonActiveDelay: the delay in miliseconds before unrendering the tooltip once mouse has left.
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
  import { writable } from "svelte/store";
  import FloatingElement from "../floating-element/FloatingElement.svelte";
  import type {
    Alignment,
    Location,
  } from "@rilldata/web-common/lib/place-element";
  import HoverIntent from "./HoverIntent.svelte";

  export let location: Location = "bottom";
  export let alignment: Alignment = "middle";
  export let distance = 0;
  export let pad = 8;
  export let suppress = false;
  export let activeDelay = 200;
  export let nonActiveDelay = 0;
  export let hoverIntentThreshold = 5;
  export let hoverIntentTimeout = 100;
  export let active = false;

  let parent: HTMLDivElement;

  /** create child-supported suppression.
   * If a child changes this context store to true, we should
   * suppress the tooltip.
   * This enables us to disentangle the tooltip state in certain cases
   * where it doesn't make sense to have the user deal with the logic of suppression.
   */
  const childRequestedTooltipSuppression = writable(false);
  setContext(
    "rill:app:childRequestedTooltipSuppression",
    childRequestedTooltipSuppression,
  );

  function handleActive(event: CustomEvent<boolean>) {
    active = event.detail;
  }
</script>

<div class="contents">
  <HoverIntent
    threshold={hoverIntentThreshold}
    timeout={hoverIntentTimeout}
    {activeDelay}
    {nonActiveDelay}
    on:active={handleActive}
  >
    <div bind:this={parent}>
      <slot />
    </div>
  </HoverIntent>
  {#if active && !suppress && !$childRequestedTooltipSuppression}
    <FloatingElement target={parent} {location} {alignment} {distance} {pad}>
      <slot name="tooltip-content" />
    </FloatingElement>
  {/if}
</div>
