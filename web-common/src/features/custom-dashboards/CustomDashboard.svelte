<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import GridLines from "./GridLines.svelte";
  import Element from "./Element.svelte";
  import type { Vector } from "./types";
  import { vector } from "./util";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let columns = 20;
  export let charts: V1DashboardComponent[];
  export let gap = 4;
  export let showGrid = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;

  let movingIndex: number | null = null;
  let resizingIndex: number | null = null;

  let startMouse: Vector | null = null;
  let mousePosition: Vector | null = null;

  let initialOffset: Vector | null = null;
  let initialResizeElementDimensions: Vector | null = null;
  let initialMoveElementPosition: Vector | null = null;

  $: gapSize = contentRect.width * (gap / 200);

  $: gridCellSize = contentRect.width / columns;

  $: dragPosition =
    movingIndex !== null && mousePosition && initialOffset
      ? vector.add(mousePosition, initialOffset)
      : null;

  $: resizeOffset =
    resizingIndex !== null && mousePosition && startMouse
      ? vector.subtract(mousePosition, startMouse)
      : null;

  $: resizeDimenions =
    resizingIndex !== null && resizeOffset && initialResizeElementDimensions
      ? vector.add(resizeOffset, initialResizeElementDimensions)
      : null;

  function handleMouseUp() {
    window.removeEventListener("mouseup", handleMouseUp);
    window.removeEventListener("mousemove", handleMouseMove);

    if (movingIndex !== null && dragPosition) {
      const cellPosition = getCell(dragPosition);
      charts[movingIndex].x = cellPosition[0];
      charts[movingIndex].y = cellPosition[1];

      dispatch("update", {
        index: movingIndex,
        change: "position",
        vector: cellPosition,
      });

      movingIndex = null;
    }

    if (resizingIndex !== null && resizeDimenions) {
      const dimensions = getCell(resizeDimenions);
      charts[resizingIndex].width = dimensions[0];
      charts[resizingIndex].height = dimensions[1];

      dispatch("update", {
        index: resizingIndex,
        change: "dimension",
        vector: dimensions,
      });

      resizingIndex = null;
    }

    mousePosition = null;
    startMouse = null;
    initialOffset = null;
    initialMoveElementPosition = null;
    initialResizeElementDimensions = null;
    resizeDimenions = null;
  }

  function handleMouseDown(
    e: CustomEvent<{
      e: MouseEvent & { currentTarget: HTMLButtonElement };
      position: Vector;
    }>,
  ) {
    e.preventDefault();

    const index = Number(e.detail.e.currentTarget.dataset.index);

    initialMoveElementPosition = e.detail.position;

    startMouse = [
      e.detail.e.clientX - contentRect.left,
      e.detail.e.clientY - contentRect.top - scrollOffset,
    ];

    initialOffset = vector.subtract(initialMoveElementPosition, startMouse);

    movingIndex = index;

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleResizeStart(
    e: CustomEvent<{
      e: MouseEvent & { currentTarget: HTMLButtonElement };
      dimensions: Vector;
    }>,
  ) {
    e.preventDefault();

    const index = Number(e.detail.e.currentTarget.dataset.index);

    initialResizeElementDimensions = e.detail.dimensions;

    startMouse = [
      e.detail.e.clientX - contentRect.left,
      e.detail.e.clientY - contentRect.top - scrollOffset,
    ];

    initialOffset = vector.subtract(startMouse, startMouse);

    resizingIndex = index;

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleMouseMove(e: MouseEvent) {
    mousePosition = [
      e.clientX - contentRect.left,
      e.clientY - contentRect.top - scrollOffset,
    ];
  }

  function handleScroll(e: Event) {
    scrollOffset = (e.target as HTMLDivElement).scrollTop;
  }

  function getCell(vector: Vector): Vector {
    return [
      Math.round(vector[0] / gridCellSize),
      Math.round(vector[1] / gridCellSize),
    ];
  }
</script>

<div role="presentation" class="container" bind:contentRect>
  {#if showGrid || movingIndex !== null || resizingIndex !== null}
    <GridLines {gridCellSize} {scrollOffset} />
  {/if}

  <div class="dashboard" on:scroll={handleScroll}>
    {#each charts as chart, i (i)}
      {@const isMoving = i === movingIndex}
      {@const isResizing = i === resizingIndex}
      {#if chart.chart}
        <Element
          {i}
          {gapSize}
          {isMoving}
          chartName={chart.chart}
          position={isMoving && dragPosition
            ? dragPosition
            : [Number(chart.x) * gridCellSize, Number(chart.y) * gridCellSize]}
          dimensions={isResizing && resizeDimenions
            ? resizeDimenions
            : [
                Number(chart.width) * gridCellSize,
                Number(chart.height) * gridCellSize,
              ]}
          on:mousedown={handleMouseDown}
          on:resizestart={handleResizeStart}
        />
      {/if}
    {/each}
  </div>
</div>

<style lang="postcss">
  .dashboard {
    @apply w-full h-full overflow-y-scroll overflow-x-hidden relative;
  }

  .container {
    @apply w-full h-full overflow-y-scroll relative;
  }
</style>
