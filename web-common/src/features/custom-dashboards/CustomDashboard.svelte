<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import GridLines from "./GridLines.svelte";
  import Element from "./Element.svelte";
  import type { Vector } from "./types";
  import { vector } from "./util";
  import { createEventDispatcher } from "svelte";

  const DEFAULT_WIDTH = 2000;

  const dispatch = createEventDispatcher();
  const zeroVector = [0, 0] as [0, 0];

  export let columns = 20;
  export let charts: V1DashboardComponent[];
  export let gap = 4;
  export let showGrid = false;
  export let snap = false;
  export let selectedChartName: string | null;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;

  let selectedIndex: number | null = null;
  let changing = false;

  let startMouse: Vector = [0, 0];
  let mousePosition: Vector = [0, 0];
  let initialElementDimensions: Vector = [0, 0];
  let initialElementPosition: Vector = [0, 0];

  let dimensionChange: [0 | 1 | -1, 0 | 1 | -1] = [0, 0];
  let positionChange: [0 | 1, 0 | 1] = [0, 0];

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / DEFAULT_WIDTH;

  $: gapSize = DEFAULT_WIDTH * (gap / 1000);
  $: gridCell = DEFAULT_WIDTH / columns;
  $: radius = gridCell * 0.08;
  $: gridVector = [gridCell, gridCell] as Vector;

  $: mouseDelta = vector.divide(vector.subtract(mousePosition, startMouse), [
    scale,
    scale,
  ]);

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
    if (selectedIndex === null || !changing) return;

    const cellPosition = getCell(dragPosition, true);
    const dimensions = getCell(resizeDimenions, true);

    charts[selectedIndex].x =
      dimensions[0] < 0 ? cellPosition[0] + dimensions[0] : cellPosition[0];
    charts[selectedIndex].y =
      dimensions[1] < 0 ? cellPosition[1] + dimensions[1] : cellPosition[1];

    charts[selectedIndex].width = Math.abs(dimensions[0]);
    charts[selectedIndex].height = Math.abs(dimensions[1]);

    dispatch("update", {
      index: selectedIndex,
      position: [charts[selectedIndex].x, charts[selectedIndex].y],
      dimensions: [charts[selectedIndex].width, charts[selectedIndex].height],
    });

    reset();
  }

  function reset() {
    changing = false;
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

    selectedIndex = index;
    selectedChartName = charts[index].chart ?? null;
    changing = true;
  }

  function handleMouseMove(e: MouseEvent) {
    if (selectedIndex === null || !changing) return;

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

  function deselect() {
    selectedIndex = null;
    selectedChartName = null;
  }

  $: maxBottom = charts.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<svelte:window on:mouseup={handleMouseUp} on:mousemove={handleMouseMove} />

<div bind:contentRect class="wrapper">
  {#if showGrid || changing}
    <GridLines {gridCell} {scrollOffset} {gapSize} {radius} {scale} />
  {/if}
  <div
    role="presentation"
    class="size-full overflow-y-auto overflow-x-hidden relative"
    on:scroll={handleScroll}
  >
    <div
      class="dash"
      role="presentation"
      style:width="{DEFAULT_WIDTH}px"
      style:height="{maxBottom * gridCell}px"
      style:transform="scale({scale})"
      on:click|self={deselect}
    >
      {#each charts as chart, i (i)}
        {@const selected = i === selectedIndex}
        {@const interacting = selected && changing}
        {#if chart.chart}
          <Element
            {scale}
            {i}
            {chart}
            {radius}
            {selected}
            {interacting}
            {gapSize}
            width={interacting
              ? finalResize[0]
              : Number(chart.width) * gridCell}
            height={interacting
              ? finalResize[1]
              : Number(chart.height) * gridCell}
            top={interacting ? finalDrag[1] : Number(chart.y) * gridCell}
            left={interacting ? finalDrag[0] : Number(chart.x) * gridCell}
            on:change={handleChange}
          />
        {/if}
      {/each}
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    user-select: none;
    margin: 0;
    pointer-events: auto;
  }

  .dash {
    transform-origin: top left;
    position: absolute;
    touch-action: none;
  }
</style>
