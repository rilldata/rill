<script lang="ts">
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";
  import Divider from "./Divider.svelte";
  import { dropZone, activeDivider } from "./stores/ui-stores";

  export let allowDrop: boolean;
  export let resizeIndex = -1;
  export let dropIndex: number;
  export let resizingRow = false;
  export let onDrop: (row: number, column: number | null) => void;
  export let onRowResizeStart: (e: MouseEvent) => void = () => {};
  export let addItem: (type: CanvasComponentType) => void;

  let menuOpen = false;

  $: dividerId = `row:${resizeIndex}::column:null`;
  $: dropId = `row:${dropIndex}::column:null`;

  $: isDropZone = $dropZone === dropId;
  $: isActiveDivider = $activeDivider === dividerId;

  $: notActiveDivider = !isActiveDivider && !!$activeDivider;

  $: notResizable = resizeIndex === -1;

  $: forceShowDivider = menuOpen || resizingRow || isDropZone;
</script>

<div
  role="presentation"
  style:pointer-events={allowDrop ? "auto" : "none"}
  class:top={notResizable}
  class:bottom={!notResizable}
  class="absolute z-10 w-full h-12 flex items-center justify-center px-2"
  on:mouseenter={() => {
    if (!allowDrop) return;

    dropZone.set(dropId);
  }}
  on:mouseleave={() => {
    if (!allowDrop || menuOpen) return;
    dropZone.clear();
  }}
  on:mouseup={() => {
    if (!allowDrop) return;
    onDrop(dropIndex, null);
  }}
>
  <div class="w-full h-4 z-40 group flex items-center justify-center relative">
    <button
      data-row={resizeIndex}
      class:cursor-default={notResizable && !allowDrop}
      class:cursor-row-resize={!allowDrop && !notResizable}
      style:pointer-events={notActiveDivider || allowDrop || menuOpen
        ? "none"
        : "auto"}
      class="peer size-full flex items-center justify-center px-px"
      on:mousedown={(e) => {
        onRowResizeStart(e);
        activeDivider.set(dividerId);

        window.addEventListener(
          "mouseup",
          () => {
            activeDivider.set(null);
          },
          { once: true },
        );
      }}
    >
      <Divider horizontal show={forceShowDivider} />
    </button>

    <span
      role="presentation"
      class:shift-down={notResizable}
      class:not-sr-only={menuOpen}
      class="sr-only peer-hover:not-sr-only peer-active:sr-only hover:not-sr-only flex shadow-sm pointer-events-auto !absolute left-1/2 w-fit z-50 bg-white -translate-x-1/2 border rounded-sm"
    >
      <AddComponentDropdown
        bind:open={menuOpen}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            activeDivider.set(null);
          } else {
            activeDivider.set(dividerId);
          }
        }}
        onItemClick={addItem}
      />
    </span>
  </div>
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

  .shift-down {
    @apply translate-y-[8px];
  }
</style>
