<script lang="ts">
  import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
  import { getContext, setContext } from "svelte";
  import { tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";
  import {
    DEFAULT_INSPECTOR_WIDTH,
    DEFAULT_PREVIEW_TABLE_HEIGHT,
    SIDE_PAD,
    SURFACE_SLIDE_DURATION,
    SURFACE_SLIDE_EASING,
  } from "../config";
  import Inspector from "./Inspector.svelte";
  import type { LayoutElement } from "./types";

  export let assetID: string;
  export let inspector = true;
  export let bgClass = "bg-gray-100";

  const inspectorLayout = localStorageStore<LayoutElement>(assetID, {
    value: inspector ? DEFAULT_INSPECTOR_WIDTH : 0,
    visible: true,
  });
  const inspectorWidth = tweened(
    inspector ? $inspectorLayout?.value || DEFAULT_INSPECTOR_WIDTH : 0,
    {
      duration: 50,
    },
  );
  inspectorLayout.subscribe((state) => {
    inspectorWidth.set(state.value);
  });

  export const visibilityTween = tweened($inspectorLayout?.visible ? 1 : 0, {
    duration: SURFACE_SLIDE_DURATION,
    easing: SURFACE_SLIDE_EASING,
  });

  /** when the inspector visibility changes, trigger the tween. */
  inspectorLayout.subscribe((state) => {
    visibilityTween.set(state.visible ? 1 : 0);
  });

  setContext("rill:app:inspector-layout", inspectorLayout);
  setContext("rill:app:inspector-width-tween", inspectorWidth);
  setContext("rill:app:inspector-visibility-tween", visibilityTween);

  const outputLayout = localStorageStore<LayoutElement>(`${assetID}-output`, {
    value: inspector ? DEFAULT_PREVIEW_TABLE_HEIGHT : 0,
    visible: true,
  });

  const outputHeight = tweened(
    inspector ? $outputLayout?.value || DEFAULT_PREVIEW_TABLE_HEIGHT : 0,
    {
      duration: 50,
    },
  );

  outputLayout.subscribe((state) => {
    outputHeight.set(state.value);
  });

  export const outputVisibilityTween = tweened(
    $inspectorLayout?.visible ? 1 : 0,
    {
      duration: SURFACE_SLIDE_DURATION,
      easing: SURFACE_SLIDE_EASING,
    },
  );

  outputLayout.subscribe((state) => {
    outputVisibilityTween.set(state.visible ? 1 : 0);
  });

  setContext("rill:app:output-layout", outputLayout);
  setContext("rill:app:output-height-tween", outputHeight);
  setContext("rill:app:output-visibility-tween", outputVisibilityTween);

  const navigationWidth = getContext<Writable<number>>(
    "rill:app:navigation-width-tween",
  );
  const navVisibilityTween = getContext<Writable<number>>(
    "rill:app:navigation-visibility-tween",
  );

  // Unclear on usage of this variable
  // Holding off on removing it for now
  let hasNoError = 1;
</script>

<div
  class="flex flex-col h-screen overflow-hidden absolute"
  style:left="{($navigationWidth || 0) * (1 - $navVisibilityTween)}px"
  style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
  style:right="0px"
>
  {#if $$slots.header}
    <header class="bg-white w-full h-fit z-10">
      <slot name="header" />
    </header>
  {/if}

  <div
    class="h-full {bgClass}"
    style:padding-right="{inspector && hasNoError
      ? $inspectorWidth * $visibilityTween
      : 0}px"
  >
    <slot name="body" />
  </div>
</div>

{#if inspector}
  <Inspector>
    <slot name="inspector" />
  </Inspector>
{/if}
