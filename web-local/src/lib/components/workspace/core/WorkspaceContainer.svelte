<script lang="ts">
  import {
    DEFAULT_INSPECTOR_WIDTH,
    DEFAULT_PREVIEW_TABLE_HEIGHT,
    SURFACE_SLIDE_DURATION,
    SURFACE_SLIDE_EASING,
  } from "@rilldata/web-local/lib/application-config";
  import { localStorageStore } from "@rilldata/web-local/lib/store-utils";
  import type { LayoutElement } from "@rilldata/web-local/lib/types";
  import { getContext, setContext } from "svelte";
  import { tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";

  export let assetID;
  export let inspector = true;
  /** separate out the view having an inspector from it being currently rendered
   * from the user deciding to hide it
   */
  export let viewHasInspector = inspector;
  export let bgClass = "bg-gray-100";
  export let top = "var(--header-height)";

  const inspectorLayout = localStorageStore<LayoutElement>(assetID, {
    value: viewHasInspector ? DEFAULT_INSPECTOR_WIDTH : 0,
    visible: true,
  });
  const inspectorWidth = tweened(
    viewHasInspector ? $inspectorLayout?.value || DEFAULT_INSPECTOR_WIDTH : 0,
    {
      duration: 50,
    }
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
    value: viewHasInspector ? DEFAULT_PREVIEW_TABLE_HEIGHT : 0,
    visible: true,
  });

  const outputHeight = tweened(
    viewHasInspector ? $outputLayout?.value || DEFAULT_PREVIEW_TABLE_HEIGHT : 0,
    {
      duration: 50,
    }
  );

  outputLayout.subscribe((state) => {
    outputHeight.set(state.value);
  });

  export const outputVisibilityTween = tweened(
    $inspectorLayout?.visible ? 1 : 0,
    {
      duration: SURFACE_SLIDE_DURATION,
      easing: SURFACE_SLIDE_EASING,
    }
  );

  outputLayout.subscribe((state) => {
    outputVisibilityTween.set(state.visible ? 1 : 0);
  });

  setContext("rill:app:output-layout", outputLayout);
  setContext("rill:app:output-height-tween", outputHeight);
  setContext("rill:app:output-visibility-tween", outputVisibilityTween);

  const navigationWidth = getContext(
    "rill:app:navigation-width-tween"
  ) as Writable<number>;
  const navVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Writable<number>;

  let hasNoError = 1;
  let hasInspector = true;
</script>

<div
  class="fixed bg-white"
  style:left="{($navigationWidth || 0) * (1 - $navVisibilityTween)}px"
  style:right="0px"
>
  <slot name="header" />
  <slot />
</div>
<!-- <div
  class="box-border fixed {bgClass}"
  style:top
  style:left="{($navigationWidth || 0) * (1 - $navVisibilityTween)}px"
  style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
  style:padding-right="{(1 - $visibilityTween) *
    SIDE_PAD *
    hasNoError *
    (hasInspector ? 1 : 0)}px"
  style:right="{hasInspector && hasNoError
    ? $inspectorWidth * $visibilityTween
    : 0}px"
> 
  <slot name="body" /> -->
<!-- </div> -->
<!-- {#if inspector}
  <Inspector>
    <slot name="inspector" />
  </Inspector>
{/if} -->
