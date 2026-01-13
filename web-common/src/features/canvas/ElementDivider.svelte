<script lang="ts">
  import { ArrowLeftRight } from "lucide-svelte";
  import AddComponentDropdown from "./AddComponentDropdown.svelte";
  import type { CanvasComponentType } from "./components/types";
  import { dropZone, activeDivider } from "./stores/ui-stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Divider from "./Divider.svelte";

  export let resizeIndex: number;
  export let addIndex: number;
  export let rowLength: number;
  export let rowIndex: number;
  export let columnWidth: number | undefined = undefined;
  export let isSpreadEvenly: boolean;
  export let dragging: boolean;
  export let addItems: (
    position: { row: number; column: number },
    item: CanvasComponentType[],
  ) => void;
  export let onColumnResizeStart: ((columnIndex: number) => void) | undefined =
    undefined;
  export let spreadEvenly: (rowIndex: number) => void;

  let menuOpen = false;

  $: firstElement = addIndex === 0;
  $: lastElement = addIndex === rowLength;

  $: dividerId = `row:${rowIndex}::column:${resizeIndex}`;

  $: isActiveDivider = $activeDivider === dividerId;

  $: dropId = `row:${rowIndex}::column:${addIndex}`;
  $: isDropZone = $dropZone === dropId;

  $: notActiveDivider = !isActiveDivider && !!$activeDivider;

  $: forceShowDivider = menuOpen || isActiveDivider || isDropZone;

  $: if (isActiveDivider) {
    document.body.style.cursor = "col-resize";
  } else {
    document.body.style.cursor = "";
  }

  $: addDisabled = rowLength >= 4;

  $: resizeDisabled =
    resizeIndex === -1 || rowLength >= 4 || resizeIndex === rowLength - 1;

  function onItemClick(type: CanvasComponentType) {
    if (type) {
      addItems({ row: rowIndex, column: addIndex }, [type]);
    }
  }
</script>

<div
  class="group absolute top-2 z-50 w-4"
  class:show-on-left={firstElement}
  class:show-on-right={!firstElement}
  style:height="calc(100% - 16px)"
>
  {#if !addDisabled || !isSpreadEvenly || isDropZone}
    <button
      aria-label="Resize row {rowIndex + 1} column {resizeIndex + 1}"
      disabled={resizeDisabled}
      data-width={columnWidth}
      data-row={rowIndex}
      data-column={resizeIndex}
      style:pointer-events={notActiveDivider || dragging || menuOpen
        ? "none"
        : "auto"}
      class:!opacity-100={isDropZone}
      class="peer h-full flex items-center justify-center w-4 disabled:opacity-60 disabled:cursor-default cursor-col-resize"
      on:mousedown={() => {
        if (onColumnResizeStart) onColumnResizeStart(resizeIndex);
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
      <Divider vertical show={forceShowDivider} />
    </button>

    <div
      role="presentation"
      class:nudge-right={firstElement}
      class:nudge-left={lastElement}
      class:not-sr-only={menuOpen}
      class="sr-only peer-hover:not-sr-only peer-active:sr-only hover:not-sr-only !overflow-hidden flex flex-col pointer-events-auto shadow-sm !absolute -translate-x-1/2 left-1/2 top-1/2 w-fit z-20 bg-surface -translate-y-1/2 border rounded-sm"
    >
      <AddComponentDropdown
        {rowIndex}
        columnIndex={addIndex}
        {onItemClick}
        onOpenChange={(isOpen) => {
          if (!isOpen) {
            activeDivider.set(null);
          } else {
            activeDivider.set(dividerId);
          }
        }}
        bind:open={menuOpen}
        disabled={rowLength >= 4}
      />

      {#if !isSpreadEvenly}
        <Tooltip distance={8} location="bottom">
          <button
            class="h-7 px-1 grid place-content-center border-t hover:bg-gray-50 active:bg-gray-100 text-muted-foreground"
            on:click={(e) => {
              e.stopPropagation();
              e.preventDefault();
              spreadEvenly(rowIndex);
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
</div>

<style lang="postcss">
  .show-on-left {
    @apply left-0 -translate-x-1/2;
  }

  .show-on-right {
    @apply right-0 translate-x-1/2;
  }

  .nudge-right {
    @apply translate-x-0 left-0;
  }

  .nudge-left {
    @apply left-0;
  }
</style>
