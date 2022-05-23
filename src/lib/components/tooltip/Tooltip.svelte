<script lang="ts">
  import FloatingElement from "./FloatingElement.svelte";
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
  {#if active && !suppress}
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
