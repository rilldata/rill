<script lang="ts">
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { DEFAULT_INSPECTOR_WIDTH } from "../config";
  import { drag } from "../drag";
  import type { LayoutElement } from "./types";
  import { portal } from "@rilldata/web-common/lib/actions/portal";

  /** the core inspector width element is stored in localStorage. */
  const inspectorLayout = getContext<Writable<LayoutElement>>(
    "rill:app:inspector-layout",
  );

  const inspectorWidth = getContext<Writable<number>>(
    "rill:app:inspector-width-tween",
  );

  const visibilityTween = getContext<Writable<number>>(
    "rill:app:inspector-visibility-tween",
  );
</script>

<div
  class="fixed"
  aria-hidden={!$inspectorLayout.visible}
  style:right="{$inspectorWidth * $visibilityTween}px"
  style:top="var(--header-height)"
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
    style:width="{$inspectorWidth}px"
  >
    <!-- draw handler -->
    {#if $inspectorLayout.visible}
      <div
        use:portal
        role="separator"
        class="fixed drawer-handler w-4 hover:cursor-col-resize translate-x-2 h-screen"
        style:right="{$visibilityTween * $inspectorWidth}px"
        style:top="var(--header-height)"
        style:bottom="0px"
        use:drag={{ minSize: 300, store: inspectorLayout, reverse: true }}
        on:dblclick={() => {
          inspectorLayout.update((state) => {
            state.value = DEFAULT_INSPECTOR_WIDTH;
            return state;
          });
        }}
      />
    {/if}

    <div class="w-full pt-2">
      <slot />
    </div>
  </div>
</div>
