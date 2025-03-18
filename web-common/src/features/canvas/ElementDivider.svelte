<script lang="ts">
  import { ArrowLeftRight } from "lucide-svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";
  import { dropZone, hoveredDivider } from "./stores/ui-stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Divider from "./Divider.svelte";

  export let resizeIndex: number;
  export let addIndex: number;
  export let rowLength: number;
  export let rowIndex: number;
  export let resizingColumn = false;
  export let columnWidth: number | undefined = undefined;
  export let isSpreadEvenly: boolean;
  export let dragging: boolean;
  export let addItems: (
    position: { row: number; column: number },
    item: CanvasComponentType[],
  ) => void;
  export let onMouseDown: ((e: MouseEvent) => void) | undefined = undefined;
  export let spreadEvenly: (rowIndex: number) => void;
  export let onMouseEnter = () => {};

  let menuOpen = false;

  $: firstElement = addIndex === 0;
  $: lastElement = addIndex === rowLength;

  $: dividerId = `row:${rowIndex}::column:${resizeIndex}`;

  const { id, isActive } = hoveredDivider;

  $: hoveredId = $id;
  $: active = $isActive;

  $: isHoveredDivider = hoveredId === dividerId;
  $: isActiveDivider = isHoveredDivider && active;

  $: dropId = `row:${rowIndex}::column:${addIndex}`;
  $: isDropZone = $dropZone === dropId;

  $: notActiveDivider = active && !isHoveredDivider;

  $: showAddComponent = menuOpen || (isHoveredDivider && !active);

  $: showDivider = isHoveredDivider || menuOpen || resizingColumn || isDropZone;

  $: if (isActiveDivider) {
    document.body.style.cursor = "col-resize";
  } else {
    document.body.style.cursor = "";
  }

  $: addDisabled = rowLength >= 4;

  $: resizeDisabled =
    resizeIndex === -1 || rowLength >= 4 || resizeIndex === rowLength - 1;

  function onItemClick(type: CanvasComponentType) {
    hoveredDivider.setActive(dividerId, false);

    if (type) {
      addItems({ row: rowIndex, column: addIndex }, [type]);
    }
  }

  function hover(bool = true) {
    if (bool) hoveredDivider.set(dividerId);
    else hoveredDivider.reset();
  }
</script>

<!--  This logic still needs tweaking -->
{#if !addDisabled || !isSpreadEvenly || isDropZone}
  <button
    disabled={resizeDisabled}
    data-width={columnWidth}
    data-row={rowIndex}
    data-column={resizeIndex}
    class:show-on-left={firstElement}
    class:show-on-right={!firstElement}
    style:pointer-events={notActiveDivider || dragging ? "none" : "auto"}
    style:height="calc(100% - 16px)"
    class:!opacity-100={isDropZone}
    class="absolute group top-2 flex items-center justify-center w-4 disabled:opacity-60 z-10 disabled:cursor-default cursor-col-resize"
    on:mousedown={(e) => {
      if (onMouseDown) onMouseDown(e);
      hoveredDivider.setActive(dividerId, true);
    }}
    on:mouseenter={() => {
      // menuOpen = false;
      onMouseEnter();

      hover();
    }}
    on:mouseleave={() => {
      if (!isActiveDivider) hover(false);
    }}
  >
    <Divider vertical show={showDivider} />
  </button>

  {#if showAddComponent}
    <div
      role="presentation"
      class:show-on-left={firstElement}
      class:show-on-right={!firstElement}
      class:nudge-right={firstElement}
      class:nudge-left={lastElement}
      class="flex flex-col pointer-events-auto shadow-sm absolute top-1/2 w-fit z-20 bg-white -translate-y-1/2 border rounded-sm"
      on:mouseleave={() => {
        if (!menuOpen) hoveredDivider.reset();
      }}
      on:mouseenter={hover}
    >
      <AddComponentDropdown
        {dividerId}
        onItemClick={(e) => {
          onItemClick(e);
          hoveredDivider.reset();
        }}
        bind:open={menuOpen}
        onMouseEnter={() => {
          if (!open) hover();
        }}
        disabled={rowLength >= 4}
      />

      {#if !isSpreadEvenly}
        <Tooltip distance={8} location="bottom">
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

          <TooltipContent slot="tooltip-content" side="bottom">
            Evenly distribute widgets
          </TooltipContent>
        </Tooltip>
      {/if}
    </div>
  {/if}
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
