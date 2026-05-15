<script lang="ts">
  import type { Snippet } from "svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";

  const ITEM_HEIGHT = 28;
  const UPPER_BOUND = 12;
  const THROTTLE_MS = 16; // ~60fps

  type DraggableItem = {
    id: string;
    [key: string]: any;
  };

  let {
    items = [] as DraggableItem[],
    searchValue = $bindable(""),
    showSearch = false,
    minHeight = "100px",
    maxHeight = "400px",
    onReorder,
    onItemClick,
    draggable = true,
    header,
    footer,
    empty,
    item: itemSnippet,
    search: searchSnippet,
  }: {
    items?: DraggableItem[];
    searchValue?: string;
    showSearch?: boolean;
    minHeight?: string;
    maxHeight?: string;
    onReorder?: (data: {
      items: DraggableItem[];
      fromIndex: number;
      toIndex: number;
    }) => void;
    onItemClick?: (data: { item: DraggableItem; index: number }) => void;
    draggable?: boolean;
    header?: Snippet<[{ items: DraggableItem[] }]>;
    footer?: Snippet<[{ items: DraggableItem[] }]>;
    empty?: Snippet<[{ searchValue: string }]>;
    item?: Snippet<
      [{ item: DraggableItem; index: number; isDragItem: boolean }]
    >;
    search?: Snippet<[{ searchValue: string }]>;
  } = $props();

  let initialMousePosition = 0;
  let contentRect = $state(new DOMRectReadOnly());
  let dragContainer: HTMLDivElement;
  let dropIndex: number | null = $state(null);
  let clone: HTMLElement | null = null;
  let dragId: string | null = $state(null);
  let dragIndex = $state(-1);
  let dragItemInitialTop = 0;
  let lastUpdateTime = 0;

  let height = $derived(contentRect.height);
  let lowerBound = $derived(
    Math.max(
      height - ITEM_HEIGHT - 6,
      UPPER_BOUND + (items.length - 1) * ITEM_HEIGHT,
    ),
  );

  let filteredItems = $derived(
    searchValue
      ? items.filter((filterItem) => {
          const normalizedSearch = searchValue.trim().toLowerCase();
          if (normalizedSearch === "") return true;
          const itemId = filterItem.id.toLowerCase();
          const itemDisplayName =
            (filterItem.displayName as string | undefined)?.toLowerCase() ?? "";
          return (
            itemId.includes(normalizedSearch) ||
            itemDisplayName.includes(normalizedSearch)
          );
        })
      : items,
  );

  function handleMouseDown(e: MouseEvent) {
    if (!draggable) return;
    e.preventDefault();

    if (e.button !== 0) return;

    // If only one item, don't allow dragging
    if (items.length === 1) return;

    const dragElement =
      e.target instanceof HTMLElement
        ? (e.target.closest("[data-drag-item]") as HTMLElement)
        : null;

    if (!dragElement) return;

    const { index, itemId } = dragElement.dataset;
    if (!itemId || index === undefined) return;

    dragId = itemId;
    dragIndex = parseInt(index);
    dropIndex = dragIndex;

    clone = dragElement.cloneNode(true) as HTMLElement;

    const rect = dragContainer.getBoundingClientRect();
    dragItemInitialTop = dragElement.getBoundingClientRect().top - rect.top;

    clone.style.transform = `translateY(${dragItemInitialTop}px)`;
    clone.style.width = dragElement.clientWidth + "px";
    clone.style.left = "6px";
    clone.style.position = "absolute";
    clone.style.zIndex = "50";

    clone.classList.add(
      "bg-gray-100",
      "cursor-grabbing",
      "shadow-md",
      "outline",
      "outline-gray-300",
      "outline-1",
      "opacity-60",
      "will-change-transform",
    );

    dragContainer.appendChild(clone);

    initialMousePosition = e.clientY;
    lastUpdateTime = performance.now();

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp, { once: true });
  }

  function handleMouseMove(e: MouseEvent) {
    e.preventDefault();

    if (!clone) return;

    const now = performance.now();
    if (now - lastUpdateTime < THROTTLE_MS) return;
    lastUpdateTime = now;

    const delta = e.clientY - initialMousePosition;
    const newPxValue = dragItemInitialTop + delta;

    clone.style.transform = `translateY(${clamp(UPPER_BOUND, newPxValue, lowerBound)}px)`;

    const newIndex = Math.round(delta / ITEM_HEIGHT);
    dropIndex = Math.max(0, Math.min(items.length - 1, newIndex + dragIndex));
  }

  function handleMouseUp() {
    if (dropIndex !== null && dragIndex !== dropIndex) {
      const reorderedItems = reorderItems(items, dragIndex, dropIndex);
      onReorder?.({
        items: reorderedItems,
        fromIndex: dragIndex,
        toIndex: dropIndex,
      });
    }

    cleanup();
  }

  function cleanup() {
    dragIndex = -1;
    dragId = null;
    dropIndex = null;
    clone?.remove();
    clone = null;
    window.removeEventListener("mousemove", handleMouseMove);
  }

  function reorderItems<T>(items: T[], from: number, to: number): T[] {
    const result = Array.from(items);
    const [removed] = result.splice(from, 1);
    result.splice(to, 0, removed);
    return result;
  }

  function handleItemClick(clickedItem: DraggableItem, index: number) {
    onItemClick?.({ item: clickedItem, index });
  }
</script>

<div
  bind:this={dragContainer}
  bind:contentRect
  class="flex flex-col relative overflow-x-hidden"
  style:min-height={minHeight}
  style:max-height={maxHeight}
  role="presentation"
>
  {#if showSearch}
    <div class="px-3 pt-3 pb-1">
      {#if searchSnippet}
        {@render searchSnippet({ searchValue })}
      {:else}
        <input
          bind:value={searchValue}
          placeholder="Search..."
          class="w-full px-2 py-1 border rounded text-sm"
        />
      {/if}
    </div>
  {/if}

  <div class="flex-1 overflow-y-auto">
    {#if header}
      {@render header({ items: filteredItems })}
    {/if}

    <div
      role="presentation"
      class="flex flex-col p-1.5"
      onmousedown={handleMouseDown}
    >
      {#if filteredItems.length === 0}
        <div class="px-2 py-2 text-xs text-fg-secondary">
          {#if empty}
            {@render empty({ searchValue })}
          {:else}
            {searchValue ? "No matching items" : "No items"}
          {/if}
        </div>
      {:else}
        {#each filteredItems as currentItem, i (currentItem.id)}
          {@const isDragItem = dragId === currentItem.id}
          {@const isDropTarget =
            dropIndex !== null &&
            !isDragItem &&
            i === dropIndex + (i > dragIndex ? 1 : 0)}
          {@const isLastItem =
            dropIndex === items.length - 1 &&
            i === items.length - 1 - (dragIndex === items.length - 1 ? 1 : 0)}
          {#if onItemClick}
            <button
              type="button"
              data-drag-item
              data-index={i}
              data-item-id={currentItem.id}
              class:sr-only={isDragItem}
              class:transition-margin={dragIndex !== -1 &&
                dropIndex !== dragIndex}
              class:drag-transition={dragIndex !== -1}
              class:mt-7={isDropTarget}
              class:mb-7={isLastItem}
              style:pointer-events={isDragItem ? "none" : "auto"}
              style:height="{ITEM_HEIGHT}px"
              class="w-full flex gap-x-1 flex-none py-1 pointer-events-auto items-center hover:bg-surface-background rounded-sm text-left"
              class:cursor-grab={draggable}
              class:cursor-not-allowed={draggable && items.length === 1}
              class:cursor-pointer={!draggable && !!onItemClick}
              class:cursor-default={!draggable && !onItemClick}
              onclick={() => handleItemClick(currentItem, i)}
            >
              {#if itemSnippet}
                {@render itemSnippet({
                  item: currentItem,
                  index: i,
                  isDragItem,
                })}
              {:else}
                <span class="truncate flex-1 text-left pointer-events-none">
                  {currentItem.id}
                </span>
              {/if}
            </button>
          {:else}
            <div
              data-drag-item
              data-index={i}
              data-item-id={currentItem.id}
              class:sr-only={isDragItem}
              class:transition-margin={dragIndex !== -1 &&
                dropIndex !== dragIndex}
              class:drag-transition={dragIndex !== -1}
              class:mt-7={isDropTarget}
              class:mb-7={isLastItem}
              style:pointer-events={isDragItem ? "none" : "auto"}
              style:height="{ITEM_HEIGHT}px"
              class="w-full flex gap-x-1 flex-none py-1 pointer-events-auto items-center text-fg-primary hover:bg-popover-accent rounded-sm"
              class:cursor-grab={draggable}
              class:cursor-not-allowed={draggable && items.length === 1}
              class:cursor-pointer={!draggable && !!onItemClick}
              class:cursor-default={!draggable && !onItemClick}
            >
              {#if itemSnippet}
                {@render itemSnippet({
                  item: currentItem,
                  index: i,
                  isDragItem,
                })}
              {:else}
                <span class="truncate flex-1 text-left pointer-events-none">
                  {currentItem.id}
                </span>
              {/if}
            </div>
          {/if}
        {/each}
      {/if}
    </div>

    {#if footer}
      {@render footer({ items: filteredItems })}
    {/if}
  </div>
</div>

<style lang="postcss">
  .transition-margin {
    transition-property: margin-top, margin-bottom;
    transition-duration: 100ms;
    will-change: margin-top, margin-bottom;
  }

  .drag-transition {
    transition: none;
  }
</style>
