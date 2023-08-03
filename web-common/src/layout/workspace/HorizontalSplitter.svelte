<script lang="ts">
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { SIDE_PAD } from "../config";
  import { drag } from "../drag";
  import type { LayoutElement } from "./types";

  const outputLayout = getContext(
    "rill:app:output-layout"
  ) as Writable<LayoutElement>;

  const outputPosition = getContext(
    "rill:app:output-height-tween"
  ) as Writable<number>;

  const outputVisibilityTween = getContext(
    "rill:app:output-visibility-tween"
  ) as Writable<number>;

  const inspectorWidth = getContext(
    "rill:app:inspector-width-tween"
  ) as Writable<number>;

  const inspectorVisibilityTween = getContext(
    "rill:app:inspector-visibility-tween"
  ) as Writable<number>;

  const navigationWidth = getContext(
    "rill:app:navigation-width-tween"
  ) as Writable<number>;

  const navVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Writable<number>;
</script>

<div
  class="fixed drawer-handler h-4 hover:cursor-col-resize translate-y-2 grid items-center"
  style:bottom="{$outputPosition * $outputVisibilityTween}px"
  style:left="{(1 - $navVisibilityTween) * $navigationWidth + 20}px"
  style:right="{$inspectorVisibilityTween * $inspectorWidth + 20}px"
  style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
  style:padding-right="{(1 - $inspectorVisibilityTween) * SIDE_PAD}px"
  use:drag={{
    minSize: 200,
    maxSize: innerHeight - 200,
    store: outputLayout,
    orientation: "vertical",
    reverse: true,
  }}
>
  <div class="border-t border-gray-300" />
  <div class="absolute right-1/2 left-1/2 top-1/2 bottom-1/2">
    <div
      class="border-gray-400 border bg-white rounded h-1 w-8 absolute -translate-y-1/2"
    />
  </div>
</div>
