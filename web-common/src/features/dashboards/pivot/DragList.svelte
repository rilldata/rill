<script context="module" lang="ts">
  import AddField from "./AddField.svelte";
  import PivotChip from "./PivotChip.svelte";
  import { createEventDispatcher } from "svelte";
  import { PivotChipData, PivotChipType } from "./types";
  import { writable } from "svelte/store";

  type DragEvent<T> = MouseEvent & {
    currentTarget: EventTarget & T;
  };

  export type Section = "rows" | "columns" | "Time" | "Measures" | "Dimensions";

  export type DragData = {
    chip: PivotChipData;
    source: Section;
    width: number;
    initialIndex: number;
  };

  export const dragging = writable<null | DragData>(null);
  export const controllerStore = writable<AbortController | null>(null);
</script>

<script lang="ts">
  export let items: PivotChipData[] = [];
  export let placeholder: string | null = null;
  export let type: Section;

  const isDropLocation = type === "columns" || type === "rows";

  const dispatch = createEventDispatcher();

  $: dragData = $dragging;
  $: currentlyDragging = Boolean(dragData);
  $: source = dragData?.source;
  $: dragChip = dragData?.chip;
  $: ghostWidth = dragData?.width;
  $: initialIndex = dragData?.initialIndex ?? -1;
  $: zoneStartedDrag = source === type;
  $: lastDimensionIndex = items.findIndex(
    (i) => i.type === PivotChipType.Dimension,
  );

  $: isValidDropZone =
    isDropLocation &&
    currentlyDragging &&
    (type === "columns" || dragChip?.type !== PivotChipType.Measure);

  let container: HTMLDivElement;
  let node: HTMLButtonElement;
  let offset = { x: 0, y: 0 };
  let ghostIndex: number | undefined = undefined;

  function handleMouseDown(e: MouseEvent, item: PivotChipData) {
    const dragItem = document.getElementById(item.id);
    if (!dragItem) return;

    node = dragItem.cloneNode(true) as HTMLButtonElement;

    document.body.appendChild(node);

    window.addEventListener("mousemove", trackDragItem);

    const { width, left, top } = dragItem.getBoundingClientRect();

    node.style.left = `${left}px`;
    node.style.top = `${top}px`;

    offset = {
      x: e.clientX - left,
      y: e.clientY - top,
    };

    node.classList.add(
      "shadow-lg",
      "shadow-slate-300",
      "absolute",
      "pointer-events-none",
    );

    document.body.style.cursor = "grabbing ";

    const index = Number(dragItem.dataset.index);
    initialIndex = index;
    ghostIndex = index;

    if (isDropLocation) {
      addSwapListeners(container);

      const temp = [...items];
      temp.splice(index, 1);
      items = temp;

      const controller = new AbortController();

      controllerStore.set(controller);

      window.addEventListener(
        "mouseup",
        () => {
          dispatch("update", temp);
        },
        {
          once: true,
          signal: controller.signal,
        },
      );
    }

    window.addEventListener(
      "mouseup",
      () => {
        window.removeEventListener("mousemove", trackDragItem);
        resetGlobalDrag();
      },
      {
        once: true,
      },
    );

    dragging.set({
      chip: item,
      source: type,
      width,
      initialIndex,
    });
  }

  function resetGlobalDrag() {
    document.body.removeChild(node);
    dragging.set(null);
    document.body.style.cursor = "auto";
  }

  function trackDragItem(e: MouseEvent) {
    e.preventDefault();

    requestAnimationFrame(() => {
      node.style.left = `${e.clientX - offset.x}px`;
      node.style.top = `${e.clientY - offset.y}px`;
    });
  }

  function handleMouseMove(e: DragEvent<HTMLButtonElement>) {
    const index = Number(e.currentTarget.dataset.index);

    const { width, left } = e.currentTarget.getBoundingClientRect();
    const midwayPoint = left + width / 2;

    const isLeft = e.clientX <= midwayPoint;

    let newIndex = isLeft ? index : index + 1;

    if (dragChip?.type === PivotChipType.Dimension) {
      const maxIndex = items.findIndex((i) => i.type === PivotChipType.Measure);
      if (maxIndex !== -1) {
        newIndex = Math.min(maxIndex, newIndex);
      }
    } else {
      newIndex = Math.max(lastDimensionIndex + 1, newIndex);
    }

    ghostIndex = newIndex;
  }

  function addSwapListeners(currentTarget: EventTarget & HTMLDivElement) {
    const children = currentTarget.querySelectorAll(".drag-item");
    children.forEach((child) => {
      child.addEventListener("mousemove", handleMouseMove);
    });
  }

  function removeSwapListeners(currentTarget: EventTarget & HTMLDivElement) {
    const children = currentTarget.querySelectorAll(".drag-item");
    children.forEach((child) => {
      child.removeEventListener("mousemove", handleMouseMove);
    });
  }

  function handleDrop(e: DragEvent<HTMLDivElement>) {
    if (zoneStartedDrag) $controllerStore?.abort();

    if (isValidDropZone) {
      if (dragChip && ghostIndex !== undefined) {
        const temp = [...items];

        temp.splice(ghostIndex, 0, dragChip);

        items = temp;

        dispatch("update", items);
      }
      removeSwapListeners(e.currentTarget);
    }
    ghostIndex = undefined;
  }

  function handleDragEnter(e: DragEvent<HTMLDivElement>) {
    if (!currentlyDragging) return;

    if (zoneStartedDrag && !isDropLocation) {
      ghostIndex = initialIndex;
      return;
    }
    if (!isValidDropZone) return;

    ghostIndex =
      dragChip?.type === PivotChipType.Dimension
        ? lastDimensionIndex + 1
        : items.length;

    addSwapListeners(e.currentTarget);
  }

  function handleDragLeave(e: DragEvent<HTMLDivElement>) {
    if (!currentlyDragging) return;
    ghostIndex = undefined;

    removeSwapListeners(e.currentTarget);
  }
</script>

<div
  role="presentation"
  class="dnd-zone group"
  class:horizontal={isDropLocation}
  class:valid={isValidDropZone}
  style:--ghost-width="{ghostWidth ?? 0}px"
  on:mouseup={handleDrop}
  on:mouseenter={handleDragEnter}
  on:mouseleave={handleDragLeave}
  bind:this={container}
>
  {#each items as item, i (item.id)}
    {#if i === ghostIndex}
      <span
        class="ghost"
        class:rounded={dragChip?.type !== PivotChipType.Measure}
      />
    {/if}

    <div
      id={item.id}
      title={item.title}
      data-type={item.type}
      data-index={i}
      class="drag-item"
      class:hidden={dragChip?.id === item.id && zoneStartedDrag}
      class:rounded={item.type !== PivotChipType.Measure}
    >
      <PivotChip
        grab
        removable={isDropLocation}
        {item}
        on:mousedown={(e) => handleMouseDown(e, item)}
        on:remove={() => {
          items = items.filter((i) => i.id !== item.id);
          dispatch("update", items);
        }}
      />
    </div>
  {:else}
    {#if ghostIndex === undefined}
      <p>{placeholder}</p>
    {/if}
  {/each}

  {#if ghostIndex === items.length}
    <span
      class="ghost"
      class:rounded={dragChip?.type !== PivotChipType.Measure}
    />
  {/if}

  {#if type === "columns" || type === "rows"}
    <AddField {type} />
  {/if}
</div>

<style lang="postcss">
  .drag-item {
    @apply w-fit;
  }

  .ghost {
    @apply bg-gray-200 rounded-sm pointer-events-none;
    height: 26px;
    width: var(--ghost-width);
  }

  .dnd-zone {
    @apply w-full max-w-full rounded-sm;
    @apply flex flex-col;
    @apply gap-y-2 py-2  text-gray-500;
  }

  .horizontal {
    @apply flex flex-row flex-wrap bg-slate-50 w-full p-1 px-2 gap-x-2 h-fit;
    @apply items-center;
    @apply border border-slate-50;
  }

  .valid {
    @apply border-blue-400;
  }

  .valid:hover {
    @apply bg-white;
  }

  .rounded {
    @apply rounded-full;
  }
</style>
