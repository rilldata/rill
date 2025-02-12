<script lang="ts" context="module">
  import { writable } from "svelte/store";

  export const dropZone = (() => {
    const { subscribe, set } = writable<string | null>(null);

    return {
      subscribe,
      set: (id: string) => {
        set(id);
      },
      clear: () => {
        set(null);
      },
    };
  })();

  export const hoveredDivider = (() => {
    const { subscribe, set } = writable<string | null>(null);
    let timeout: ReturnType<typeof setTimeout> | null = null;

    return {
      subscribe,
      set: (id: string) => {
        if (timeout) clearTimeout(timeout);
        set(id);
      },
      reset: () => {
        timeout = setTimeout(() => set(null), 50);
      },
    };
  })();

  export const activeDivider = (() => {
    const { subscribe, set } = writable<string | null>(null);

    return {
      subscribe,
      set,
      reset: () => set(null),
    };
  })();
</script>

<script lang="ts">
  import { ArrowLeftRight } from "lucide-svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";

  export let resizeIndex: number;
  export let addIndex: number;
  export let rowLength: number;
  export let rowIndex: number;
  export let columnWidth: number | undefined = undefined;
  export let isSpreadEvenly: boolean;
  export let addItems: (
    item: {
      position: { row: number; order: number };
      type: CanvasComponentType;
    }[],
  ) => void;
  export let onMouseDown: ((e: MouseEvent) => void) | undefined = undefined;
  export let spreadEvenly: (rowIndex: number) => void;
  export let onMouseEnter = () => {};

  $: firstElement = addIndex === 0;
  $: lastElement = addIndex === rowLength;

  $: dividerId = `row:${rowIndex}::column:${resizeIndex}`;

  $: isActiveDivider = $activeDivider === dividerId;
  $: isHoveredDivider = $hoveredDivider === dividerId;

  $: dropId = `row:${rowIndex}::column:${addIndex}`;
  $: isDropZone = $dropZone === dropId;

  $: notActiveDivider = !!$activeDivider && !isActiveDivider;

  $: showAddComponent = isHoveredDivider && !isActiveDivider;

  $: if (isActiveDivider) {
    document.body.style.cursor = "col-resize";
  } else {
    document.body.style.cursor = "";
  }

  $: resizeDisabled =
    resizeIndex === -1 || rowLength >= 4 || resizeIndex === rowLength - 1;

  function onItemClick(type: CanvasComponentType) {
    activeDivider.reset();

    if (type) {
      addItems([{ position: { row: rowIndex, order: addIndex }, type }]);
    }
  }

  function focus() {
    activeDivider.set(dividerId);
  }

  function hover() {
    hoveredDivider.set(dividerId);
  }
</script>

<button
  disabled={resizeDisabled}
  data-width={columnWidth}
  data-row={rowIndex}
  data-column={resizeIndex}
  class:show-on-left={firstElement}
  class:show-on-right={!firstElement}
  style:pointer-events={notActiveDivider || isDropZone ? "none" : "auto"}
  style:height="calc(100% - 16px)"
  class="absolute top-2 flex items-center justify-center w-4 disabled:opacity-60 z-10 disabled:cursor-not-allowed cursor-col-resize"
  on:mousedown={(e) => {
    if (onMouseDown) onMouseDown(e);
    focus();
  }}
  on:mouseenter={() => {
    onMouseEnter();

    hover();
  }}
  on:mouseleave={hoveredDivider.reset}
>
  <span
    class:bg-primary-300={isActiveDivider || isHoveredDivider || isDropZone}
    class="pointer-events-none flex-none z-10 w-[3px] h-full rounded-full"
  />
</button>

{#if showAddComponent}
  <div
    role="presentation"
    class:show-on-left={firstElement}
    class:show-on-right={!firstElement}
    class:nudge-right={firstElement}
    class:nudge-left={lastElement}
    class="flex flex-col pointer-events-auto shadow-sm absolute top-1/2 w-fit z-20 bg-white -translate-y-1/2 border rounded-sm"
    on:mouseleave={hoveredDivider.reset}
    on:mouseenter={hover}
  >
    <AddComponentDropdown
      {onItemClick}
      onMouseEnter={hover}
      disabled={rowLength >= 4}
    />

    {#if !isSpreadEvenly}
      <button
        class="h-7 px-1 grid place-content-center border-t hover:bg-gray-100 text-slate-500"
        on:click={(e) => {
          e.stopPropagation();
          e.preventDefault();
          spreadEvenly(rowIndex);
          hoveredDivider.reset();
        }}
      >
        <ArrowLeftRight size="15px" />
      </button>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .show-on-left {
    @apply left-0 -translate-x-1/2;
  }

  .show-on-right {
    @apply right-0 translate-x-1/2;
  }

  .nudge-right {
    @apply left-3;
  }

  .nudge-left {
    @apply right-3;
  }
</style>
