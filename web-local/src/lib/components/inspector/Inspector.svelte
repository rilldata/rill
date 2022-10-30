<script lang="ts">
  import { getContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { Writable } from "svelte/store";
  import { drag } from "../../drag";
  import HideRightSidebar from "../icons/HideRightSidebar.svelte";
  import MoreHorizontal from "../icons/MoreHorizontal.svelte";
  import Portal from "../Portal.svelte";
  import SurfaceControlButton from "../surface/SurfaceControlButton.svelte";

  export let inspectorID: string;

  /** the core inspector width element is stored in localStorage. */
  const inspectorLayout = getContext(
    "rill:app:inspector-layout"
  ) as Writable<number>;

  const inspectorWidth = getContext(
    "rill:app:inspector-width-tween"
  ) as Writable<number>;

  export const SURFACE_SLIDE_DURATION = 400;
  export const SURFACE_SLIDE_EASING = cubicOut;

  export const SURFACE_DRAG_DURATION = 50;

  const visibilityTween = getContext(
    "rill:app:inspector-visibility-tween"
  ) as Writable<number>;

  let inspectorVisible = $inspectorLayout.visible;
  inspectorLayout.subscribe((state) => {
    if (state.visible !== inspectorVisible) {
      visibilityTween.set(state.visible ? 1 : 0);
      inspectorVisible = state.visible;
    }
  });

  let hasNoError = 1;
</script>

<div
  class="fixed"
  aria-hidden={!$inspectorLayout.visible}
  style:right="{$inspectorWidth * $visibilityTween}px"
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
    class:hidden={$visibilityTween === 0}
    class:pointer-events-none={!$inspectorLayout.visible}
    style:top="0px"
    style:width="{$inspectorWidth}px"
  >
    <!-- draw handler -->
    {#if $inspectorLayout.visible}
      <Portal>
        <div
          class="fixed drawer-handler w-4 hover:cursor-col-resize translate-x-2 h-screen"
          style:right="{$visibilityTween * $inspectorWidth}px"
          use:drag={{ minSize: 300, store: inspectorLayout, reverse: true }}
          on:dblclick={() => {
            inspectorLayout.update((state) => {
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

<SurfaceControlButton
  show={true}
  right="{($inspectorWidth - 12 - 24) * ($visibilityTween * hasNoError) +
    12 * (1 - $visibilityTween) * hasNoError}px"
  on:click={() => {
    //inspectorVisible.set(!$inspecto);
    inspectorLayout.update((state) => {
      state.visible = !state.visible;
      return state;
    });
  }}
>
  {#if $inspectorLayout.visible}
    <HideRightSidebar size="20px" />
  {:else}
    <MoreHorizontal size="16px" />
  {/if}
  <svelte:fragment slot="tooltip-content">
    {#if $visibilityTween === 1} close {:else} show {/if} sidebar
  </svelte:fragment>
</SurfaceControlButton>
