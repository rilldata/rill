<script lang="ts">
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import Portal from "@rilldata/web-common/components/Portal.svelte";
  import SurfaceControlButton from "@rilldata/web-local/lib/components/surface/SurfaceControlButton.svelte";
  import { drag } from "@rilldata/web-local/lib/drag";
  import type { LayoutElement } from "@rilldata/web-local/lib/types";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";

  /** the core inspector width element is stored in localStorage. */
  const inspectorLayout = getContext(
    "rill:app:inspector-layout"
  ) as Writable<LayoutElement>;

  const inspectorWidth = getContext(
    "rill:app:inspector-width-tween"
  ) as Writable<number>;

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
    <HideRightSidebar size="18px" />
  {:else}
    <MoreHorizontal size="16px" />
  {/if}
  <svelte:fragment slot="tooltip-content">
    {#if $visibilityTween === 1} Close {:else} Show {/if} sidebar
  </svelte:fragment>
</SurfaceControlButton>
