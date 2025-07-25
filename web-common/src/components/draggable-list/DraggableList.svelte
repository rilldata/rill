<script lang="ts">
  import { clamp } from "@rilldata/web-common/lib/clamp";

  const ITEM_HEIGHT = 28;
  const UPPER_BOUND = 12;
  const THROTTLE_MS = 16; // ~60fps

  type DraggableItem = {
    id: string;
    [key: string]: any;
  };

  export let items: DraggableItem[] = [];
  export let searchValue: string = "";
  export let showSearch: boolean = false;
  export let minHeight: string = "100px";
  export let maxHeight: string = "400px";
  export let onReorder:
    | ((data: {
        items: DraggableItem[];
        fromIndex: number;
        toIndex: number;
      }) => void)
    | undefined = undefined;
  export let onItemClick:
    | ((data: { item: DraggableItem; index: number }) => void)
    | undefined = undefined;

  let initialMousePosition = 0;
  let contentRect = new DOMRectReadOnly();
  let dragContainer: HTMLDivElement;
  let dropIndex: number | null = null;
  let clone: HTMLElement | null = null;
  let dragId: string | null = null;
  let dragIndex = -1;
  let dragItemInitialTop = 0;
  let lastUpdateTime = 0;

  $: ({ height } = contentRect);
  $: lowerBound = Math.max(
    height - ITEM_HEIGHT - 6,
    UPPER_BOUND + (items.length - 1) * ITEM_HEIGHT,
  );

  $: filteredItems = searchValue
    ? items.filter((item) =>
        item.id.toLowerCase().includes(searchValue.toLowerCase()),
      )
    : items;

  function handleMouseDown(e: MouseEvent) {
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
      "bg-slate-100",
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

  function handleItemClick(item: DraggableItem, index: number) {
    onItemClick?.({ item, index });
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
      <slot name="search" {searchValue}>
        <input
          bind:value={searchValue}
          placeholder="Search..."
          class="w-full px-2 py-1 border border-gray-300 rounded text-sm"
        />
      </slot>
    </div>
  {/if}

  <div class="flex-1 overflow-y-auto">
    <slot name="header" items={filteredItems}>
      <!-- Optional header slot -->
    </slot>

    <div
      role="presentation"
      class="flex flex-col p-1.5"
      on:mousedown={handleMouseDown}
    >
      {#if filteredItems.length === 0}
        <div class="px-2 py-2 text-xs text-gray-500">
          <slot name="empty" {searchValue}>
            {searchValue ? "No matching items" : "No items"}
          </slot>
        </div>
      {:else}
        {#each filteredItems as item, i (item.id)}
          {@const isDragItem = dragId === item.id}
          {@const isDropTarget =
            dropIndex !== null &&
            !isDragItem &&
            i === dropIndex + (i > dragIndex ? 1 : 0)}
          {@const isLastItem =
            dropIndex === items.length - 1 &&
            i === items.length - 1 - (dragIndex === items.length - 1 ? 1 : 0)}
          <div
            data-drag-item
            data-index={i}
            data-item-id={item.id}
            class:sr-only={isDragItem}
            class:transition-margin={dragIndex !== -1 &&
              dropIndex !== dragIndex}
            class:drag-transition={dragIndex !== -1}
            class:mt-7={isDropTarget}
            class:mb-7={isLastItem}
            style:pointer-events={isDragItem ? "none" : "auto"}
            style:height="{ITEM_HEIGHT}px"
            class="w-full flex gap-x-1 flex-none py-1 pointer-events-auto cursor-grab items-center hover:bg-slate-50 rounded-sm"
            class:cursor-not-allowed={items.length === 1}
            on:click={() => handleItemClick(item, i)}
            on:keydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleItemClick(item, i);
              }
            }}
            role="button"
            tabindex="0"
          >
            <slot name="item" {item} index={i} {isDragItem}>
              <span class="truncate flex-1 text-left pointer-events-none">
                {item.id}
              </span>
            </slot>
          </div>
        {/each}
      {/if}
    </div>

    <slot name="footer" items={filteredItems}>
      <!-- Optional footer slot -->
    </slot>
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
