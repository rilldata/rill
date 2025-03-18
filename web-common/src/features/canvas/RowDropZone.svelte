<script lang="ts">
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";
  import Divider from "./Divider.svelte";
  import { dropZone, hoveredDivider } from "./stores/ui-stores";

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

  const { id, isActive } = hoveredDivider;

  $: hoveredId = $id;
  $: active = $isActive;

  $: isHoveredDivider = hoveredId === dividerId;
  $: isDropZone = $dropZone === dropId;
  $: isActiveDivider = isHoveredDivider && active;

  $: notActiveDivider = active && !isHoveredDivider;

  $: showAddComponent = menuOpen || (isHoveredDivider && !active);
  $: notResizable = resizeIndex === -1;

  $: showDivider = isHoveredDivider || menuOpen || resizingRow || isDropZone;

  // $: if (!showAddComponent) {
  //   open = false;
  // }

  // $: console.log(
  //   dividerId,
  //   hoveredId,
  //   active,
  //   isHoveredDivider,
  //   isActiveDivider,
  //   showAddComponent,
  // );

  // function focus(bool = true) {
  //   if (bool) hoveredDivider.setActive(true);
  //   else hoveredDivider.setActive(false);
  // }

  function hover(bool = true) {
    if (bool) hoveredDivider.set(dividerId);
    else hoveredDivider.reset();
  }
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
  <button
    data-row={resizeIndex}
    class:cursor-default={notResizable && !allowDrop}
    class:cursor-row-resize={!allowDrop && !notResizable}
    style:pointer-events={notActiveDivider || allowDrop ? "none" : "auto"}
    class="w-full h-4 z-40 group flex items-center justify-center px-px"
    on:mousedown={(e) => {
      onRowResizeStart(e);
      hoveredDivider.setActive(dividerId, true);
    }}
    on:mouseenter={() => hover()}
    on:mouseleave={() => {
      if (!isActiveDivider) hover(false);
    }}
  >
    <Divider horizontal show={showDivider} />
  </button>

  {#if showAddComponent}
    <span
      role="presentation"
      class:shift-down={notResizable}
      class="flex shadow-sm pointer-events-auto absolute left-1/2 w-fit z-50 bg-white -translate-x-1/2 border rounded-sm"
      on:mouseleave={() => {
        console.log("mouseleave");
        // if (!open) hoveredDivider.reset();
      }}
    >
      <AddComponentDropdown
        bind:open={menuOpen}
        {dividerId}
        onMouseEnter={() => {
          if (!menuOpen) hover();
        }}
        onItemClick={(type) => {
          // hoveredDivider.reset();
          addItem(type);
        }}
      />
    </span>
  {/if}
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
