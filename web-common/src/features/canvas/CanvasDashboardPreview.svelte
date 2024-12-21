<script lang="ts">
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { canvasStore } from "@rilldata/web-common/features/canvas/stores/canvas-stores";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import * as defaults from "./constants";
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

  const { canvasName, canvasStore: canvasStoreStore } =
    getCanvasStateManagers();
  const dispatch = createEventDispatcher();

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;
  let changing = false;

  $: instanceId = $runtime.instanceId;

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;

  $: gridCell = defaults.DASHBOARD_WIDTH / defaults.COLUMN_COUNT;

  $: options = {
    column: 12,
    resizable: {
      handles: "e,se,s,sw,w",
    },
    animate: true,
    float: true,
  };

  let grid: GridStack;

  function handleMousedown(
    e: CustomEvent<{
      e: MouseEvent & { currentTarget: HTMLButtonElement };
    }>,
  ) {
    console.log("CanvasDashboardPreview handleMousedown");
    e.preventDefault();
    const index = Number(e.detail.e.currentTarget.dataset.index);
    selectedIndex = index;
    canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
    changing = true;
  }

  function handleDelete(e: CustomEvent<{ index: number }>) {
    console.log("CanvasDashboardPreview handleDelete");
    items.splice(e.detail.index, 1);
  }

  function handleScroll(
    e: UIEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) {
    scrollOffset = e.currentTarget.scrollTop;
  }

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);

  $: items = items.map((item) => {
    const { ...rest } = item;
    return {
      ...rest,
      w: Number(item.width),
      h: Number(item.height),
    };
  });

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

  // function handlePointerEnter(
  //   e: CustomEvent<{
  //     index: number;
  //   }>,
  // ) {
  //   selectedIndex = e.detail.index;
  //   canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
  // }

  // function handlePointerLeave(
  //   e: CustomEvent<{
  //     index: number;
  //   }>,
  // ) {
  //   selectedIndex = null;
  //   canvasStore.setSelectedComponentIndex($canvasName, selectedIndex);
  // }

  // TODO: add pointer events to .grid-stack-item-content
</script>

<CanvasDashboardWrapper
  bind:contentRect
  height={maxBottom * gridCell * scale}
  on:scroll={handleScroll}
>
  <SvelteGridStack
    bind:grid
    {options}
    {items}
    let:index
    let:item
    on:resizestop={handleResizeStop}
    on:dragstop={handleDragStop}
  >
    {@const selected = index === selectedIndex}
    {@const interacting = selected && changing}
    <PreviewElement
      {instanceId}
      i={index}
      component={item}
      {selected}
      {interacting}
      width={Number(item.w) * gridCell}
      height={Number(item.h) * gridCell}
      top={Number(item.y) * gridCell}
      left={Number(item.x) * gridCell}
      on:mousedown={handleMousedown}
      on:delete={handleDelete}
    />
  </SvelteGridStack>
</CanvasDashboardWrapper>
