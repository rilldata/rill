<script lang="ts">
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";

  export let items: any[] = [];
  export let onReorder: (items: any[]) => void;
  export let itemHeight = 28;
  export let throttleMs = 16;
  export let upperBound = 12 + 28 + 25;
  export let disabled = false;

  let dragContainer: HTMLDivElement;
  let contentRect = new DOMRectReadOnly();
  let initialMousePosition = 0;
  let dragIndex = -1;
  let dropIndex: number | null = null;
  let clone: HTMLElement | null = null;
  let dragId: string | null = null;
  let dragItemInitialTop = 0;
  let threshold = 0;
  let lastUpdateTime = 0;

  $: ({ height } = contentRect);
  $: lowerBound = Math.min(
    height - itemHeight - 6,
    upperBound + (items.length - 1) * itemHeight,
  );

  function handleMouseDown(e: MouseEvent) {
    e.preventDefault();

    if (e.button !== 0) return;

    // If only one item is selected, don't allow dragging
    if (items.length === 1) return;

    const dragItem = e.target;

    // Find the closest parent element with drag data
    const dragElement =
      dragItem instanceof HTMLElement
        ? dragItem
        : dragItem instanceof Element
          ? (dragItem.closest("[data-item-index]") as HTMLElement)
          : null;

    if (!dragElement) return;

    dragId = dragElement.id ?? null;

    const { itemIndex } = dragElement.dataset;

    if (itemIndex === undefined || itemIndex === null || itemIndex === "")
      return;

    clone = dragElement.cloneNode(true) as HTMLElement;

    const rect = dragContainer.getBoundingClientRect();
    dragItemInitialTop = dragElement.getBoundingClientRect().top - rect.top;

    if (+itemIndex > items.length - 1) {
      threshold = e.clientY - rect.bottom + itemHeight;
      dropIndex = null;
      dragIndex = items.length;
    } else {
      threshold = 0;
      dragIndex = +itemIndex;
      dropIndex = dragIndex;
    }

    clone.style.transform = `translateY(${dragItemInitialTop}px)`;
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
      "will-change-transform",
    );

    clone.style.position = "absolute";
    dragContainer.appendChild(clone);

    initialMousePosition = e.clientY;
    lastUpdateTime = performance.now();

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener(
      "mouseup",
      () => {
        if (dropIndex !== null && dragIndex !== dropIndex) {
          const newItems = reorderItems(items, dragIndex, dropIndex);
          onReorder(newItems);
        }

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

    const now = performance.now();
    if (now - lastUpdateTime < throttleMs) return;
    lastUpdateTime = now;

    const delta = e.clientY - initialMousePosition;
    const newPxValue = dragItemInitialTop + delta;

    clone.style.transform = `translateY(${clamp(upperBound, newPxValue, lowerBound)}px)`;

    if (threshold && delta > threshold) {
      dropIndex = null;
    } else {
      const newIndex = Math.round((delta - threshold) / itemHeight);
      dropIndex = Math.max(0, Math.min(items.length - 1, newIndex + dragIndex));
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

<div
  bind:this={dragContainer}
  bind:contentRect
  class="flex flex-col relative"
  on:mousedown={handleMouseDown}
>
  {#each items as item, i (i)}
    {@const elementId = `draggable-item-${i}`}
    {@const isDragItem = dragId === elementId}
    <div
      data-item-index={i}
      id={elementId}
      class:sr-only={isDragItem}
      class:transition-margin={dragIndex !== -1 && dropIndex !== dragIndex}
      class:drag-transition={dragIndex !== -1}
      class:mt-7={dropIndex !== null &&
        !isDragItem &&
        i === dropIndex + (i > dragIndex ? 1 : 0)}
      class:mb-7={dropIndex === items.length - 1 &&
        i === items.length - 1 - (dragIndex === items.length - 1 ? 1 : 0)}
      style:pointer-events={isDragItem ? "none" : "auto"}
      style:height="{itemHeight}px"
      class="w-full flex gap-x-1 flex-none px-2 py-1 pointer-events-auto cursor-grab items-center hover:bg-slate-50 rounded-sm"
      class:cursor-not-allowed={items.length === 1 || disabled}
    >
      <DragHandle size="16px" className="text-gray-400" />
      <slot {item} {i} />
    </div>
  {/each}
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
