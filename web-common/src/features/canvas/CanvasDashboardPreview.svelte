<script lang="ts">
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type {
    V1CanvasItem,
    V1CanvasSpec,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import PreviewElement from "./PreviewElement.svelte";
  import SvelteGridStack from "./SvelteGridStack.svelte";
  import type {
    GridItemHTMLElement,
    GridStack,
    GridStackNode,
  } from "gridstack";
  import { createEventDispatcher } from "svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { clickOutside } from "@rilldata/web-common/lib/actions/click-outside";

  export let items: V1CanvasItem[];
  export let showFilterBar = true;
  export let activeIndex: number | null = null;
  export let spec: V1CanvasSpec;

  const { canvasEntity } = getCanvasStateManagers();
  const dispatch = createEventDispatcher();

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let grid: GridStack;
  let gridContainer: HTMLElement;

  $: instanceId = $runtime.instanceId;

  $: if (grid) {
    canvasEntity.setGridstack(grid);
  }

  function handleDelete(e: CustomEvent<{ index: number }>) {
    items.splice(e.detail.index, 1);
  }

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);

  function handleResizeStop(
    e: CustomEvent<{
      event: Event;
      el: GridItemHTMLElement;
      target: GridItemHTMLElement;
    }>,
  ) {
    // Get index from the resized element's data attribute
    const contentEl = e.detail.target?.querySelector(
      ".grid-stack-item-content-item",
    );
    const index = contentEl?.getAttribute("data-index");

    if (index === null) {
      console.error("No index found for resized widget");
      return;
    }

    const { w, h, x, y } =
      (e.detail.target?.gridstackNode as GridStackNode) || {};

    dispatch("update", {
      index: Number(index), // Use the data-index from the element
      x: Number(x),
      y: Number(y),
      w: Number(w),
      h: Number(h),
    });
  }

  function handleDragStop(
    e: CustomEvent<{
      event: Event;
      el: GridItemHTMLElement;
      target: GridItemHTMLElement;
    }>,
  ) {
    const contentEl = e.detail.target?.querySelector(
      ".grid-stack-item-content-item",
    );
    const index = contentEl?.getAttribute("data-index");

    if (index === null) {
      console.error("No index found for dragged widget");
      return;
    }

    const { w, h, x, y } =
      (e.detail.target?.gridstackNode as GridStackNode) || {};

    dispatch("update", {
      index: Number(index),
      x: Number(x),
      y: Number(y),
      w: Number(w),
      h: Number(h),
    });
  }

  function handleSelect(e: CustomEvent<{ index: number }>) {
    activeIndex = e.detail.index;
    canvasEntity.setSelectedComponentIndex(activeIndex);
  }

  function handleDeselect() {
    activeIndex = null;
    canvasEntity.setSelectedComponentIndex(activeIndex);
  }

  function handleClickOutside(event: Event) {
    const target = event.target as HTMLElement;
    const gridStackItemEl = target.closest(".grid-stack-item");
    const inspectorWrapperEl = target.closest(".inspector-wrapper");

    // Deselect if click is outside of both grid-stack-item and inspector-wrapper
    if (!gridStackItemEl && !inspectorWrapperEl) {
      activeIndex = null;
      canvasEntity.setSelectedComponentIndex(activeIndex);
    }
  }
</script>

{#if showFilterBar}
  <div
    id="header"
    class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
  >
    <CanvasFilters />
  </div>
{/if}

<CanvasDashboardWrapper bind:contentRect height={maxBottom}>
  <div bind:this={gridContainer} use:clickOutside={[null, handleClickOutside]}>
    <SvelteGridStack
      bind:grid
      {items}
      {spec}
      let:index
      let:item
      on:select={handleSelect}
      on:deselect={handleDeselect}
      on:resizestop={handleResizeStop}
      on:dragstop={handleDragStop}
    >
      {@const selected = index === activeIndex}
      <PreviewElement
        {instanceId}
        i={index}
        {selected}
        component={item}
        on:delete={handleDelete}
      />
    </SvelteGridStack>
  </div>
</CanvasDashboardWrapper>
