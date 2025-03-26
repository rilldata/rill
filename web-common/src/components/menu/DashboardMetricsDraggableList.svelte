<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import type {
    MetricsViewSpecMeasureV2,
    MetricsViewSpecDimensionV2,
  } from "@rilldata/web-common/runtime-client";
  import { Button } from "../button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import EyeIcon from "@rilldata/web-common/components/icons/Eye.svelte";
  import Search from "../search/Search.svelte";
  import { Tooltip } from "bits-ui";
  import { onMount, onDestroy } from "svelte";

  const UPPER_BOUND = 12 + 28 + 25;
  const ITEM_HEIGHT = 28;
  const WINDOW_OFFSET = 100; // Space to leave from window edges

  type SelectableItem = MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2;

  export let selectedItems: string[];
  export let allItems: SelectableItem[] = [];
  export let type: "measure" | "dimension" = "measure";
  export let onSelectedChange: (items: string[]) => void;

  let searchText = "";
  let active = false;
  let initialMousePosition = 0;
  let contentRect = new DOMRectReadOnly();
  let dragContainer: HTMLDivElement;
  let dropIndex: number | null = null;
  let clone: HTMLElement | null = null;
  let dragId: string | null = null; // {all|visible}-{measure|dimension}-{measure-name}
  let dragIndex = -1;
  let dragItemInitialTop = 0;
  let threshold = 0;
  let windowHeight = typeof window !== "undefined" ? window.innerHeight : 800;
  let maxContentHeight = windowHeight - WINDOW_OFFSET;
  let resizeHandler: () => void;

  onMount(() => {
    if (typeof window !== "undefined") {
      resizeHandler = () => {
        windowHeight = window.innerHeight;
        maxContentHeight = windowHeight - WINDOW_OFFSET;
      };
      window.addEventListener("resize", resizeHandler);
    }
  });

  onDestroy(() => {
    if (typeof window !== "undefined" && resizeHandler) {
      window.removeEventListener("resize", resizeHandler);
    }
  });

  $: ({ height } = contentRect);

  $: lowerBound = height - ITEM_HEIGHT - 6;

  $: allItemsMap = new Map(allItems.map((item) => [item.name, item]));

  $: numAvailable = allItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;

  $: numShownString =
    numAvailable === numShown ? "All" : `${numShown} of ${numAvailable}`;

  $: tooltipText = `Choose ${type === "measure" ? "measures" : "dimensions"} to display`;

  // Filter items based on search text
  $: filteredSelectedItems = searchText
    ? selectedItems.filter((id) => {
        const item = allItemsMap.get(id);
        return (
          item?.displayName?.toLowerCase().includes(searchText.toLowerCase()) ??
          false
        );
      })
    : selectedItems;

  $: filteredHiddenItems = searchText
    ? Array.from(allItemsMap.entries()).filter(
        ([id, item]) =>
          id &&
          !selectedItems.includes(id) &&
          (item.displayName?.toLowerCase().includes(searchText.toLowerCase()) ??
            false),
      )
    : Array.from(allItemsMap.entries()).filter(
        ([id]) => id && !selectedItems.includes(id),
      );

  function handleMouseDown(e: MouseEvent) {
    e.preventDefault();

    if (e.button !== 0) return;

    // If only one item is selected, don't allow dragging
    if (selectedItems.length === 1) return;

    const dragItem = e.target;

    // Find the closest parent element with drag data
    const dragElement =
      dragItem instanceof HTMLElement
        ? dragItem
        : dragItem instanceof Element
          ? (dragItem.closest("[data-item-name]") as HTMLElement)
          : null;

    if (!dragElement) return;

    dragId = dragElement.id ?? null;

    const { index, itemName } = dragElement.dataset;

    if (!itemName || index === undefined || index === null || index === "")
      return;

    clone = dragElement.cloneNode(true) as HTMLElement;

    const rect = dragContainer.getBoundingClientRect();
    dragItemInitialTop = dragElement.getBoundingClientRect().top - rect.top;

    if (+index > selectedItems.length - 1) {
      threshold = e.clientY - rect.bottom + ITEM_HEIGHT;
      dropIndex = null;
      dragIndex = selectedItems.length;
    } else {
      threshold = 0;
      dragIndex = +index;
      dropIndex = dragIndex;
    }

    clone.style.top = dragItemInitialTop + "px";
    clone.style.width = dragElement.clientWidth + "px";
    clone.style.left = "6px";

    clone.classList.add(
      "bg-slate-100",
      "cursor-grabbing",
      "shadow-md",
      "outline",
      "outline-gray-300",
      "outline-1",
      "opacity-60",
      "transition-all",
      "duration-150",
      "ease-in-out",
    );

    clone.style.position = "absolute";
    dragContainer.appendChild(clone);

    initialMousePosition = e.clientY;

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener(
      "mouseup",
      () => {
        if (dropIndex !== null && dragIndex > selectedItems.length - 1) {
          selectedItems.splice(dropIndex, 0, itemName);
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
        <strong
          >{`${numShownString} ${type === "measure" ? "Measures" : "Dimensions"}`}</strong
        >
        <span class="transition-transform" class:-rotate-180={active}>
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
      style="max-height: {maxContentHeight}px"
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
        class="flex flex-col pointer-events-none p-1.5 pt-0 overflow-y-auto"
        style="max-height: calc({maxContentHeight}px / 2)"
        on:mousedown={handleMouseDown}
      >
        <header
          class="flex w-full py-1.5 pb-1 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b z-10"
        >
          <h3 class="uppercase text-gray-500 font-semibold">
            Shown {type === "measure" ? "Measures" : "Dimensions"}
          </h3>
          {#if selectedItems.length > 1}
            <button
              class="text-primary-500 pointer-events-auto hover:text-primary-600 font-medium text-[11px]"
              on:click={() => {
                selectedItems = [selectedItems[0]];
                onSelectedChange(selectedItems);
              }}
            >
              Hide all
            </button>
          {/if}
        </header>
        {#if filteredSelectedItems.length === 0}
          <div class="px-2 py-2 text-xs text-gray-500">
            {searchText
              ? `No matching ${type === "measure" ? "measures" : "dimensions"} shown`
              : `No ${type === "measure" ? "measures" : "dimensions"} shown`}
          </div>
        {:else}
          {#each filteredSelectedItems as id, i (i)}
            {@const elementId = `visible-${type === "measure" ? "measures" : "dimensions"}-${id}`}
            {@const isDragItem = dragId === elementId}
            {#if allItemsMap.get(id)?.description || selectedItems.length === 1}
              <Tooltip.Root openDelay={200} portal="body">
                <Tooltip.Trigger>
                  <div
                    role="presentation"
                    data-index={i}
                    data-item-name={id}
                    id={elementId}
                    class:sr-only={isDragItem}
                    class:transition-margin={dragIndex !== -1 &&
                      dropIndex !== dragIndex}
                    class:drag-transition={dragIndex !== -1}
                    class:mt-7={dropIndex !== null &&
                      !isDragItem &&
                      i === dropIndex + (i > dragIndex ? 1 : 0)}
                    class:mb-7={dropIndex === selectedItems.length - 1 &&
                      i ===
                        selectedItems.length -
                          1 -
                          (dragIndex === selectedItems.length - 1 ? 1 : 0)}
                    style:pointer-events={isDragItem ? "none" : "auto"}
                    style:height="{ITEM_HEIGHT}px"
                    class="w-full flex gap-x-1 flex-none px-2 py-1 pointer-events-auto cursor-grab items-center hover:bg-slate-50 rounded-sm"
                    class:cursor-not-allowed={selectedItems.length === 1}
                  >
                    <DragHandle size="16px" className="text-gray-400" />

                    {allItemsMap.get(id)?.displayName ??
                      `Unknown ${type === "measure" ? "measure" : "dimension"}`}

                    <button
                      class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                      on:click={() => {
                        selectedItems = selectedItems.filter((_, j) => j !== i);
                        onSelectedChange(selectedItems);
                      }}
                      on:mousedown|stopPropagation={() => {
                        // NO-OP
                      }}
                      disabled={selectedItems.length === 1}
                      class:pointer-events-none={selectedItems.length === 1}
                      class:opacity-50={selectedItems.length === 1}
                    >
                      <EyeIcon size="14px" color="#6b7280" />
                    </button>
                  </div>
                </Tooltip.Trigger>

                <Tooltip.Content side="right" sideOffset={12}>
                  <div
                    class="bg-gray-800 text-gray-50 rounded p-2 pt-1 pb-1 shadow-md pointer-events-none z-50"
                  >
                    {#if selectedItems.length === 1}
                      Must show at least one {type === "measure"
                        ? "measure"
                        : "dimension"}
                    {:else}
                      {allItemsMap.get(id)?.description}
                    {/if}
                  </div>
                </Tooltip.Content>
              </Tooltip.Root>
            {:else}
              <!-- FIXME: hoist to DraggableListItem -->
              <div
                role="presentation"
                data-index={i}
                data-item-name={id}
                id={elementId}
                class:sr-only={isDragItem}
                class:transition-margin={dragIndex !== -1 &&
                  dropIndex !== dragIndex}
                class:drag-transition={dragIndex !== -1}
                class:mt-7={dropIndex !== null &&
                  !isDragItem &&
                  i === dropIndex + (i > dragIndex ? 1 : 0)}
                class:mb-7={dropIndex === selectedItems.length - 1 &&
                  i ===
                    selectedItems.length -
                      1 -
                      (dragIndex === selectedItems.length - 1 ? 1 : 0)}
                style:pointer-events={isDragItem ? "none" : "auto"}
                style:height="{ITEM_HEIGHT}px"
                class="w-full flex gap-x-1 flex-none px-2 py-1 pointer-events-auto cursor-grab items-center hover:bg-slate-50 rounded-sm"
                class:cursor-not-allowed={selectedItems.length === 1}
              >
                <DragHandle size="16px" className="text-gray-400" />

                {allItemsMap.get(id)?.displayName ??
                  `Unknown ${type === "measure" ? "measure" : "dimension"}`}

                <button
                  class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                  on:click={() => {
                    selectedItems = selectedItems.filter((_, j) => j !== i);
                    onSelectedChange(selectedItems);
                  }}
                  on:mousedown|stopPropagation={() => {
                    // NO-OP
                  }}
                  disabled={selectedItems.length === 1}
                  class:pointer-events-none={selectedItems.length === 1}
                  class:opacity-50={selectedItems.length === 1}
                >
                  <EyeIcon size="14px" color="#6b7280" />
                </button>
              </div>
            {/if}
          {/each}
        {/if}
      </div>
      {#if selectedItems.length < allItems.length}
        <span class="h-px bg-slate-200 w-full" />
        <div
          class="flex flex-col overflow-y-auto p-1.5 pt-0"
          style="max-height: calc({maxContentHeight}px / 2)"
        >
          <header
            class="flex py-1.5 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b"
          >
            <h3
              class="uppercase text-xs text-gray-500 font-semibold from-white from-80% to-transparent bg-gradient-to-b"
            >
              Hidden {type === "measure" ? "Measures" : "Dimensions"}
            </h3>
            <button
              class="pointer-events-auto text-primary-500 text-[11px] font-medium"
              on:click={() => {
                selectedItems = allItems.map((item) => item.name ?? "");
                onSelectedChange(selectedItems);
              }}
            >
              Show all
            </button>
          </header>

          {#if filteredHiddenItems.length === 0}
            <div class="px-2 py-2 text-xs text-gray-500">
              {searchText
                ? `No matching hidden ${type === "measure" ? "measures" : "dimensions"}`
                : `No hidden ${type === "measure" ? "measures" : "dimensions"}`}
            </div>
          {:else}
            {#each filteredHiddenItems as [id = "", item], i (i)}
              {@const elementId = `all-${type === "measure" ? "measures" : "dimensions"}-${id}`}
              {@const isDragItem = dragId === elementId}
              <div
                data-index={i + selectedItems.length - 1}
                id={elementId}
                data-item-name={id}
                class:z-50={isDragItem}
                class:opacity-0={isDragItem}
                style:height="{ITEM_HEIGHT}px"
                class="w-full flex gap-x-1 px-2 py-1 justify-between pointer-events-auto items-center p-1 rounded-sm"
              >
                {item.displayName}

                <button
                  class="hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                  on:click={() => {
                    selectedItems = [...selectedItems, id];
                    onSelectedChange(selectedItems);
                  }}
                >
                  <EyeOffIcon size="14px" color="#9ca3af" />
                </button>
              </div>
            {/each}
          {/if}
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
    @apply text-[11px] text-gray-500;
  }
</style>
