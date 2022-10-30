<script lang="ts">
  import { getContext, setContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { localStorageStore } from "../stores/local-storage";

  import Inspector from "../inspector/Inspector.svelte";

  export let assetID;
  export let inspector = true;

  /** the core inspector width element is stored in localStorage. */
  interface InspectorStorageValues {
    value: number;
    visible: boolean;
  }
  const inspectorLayout = localStorageStore<InspectorStorageValues>(
    { value: inspector ? 400 : 0, visible: true },
    assetID
  );
  const inspectorWidth = tweened(
    inspector ? $inspectorLayout?.value || 400 : 0,
    {
      duration: 50,
    }
  );
  inspectorLayout.subscribe((state) => {
    inspectorWidth.set(state.value);
  });

  export const SURFACE_SLIDE_DURATION = 400;
  export const SURFACE_SLIDE_EASING = cubicOut;

  export const SURFACE_DRAG_DURATION = 50;

  export const visibilityTween = tweened($inspectorLayout?.visible ? 1 : 0, {
    duration: SURFACE_SLIDE_DURATION,
    easing: SURFACE_SLIDE_EASING,
  });

  setContext("rill:app:inspector-layout", inspectorLayout);
  setContext("rill:app:inspector-width-tween", inspectorWidth);
  setContext("rill:app:inspector-visibility-tween", visibilityTween);

  const navigationWidth = getContext("rill:app:navigation-width-tween");
  const navVisibilityTween = getContext("rill:app:navigation-visibility-tween");

  const SIDE_PAD = 20;
  let hasNoError = 1;
  let hasInspector = true;
</script>

<div
  class="box-border fixed bg-gray-100"
  style:left="{($navigationWidth || 0) * (1 - $navVisibilityTween)}px"
  style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
  style:padding-right="{(1 - $visibilityTween) *
    SIDE_PAD *
    hasNoError *
    (hasInspector ? 1 : 0)}px"
  style:right="{hasInspector && hasNoError
    ? $inspectorWidth * $visibilityTween
    : 0}px"
  style:top="0px"
>
  <slot name="body" />
</div>
{#if inspector}
  <Inspector inspectorID={assetID}>
    <slot name="inspector" />
  </Inspector>
{/if}
