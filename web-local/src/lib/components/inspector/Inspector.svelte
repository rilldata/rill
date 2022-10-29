<script lang="ts">
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { Writable, writable } from "svelte/store";
  import { inspectorVisibilityTween } from "../../application-state-stores/layout-store";
  import { drag } from "../../drag";
  import Portal from "../Portal.svelte";

  export let inspectorID: string;

  /** the core inspector width element is stored in localStorage. */
  //const inspectorBasicWidth = localStorageStore<number>(400, inspectorID);
  const inspectorBasicWidth = getContext(
    "rill:app:inspector-width"
  ) as Writable<number>;

  //const inspectorWidth = tweened($inspectorBasicWidth, { duration: 50 });
  const inspectorWidth = getContext(
    "rill:app:inspector-width-tween"
  ) as Writable<number>;

  inspectorBasicWidth.subscribe((value) => {
    inspectorWidth.set(value);
  });

  export const SURFACE_SLIDE_DURATION = 400;
  export const SURFACE_SLIDE_EASING = cubicOut;

  export const SURFACE_DRAG_DURATION = 50;

  export const visibilityTween = tweened(0, {
    duration: SURFACE_SLIDE_DURATION,
    easing: SURFACE_SLIDE_EASING,
  });

  export const inspectorVisible = writable(true);
  inspectorVisible.subscribe((tf) => {
    visibilityTween.set(tf ? 0 : 1);
  });

  // create local storage elements here
</script>

<div
  class="fixed"
  aria-hidden={!$inspectorVisible}
  style:right="{$inspectorWidth.value * (1 - $inspectorVisibilityTween)}px"
>
  <div
    class="
      bg-white
        border-l 
        border-gray-200 
        fixed 
        overflow-auto 
        transition-colors
        h-screen
      "
    class:hidden={$visibilityTween === 1}
    class:pointer-events-none={!$inspectorVisible}
    style:top="0px"
    style:width="{$inspectorWidth.value}px"
  >
    <!-- draw handler -->
    {#if $inspectorVisible}
      <Portal>
        <div
          class="fixed drawer-handler w-4 hover:cursor-col-resize translate-x-2 h-screen"
          style:right="{(1 - $inspectorVisibilityTween) *
            $inspectorWidth.value}px"
          use:drag={{ minSize: 300, store: inspectorBasicWidth, reverse: true }}
          on:dblclick={() => {
            inspectorBasicWidth.update((state) => {
              state.value = 400;
              return state;
            });
          }}
        />
      </Portal>
    {/if}

    <div style="width: 100%;">
      <slot />
    </div>
  </div>
</div>
