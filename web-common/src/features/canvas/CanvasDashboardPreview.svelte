<script lang="ts">
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { canvasStore } from "@rilldata/web-common/features/canvas/stores/canvas-stores";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import PreviewElement from "./PreviewElement.svelte";
  import SvelteGridStack from "./SvelteGridStack.svelte";
  import type {
    GridItemHTMLElement,
    GridStack,
    GridStackNode,
  } from "gridstack";
  import { createEventDispatcher } from "svelte";

  export let items: V1CanvasItem[];
  export let selectedIndex: number | null = null;

  const { canvasName } = getCanvasStateManagers();
  const dispatch = createEventDispatcher();

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  // let scrollOffset = 0;
  let changing = false;

  $: instanceId = $runtime.instanceId;

  // $: gridWidth = contentRect.width;
  // $: scale = gridWidth / defaults.DASHBOARD_WIDTH;
  // $: gridCell = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;

  let grid: GridStack;

  $: if (grid) {
    canvasStore.setGridstack($canvasName, grid);
  }

  // function handleMousedown(
  //   e: CustomEvent<{
  //     e: MouseEvent & { currentTarget: HTMLButtonElement };
  //   }>,
  // ) {
  //   console.log("CanvasDashboardPreview handleMousedown");
  //   e.preventDefault();
  //   changing = true;
  // }

  function handleDelete(e: CustomEvent<{ index: number }>) {
    console.log("CanvasDashboardPreview handleDelete");
    items.splice(e.detail.index, 1);
  }

  // function handleScroll(
  //   e: UIEvent & {
  //     currentTarget: EventTarget & HTMLDivElement;
  //   },
  // ) {
  //   scrollOffset = e.currentTarget.scrollTop;
  // }

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
    const { w, h, x, y } =
      (e.detail.target?.gridstackNode as GridStackNode) || {};

    dispatch("update", {
      index: selectedIndex,
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
    const { w, h, x, y } =
      (e.detail.target?.gridstackNode as GridStackNode) || {};

    dispatch("update", {
      index: selectedIndex,
      x: Number(x),
      y: Number(y),
      w: Number(w),
      h: Number(h),
    });
  }

  function handleSelect(e: CustomEvent<{ index: number }>) {
    console.log("CanvasDashboardPreview handleSelect", e.detail.index);

    selectedIndex = e.detail.index;
    canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
  }
</script>

<CanvasDashboardWrapper bind:contentRect height={maxBottom}>
  <SvelteGridStack
    bind:grid
    {items}
    let:index
    let:item
    on:select={handleSelect}
    on:resizestop={handleResizeStop}
    on:dragstop={handleDragStop}
  >
    {@const selected = index === selectedIndex}
    {@const interacting = selected && changing}
    <PreviewElement
      {instanceId}
      i={index}
      component={item}
      {interacting}
      on:delete={handleDelete}
    />
  </SvelteGridStack>
</CanvasDashboardWrapper>
