<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { portal } from "../../../lib/actions/portal";
  import PivotChip from "./PivotChip.svelte";
  import { type PivotChipData, PivotChipType } from "./types";

  export let item: PivotChipData;
  export let removable: boolean;
  export let offset: { x: number; y: number };
  export let position = { left: 0, top: 0 };
  export let width: number | undefined = undefined;

  const dispatch = createEventDispatcher();

  function trackDragItem(e: MouseEvent) {
    requestAnimationFrame(() => {
      position = {
        left: e.clientX - offset.x,
        top: e.clientY - offset.y,
      };
    });
  }

  function onDragRelease() {
    dispatch("release");
  }
</script>

<svelte:window on:mousemove={trackDragItem} on:mouseup={onDragRelease} />

<div
  class="portal-item"
  class:rounded-full={item.type !== PivotChipType.Measure}
  style:left="{position.left}px"
  style:top="{position.top}px"
  style:width={width ? `${width}px` : "fit-content"}
  use:portal
>
  <PivotChip
    active
    slideDuration={0}
    grab
    fullWidth
    {item}
    {removable}
    on:mousedown
    on:remove
  />
</div>

<style lang="postcss">
  .portal-item {
    @apply shadow-lg;
    z-index: 100;
    @apply absolute pointer-events-none;
  }
</style>
