<script lang="ts">
  import { ArrowLeftRight } from "lucide-svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";

  export let resizeIndex: number;
  export let addIndex: number;
  export let rowLength: number;
  export let rowIndex: number;
  export let left = false;
  export let columnWidth: number | undefined = undefined;
  export let activelyResizing: boolean;
  export let hoveringOnDropZone: boolean;
  export let addItems: (
    item: {
      position: { row: number; order: number };
      type: CanvasComponentType;
    }[],
  ) => void;
  export let onMouseDown: ((e: MouseEvent) => void) | undefined = undefined;
  export let spreadEvenly: (rowIndex: number) => void;
  export let onMouseEnter = () => {};

  let timeout: ReturnType<typeof setTimeout> | null = null;
  let hovered = false;

  $: resizeDisabled =
    resizeIndex === -1 || rowLength >= 4 || resizeIndex === rowLength - 1;

  function onItemClick(type: CanvasComponentType) {
    hovered = false;

    if (type) {
      addItems([{ position: { row: rowIndex, order: addIndex }, type }]);
    }
  }

  function clearHoverTimeout() {
    if (timeout) clearTimeout(timeout);
  }

  //   $: console.log($$props);
</script>

<button
  disabled={resizeDisabled}
  data-width={columnWidth}
  data-row={rowIndex}
  data-column={resizeIndex}
  class:left
  class:right={!left}
  class:pointer-events-none={hoveringOnDropZone}
  class:pointer-events-auto={!hoveringOnDropZone}
  style:height="calc(100% - 16px)"
  class="absolute top-2 flex items-center justify-center w-3 disabled:opacity-60 z-10 disabled:cursor-not-allowed cursor-col-resize"
  on:mousedown={onMouseDown}
  on:mouseenter={() => {
    onMouseEnter();
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => (hovered = true), 75);
    hovered = true;
  }}
  on:mouseleave={() => {
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => (hovered = false), 150);
  }}
>
  <span
    class:bg-primary-300={activelyResizing || hoveringOnDropZone || hovered}
    class="pointer-events-none flex-none z-10 w-[3px] h-full"
  />
</button>

{#if !activelyResizing && hovered}
  <div
    role="presentation"
    class:left
    class:right={!left}
    class="flex pointer-events-auto shadow-sm absolute top-1/2 w-fit z-20 bg-white -translate-y-1/2 border rounded-sm"
    on:mouseleave={() => {
      timeout = setTimeout(() => (hovered = false), 150);
    }}
    on:mouseenter={clearHoverTimeout}
  >
    <AddComponentDropdown
      {onItemClick}
      onMouseEnter={clearHoverTimeout}
      disabled={rowLength >= 4}
    />
    <button
      class="h-7 px-2 grid place-content-center border-l hover:bg-gray-100 text-slate-500"
      on:click={(e) => {
        e.stopPropagation();
        e.preventDefault();
        spreadEvenly(rowIndex);
      }}
    >
      <ArrowLeftRight size="15px" />
    </button>
  </div>
{/if}

<style lang="postcss">
  .left {
    @apply left-0 -translate-x-1/2;
  }
  .right {
    @apply right-0 translate-x-1/2;
  }
</style>
