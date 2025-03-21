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
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  const UPPER_BOUND = 12 + 28 + 25;
  const ITEM_HEIGHT = 28;

  type SelectableItem = MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2;

  export let selectedItems: string[];
  export let allItems: SelectableItem[] = [];
  export let onSelectedChange: (items: string[]) => void;
  export let disabled = false;
  export let type: "measure" | "dimension" = "measure";

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

    const dragItem = e.target;

    if (!(dragItem instanceof HTMLElement)) return;

    dragId = dragItem?.id ?? null;

    const { index, itemName } = dragItem.dataset;

    if (!itemName || index === undefined || index === null || index === "")
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
        class="flex flex-col pointer-events-none p-1.5 pt-0 max-h-52 overflow-y-auto"
        on:mousedown={handleMouseDown}
      >
        <header
          class="flex w-full py-1.5 pb-1 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b"
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
            <Tooltip
              distance={12}
              location="right"
              activeDelay={200}
              suppress={!allItemsMap.get(id)?.description}
            >
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
                style:pointer-events={isDragItem || selectedItems.length === 1
                  ? "none"
                  : "auto"}
                style:height="{ITEM_HEIGHT}px"
                class="w-full flex gap-x-1 flex-none px-2 py-1 pointer-events-auto cursor-grab items-center hover:bg-slate-50 rounded-sm"
              >
                <DragHandle size="16px" className="text-gray-400" />

                {allItemsMap.get(id)?.displayName ??
                  `Unknown ${type === "measure" ? "measure" : "dimension"}`}
                {#if selectedItems.length > 1}
                  <button
                    class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                    on:click={() => {
                      selectedItems = selectedItems.filter((_, j) => j !== i);
                      onSelectedChange(selectedItems);
                    }}
                  >
                    <EyeIcon size="14px" color="#6b7280" />
                  </button>
                {/if}
              </div>

              <TooltipContent slot="tooltip-content">
                {allItemsMap.get(id)?.description}
              </TooltipContent>
            </Tooltip>
          {/each}
        {/if}
      </div>
      {#if selectedItems.length < allItems.length}
        <span class="h-px bg-slate-200 w-full" />
        <div class="flex flex-col max-h-52 overflow-y-auto p-1.5 pt-0">
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
    transition-duration: 150ms;
    transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
    will-change: margin-top, margin-bottom;
  }

  .drag-transition {
    transition: transform 150ms cubic-bezier(0.4, 0, 0.2, 1);
    will-change: transform;
  }

  h3 {
    @apply text-[11px] text-gray-500;
  }
</style>
