<script lang="ts">
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";

  export let allowDrop: boolean;
  export let resizeIndex = -1;
  export let dropIndex: number;
  export let activelyResizing: boolean;
  export let passedThreshold: boolean;
  export let onDrop: (row: number, column: number | null) => void;
  export let onRowResizeStart: (e: MouseEvent) => void;
  export let addItem: (type: CanvasComponentType) => void;

  let hovered = false;
  let timeout: ReturnType<typeof setTimeout> | null = null;

  $: showAddComponent = !allowDrop && !activelyResizing && hovered;
  $: dropOnly = resizeIndex === -1;
</script>

<div
  role="presentation"
  class:pointer-events-none={!allowDrop}
  style:width="calc(100% + 160px)"
  class:top={dropOnly}
  class:bottom={!dropOnly}
  class="absolute z-10 -left-20 h-20 flex items-center justify-center px-2"
  on:mouseenter={() => {
    if (timeout) clearTimeout(timeout);
    hovered = true;
  }}
  on:mouseleave={() => {
    timeout = setTimeout(() => {
      hovered = false;
    }, 150);
  }}
  on:mouseup={() => {
    onDrop(dropIndex, null);
  }}
>
  <span
    class:!hidden={!showAddComponent}
    class="flex pointer-events-auto shadow-sm absolute left-1/2 w-fit z-50 bg-white -translate-x-1/2 border rounded-sm"
  >
    <AddComponentDropdown
      onMouseEnter={() => {
        if (timeout) clearTimeout(timeout);
        hovered = true;
      }}
      onItemClick={(type) => {
        hovered = false;
        addItem(type);
      }}
    />
  </span>

  <button
    data-row={resizeIndex}
    class:cursor-not-allowed={dropOnly && !allowDrop}
    class:cursor-row-resize={!allowDrop && !dropOnly}
    class="mx-20 w-full h-4 group z-40 flex items-center justify-center pointer-events-auto"
    on:mousedown={onRowResizeStart}
  >
    <span
      class:bg-primary-300={hovered && (!allowDrop || passedThreshold)}
      class="w-full h-[3px] group-hover:bg-primary-300"
    />
  </button>
</div>

<style lang="postcss">
  /* div {
    @apply bg-blue-400/20;
  } */

  .top {
    @apply top-0  -translate-y-1/2;
  }

  .bottom {
    @apply bottom-0 translate-y-1/2;
  }
</style>
