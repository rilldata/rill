<script lang="ts">
  import { setContext } from "svelte";

  import { writable } from "svelte/store";

  import FloatingElement from "../floating-element/FloatingElement.svelte";
  import Portal from "../Portal.svelte";
  export let location = "bottom";
  export let alignment = "middle";
  export let distance = 0;
  export let pad = 8;
  // provide a programmatic guard to suppressing the tooltip.
  export let suppress = false;
  /** the delay in miliseconds before rendering the tooltip once mouse has entered */
  export let activeDelay = 120;
  /** the delay in miliseconds before unrendering the tooltip once mouse has left */
  export let nonActiveDelay = 0;

  export let active = false;

  let parent;

  let waitUntilTimer;
  function waitUntil(callback, time = 120) {
    if (waitUntilTimer) clearTimeout(waitUntilTimer);
    waitUntilTimer = setTimeout(callback, time);
  }

  /** create child-supported suppression.
   * If a child changes this context store to true, we should
   * suppress the tooltip.
   * This enables us to disentangle the tooltip state in certain cases.
   */
  const childRequestedTooltipSuppression = writable(false);
  setContext(
    "rill:app:childRequestedTooltipSuppression",
    childRequestedTooltipSuppression
  );
</script>

<div
  class="contents"
  bind:this={parent}
  on:mouseenter={() => {
    waitUntil(() => {
      active = true;
    }, activeDelay);
  }}
  on:mouseleave={() => {
    waitUntil(() => {
      active = false;
    }, nonActiveDelay);
  }}
>
  <slot />
  {#if active && !suppress && !$childRequestedTooltipSuppression}
    <Portal>
      <div style="z-index:50;">
        <FloatingElement
          target={parent}
          {location}
          {alignment}
          {distance}
          {pad}
        >
          <slot name="tooltip-content" />
        </FloatingElement>
      </div>
    </Portal>
  {/if}
</div>
