<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { Button } from "../button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import EyeIcon from "@rilldata/web-common/components/icons/Eye.svelte";
  import Search from "../search/Search.svelte";

  const UPPER_BOUND = 12 + 28 + 25;
  const ITEM_HEIGHT = 28;

  export let selectedItems: string[];
  export let allItems: MetricsViewSpecMeasureV2[] = [];
  export let onSelectedChange: (items: string[]) => void;
  export let disabled = false;
  export let category: string = "Measures";

  let searchText = "";
  let active = false;
  let initialMousePosition = 0;
  let contentRect = new DOMRectReadOnly();
  let dragContainer: HTMLDivElement;
  let dropIndex: number | null = null;
  let clone: HTMLElement | null = null;
  let dragId: string | null = null; // {all|visible}-measures-{measure-name}
  let dragIndex = -1;
  let dragItemInitialTop = 0;
  let threshold = 0;

  $: ({ height } = contentRect);

  $: lowerBound = height - ITEM_HEIGHT - 6;

  $: allItemsMap = new Map(allItems.map((item) => [item.name, item]));

  $: numAvailable = allItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;

  $: numShownString =
    numAvailable === numShown ? "All" : `${numShown} of ${numAvailable}`;

  $: tooltipText = `Choose ${category.toLowerCase()} to display`;

  function handleMouseDown(e: MouseEvent) {
    e.preventDefault();

    if (e.button !== 0) return;

    const dragItem = e.target;

    if (!(dragItem instanceof HTMLElement)) return;

    dragId = dragItem?.id ?? null;

    const { index, measureName } = dragItem.dataset;

    if (!measureName || index === undefined || index === null || index === "")
      return;

    clone = dragItem.cloneNode(true) as HTMLElement;

    if (+index > selectedItems.length - 1) {
      const rect = dragContainer.getBoundingClientRect();
      dragItemInitialTop = dragItem.getBoundingClientRect().top - rect.top;

      threshold = e.clientY - rect.bottom + ITEM_HEIGHT;
      dropIndex = null;
      dragIndex = selectedItems.length;
    } else {
      dragItemInitialTop = dragItem.offsetTop;
      threshold = 0;
      dragIndex = +index;
      dropIndex = dragIndex;
    }

    clone.style.top = dragItemInitialTop + "px";
    clone.style.width = dragItem.clientWidth + "px";
    clone.style.left = "6px";

    clone.classList.add(
      "bg-slate-100",
      "cursor-grabbing",
      "shadow-md",
      "outline",
      "outline-gray-300",
      "outline-1",
    );

    clone.style.position = "absolute";
    dragContainer.appendChild(clone);

    initialMousePosition = e.clientY;

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener(
      "mouseup",
      () => {
        if (dropIndex !== null && dragIndex > selectedItems.length - 1) {
          selectedItems.splice(dropIndex, 0, measureName);
          selectedItems = selectedItems;
        } else if (dropIndex !== null && dropIndex > selectedItems.length) {
          selectedItems = selectedItems.filter((_, i) => i !== dragIndex);
        } else if (dropIndex !== null && dragIndex !== dropIndex) {
          selectedItems = reorderItems(selectedItems, dragIndex, dropIndex);
        }

        onSelectedChange(selectedItems);
        dragIndex = -1;
        dragId = null;
        dropIndex = null;

        clone?.remove();

        window.removeEventListener("mousemove", handleMouseMove);
      },
      { once: true },
    );
  }

  function handleMouseMove(e: MouseEvent) {
    e.preventDefault();

    if (!clone) return;

    const delta = e.clientY - initialMousePosition;

    const newPxValue = dragItemInitialTop + delta;

    clone.style.top = clamp(UPPER_BOUND, newPxValue, lowerBound) + "px";

    if (threshold && delta > threshold) {
      dropIndex = null;
      return;
    } else {
      const newIndex = Math.round((delta - threshold) / ITEM_HEIGHT);

      dropIndex = Math.max(0, newIndex + dragIndex);
    }
  }

  function reorderItems<T>(items: T[], from: number, to: number | null): T[] {
    if (to === null) return items;
    const result = Array.from(items);
    const [removed] = result.splice(from, 1);
    result.splice(to, 0, removed);
    return result;
  }
</script>

<Popover.Root bind:open={active}>
  <Popover.Trigger asChild let:builder>
    <Button builders={[builder]} type="text" label={tooltipText} on:click>
      <div
        class="flex items-center gap-x-0.5 px-1 text-gray-700 hover:text-inherit"
      >
        <strong>{`${numShownString} ${category}`}</strong>
        <span
          class="transition-transform"
          class:hidden={disabled}
          class:-rotate-180={active}
        >
          <CaretDownIcon />
        </span>
      </div>
    </Button>
  </Popover.Trigger>
  <Popover.Content class="p-0" align="start">
    <div
      bind:this={dragContainer}
      bind:contentRect
      class="flex flex-col relative"
      role="presentation"
      on:mousedown={handleMouseDown}
    >
      <div class="px-3 pt-3 pb-0">
        <Search
          bind:value={searchText}
          label="Search list"
          showBorderOnFocus={false}
        />
      </div>

      <div
        role="presentation"
        class="flex flex-col pointer-events-none p-1.5 pt-0"
      >
        <header
          class="flex w-full pt-1.5 pb-1 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b"
        >
          <h3 class="uppercase text-gray-500 font-semibold">Shown Measures</h3>
          {#if selectedItems.length > 1}
            <button
              class="text-primary-500 pointer-events-auto hover:text-primary-600 font-medium text-[10px]"
              on:click={() => {
                selectedItems = [selectedItems[0]];
                onSelectedChange(selectedItems);
              }}
            >
              Hide all
            </button>
          {/if}
        </header>
        {#each selectedItems as id, i (i)}
          {@const elementId = `visible-measures-${id}`}
          {@const isDragItem = dragId === elementId}
          <div
            role="presentation"
            data-index={i}
            data-measure-name={id}
            id={elementId}
            class:sr-only={isDragItem}
            class:transition-margin={dragIndex !== -1 &&
              dropIndex !== dragIndex}
            class:mt-7={dropIndex !== null &&
              !isDragItem &&
              i === dropIndex + (i > dragIndex ? 1 : 0)}
            class:mb-7={dropIndex === selectedItems.length - 1 &&
              i ===
                selectedItems.length -
                  1 -
                  (dragIndex === selectedItems.length - 1 ? 1 : 0)}
            style:pointer-events={isDragItem || selectedItems.length === 1
              ? "none"
              : "auto"}
            style:height="{ITEM_HEIGHT}px"
            class="w-full flex gap-x-1 flex-none px-1 pointer-events-auto cursor-grab items-center p-1 hover:bg-slate-100 rounded-sm"
          >
            <DragHandle size="16px" className="text-gray-400" />

            {allItemsMap.get(id)?.displayName ?? "Unknown measure"}
            {#if selectedItems.length > 1}
              <button
                class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                on:click={() => {
                  selectedItems = selectedItems.filter((_, j) => j !== i);
                  onSelectedChange(selectedItems);
                }}
              >
                <EyeIcon size="14px" />
              </button>
            {/if}
          </div>
        {/each}
      </div>
      {#if selectedItems.length < allItems.length}
        <span class="h-px bg-slate-200 w-full" />
        <div class="flex flex-col max-h-52 overflow-y-auto p-1.5 pt-0">
          <header
            class="flex pt-1.5 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b"
          >
            <h3
              class="uppercase text-xs text-gray-500 font-semibold from-white from-80% to-transparent bg-gradient-to-b"
            >
              Hidden Measures
            </h3>
            <button
              class="pointer-events-auto text-primary-500 text-[10px] font-medium"
              on:click={() => {
                selectedItems = allItems.map((item) => item.name ?? "");
                onSelectedChange(selectedItems);
              }}
            >
              Show all
            </button>
          </header>

          {#each allItemsMap as [id = "", measure], i (i)}
            {@const elementId = `all-measures-${id}`}
            {@const isDragItem = dragId === elementId}
            {#if !selectedItems.includes(id)}
              <div
                data-index={i + selectedItems.length - 1}
                id={elementId}
                data-measure-name={id}
                class:z-50={isDragItem}
                class:opacity-0={isDragItem}
                style:height="{ITEM_HEIGHT}px"
                class="w-full flex gap-x-1 px-2 pr-1 justify-between pointer-events-auto cursor-grab items-center p-1 hover:bg-slate-100 rounded-sm"
              >
                {measure.displayName}

                <button
                  class="hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                  on:click={() => {
                    selectedItems = [...selectedItems, id];
                    onSelectedChange(selectedItems);
                  }}
                >
                  <EyeOffIcon size="14px" />
                </button>
              </div>
            {/if}
          {/each}
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>

<style lang="postcss">
  .transition-margin {
    transition-property: margin-top, margin-bottom;
    transition-duration: 100ms;
  }

  h3 {
    @apply text-[10px] text-gray-500;
  }
</style>
