<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import GridLines from "./GridLines.svelte";
  import Element from "./Element.svelte";
  import type { Vector } from "./types";
  import { vector } from "./util";

  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  const zeroVector = [0, 0] as [0, 0];

  export let columns = 20;
  export let charts: V1DashboardComponent[];
  export let gap = 4;
  export let showGrid = false;
  export let snap = false;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;

  let elementIndex: number | null = null;

  let startMouse: Vector = [0, 0];
  let mousePosition: Vector = [0, 0];
  let initialElementDimensions: Vector = [0, 0];
  let initialElementPosition: Vector = [0, 0];

  let dimensionChange: [0 | 1 | -1, 0 | 1 | -1] = [0, 0];
  let positionChange: [0 | 1, 0 | 1] = [0, 0];

  $: gridWidth = contentRect.width;
  $: gapSize = gridWidth * (gap / 1000);
  $: gridCell = gridWidth / columns;
  $: gridVector = [gridCell, gridCell] as Vector;

  $: mouseDelta = vector.subtract(mousePosition, startMouse);

  $: dragPosition = vector.add(
    vector.multiply(mouseDelta, positionChange),
    initialElementPosition,
  );

  $: resizeDimenions = vector.add(
    vector.multiply(mouseDelta, dimensionChange),
    initialElementDimensions,
  );

  $: finalDrag = vector.multiply(getCell(dragPosition, snap), gridVector);

  $: finalResize = vector.multiply(getCell(resizeDimenions, snap), gridVector);

  function handleMouseUp() {
    if (elementIndex === null) return;

    const cellPosition = getCell(dragPosition, true);
    const dimensions = getCell(resizeDimenions, true);

    charts[elementIndex].x =
      dimensions[0] < 0 ? cellPosition[0] + dimensions[0] : cellPosition[0];
    charts[elementIndex].y =
      dimensions[1] < 0 ? cellPosition[1] + dimensions[1] : cellPosition[1];

    charts[elementIndex].width = Math.abs(dimensions[0]);
    charts[elementIndex].height = Math.abs(dimensions[1]);

    dispatch("update", {
      index: elementIndex,
      position: [charts[elementIndex].x, charts[elementIndex].y],
      dimensions: [charts[elementIndex].width, charts[elementIndex].height],
    });

    reset();
  }

  function reset() {
    elementIndex = null;
    mousePosition =
      startMouse =
      startMouse =
      initialElementPosition =
      initialElementDimensions =
      dimensionChange =
      positionChange =
      resizeDimenions =
        zeroVector;
  }

  function handleChange(
    e: CustomEvent<{
      e: MouseEvent & { currentTarget: HTMLButtonElement };
      dimensions: Vector;
      position: Vector;
      changeDimensions: [0 | 1 | -1, 0 | 1 | -1];
      changePosition: [0 | 1, 0 | 1];
    }>,
  ) {
    e.preventDefault();

    dimensionChange = e.detail.changeDimensions;
    positionChange = e.detail.changePosition;

    const index = Number(e.detail.e.currentTarget.dataset.index);

    initialElementDimensions = e.detail.dimensions;
    initialElementPosition = e.detail.position;

    startMouse = [
      e.detail.e.clientX - contentRect.left,
      e.detail.e.clientY - contentRect.top - scrollOffset,
    ];

    mousePosition = startMouse;

    elementIndex = index;
  }

  function handleMouseMove(e: MouseEvent) {
    if (elementIndex === null) return;
    mousePosition = [
      e.clientX - contentRect.left,
      e.clientY - contentRect.top - scrollOffset,
    ];
  }

  function handleScroll(
    e: UIEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) {
    scrollOffset = e.currentTarget.scrollTop;
  }

  function getCell(rawVector: Vector, snap: boolean): Vector {
    const raw = vector.divide(rawVector, gridVector);

    if (!snap) return raw;

    return [Math.round(raw[0]), Math.round(raw[1])];
  }
</script>

<svelte:window on:mouseup={handleMouseUp} on:mousemove={handleMouseMove} />

<div role="presentation" class="container relative" bind:contentRect>
  {#if showGrid || elementIndex !== null}
    <GridLines {gridCell} {scrollOffset} {gapSize} />
  {/if}

  <div class="dashboard" on:scroll={handleScroll}>
    {#each charts as chart, i (i)}
      {@const active = i === elementIndex}
      {#if chart.chart}
        <Element
          {i}
          {chart}
          {active}
          {gapSize}
          width={active ? finalResize[0] : Number(chart.width) * gridCell}
          height={active ? finalResize[1] : Number(chart.height) * gridCell}
          top={active ? finalDrag[1] : Number(chart.y) * gridCell}
          left={active ? finalDrag[0] : Number(chart.x) * gridCell}
          on:change={handleChange}
        />
      {/if}
    {/each}
  </div>
</div>

<style lang="postcss">
  .dashboard {
    @apply w-full max-w-full h-full overflow-y-auto overflow-x-hidden relative;
  }

  .container {
    @apply w-full h-full max-w-full overflow-y-auto relative;
  }
</style>
