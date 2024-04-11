<script lang="ts">
  import { navEntryDragDropStore } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { portal } from "@rilldata/web-common/lib/actions/portal";

  export let offset: { x: number; y: number };
  export let position = { left: 0, top: 0 };
  const { navDragging } = navEntryDragDropStore;

  function trackDragItem(e: MouseEvent) {
    requestAnimationFrame(() => {
      position = {
        left: e.clientX - offset.x,
        top: e.clientY - offset.y,
      };
    });
  }

  function onDragRelease() {
    navDragging.set(null);
  }
</script>

<svelte:window on:mousemove={trackDragItem} on:mouseup={onDragRelease} />

<div
  class="portal-item"
  style:left="{position.left}px"
  style:top="{position.top}px"
  use:portal
>
  {$navDragging?.fileName ?? ""}
</div>

<style lang="postcss">
  .portal-item {
    @apply shadow-lg shadow-slate-300;
    @apply z-50;
    @apply absolute pointer-events-none;
  }
</style>
