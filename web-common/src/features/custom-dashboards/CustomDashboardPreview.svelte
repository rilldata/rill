<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import Element from "./Element.svelte";
  import type { Vector } from "./types";
  import { vector } from "./util";
  import { createEventDispatcher } from "svelte";
  import { DEFAULT_WIDTH, DEFAULT_RADIUS } from "./constants";
  import Wrapper from "./Wrapper.svelte";

  const dispatch = createEventDispatcher();
  const zeroVector = [0, 0] as [0, 0];

  export let columns = 20;
  export let components: V1DashboardComponent[];
  export let gap = 4;
  export let showGrid = false;
  export let snap = false;
  export let selectedChartName: string | null;
  export let chartView = false;

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
  $: radius = gridCell * DEFAULT_RADIUS;
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

    components[selectedIndex].x =
      dimensions[0] < 0 ? cellPosition[0] + dimensions[0] : cellPosition[0];
    components[selectedIndex].y =
      dimensions[1] < 0 ? cellPosition[1] + dimensions[1] : cellPosition[1];

    components[selectedIndex].width = Math.max(1, Math.abs(dimensions[0]));
    components[selectedIndex].height = Math.max(1, Math.abs(dimensions[1]));

    dispatch("update", {
      index: selectedIndex,
      position: [components[selectedIndex].x, components[selectedIndex].y],
      dimensions: [
        components[selectedIndex].width,
        components[selectedIndex].height,
      ],
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
    selectedChartName = components[index].chart ?? null;
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

  $: maxBottom = components.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<svelte:window on:mouseup={handleMouseUp} on:mousemove={handleMouseMove} />

<Wrapper
  width={DEFAULT_WIDTH}
  height={maxBottom * gridCell}
  {scale}
  {showGrid}
  {gapSize}
  {gridCell}
  {radius}
  {changing}
  bind:contentRect
  on:scroll={handleScroll}
  on:click={deselect}
>
  {#each components as component, i (i)}
    {@const selected = i === selectedIndex}
    {@const interacting = selected && changing}
    {#if component.chart && component.width && component.height}
      <Element
        {chartView}
        {scale}
        {i}
        chart={component}
        {radius}
        {selected}
        {interacting}
        {gapSize}
        width={interacting
          ? finalResize[0]
          : Number(component.width) * gridCell}
        height={interacting
          ? finalResize[1]
          : Number(component.height) * gridCell}
        top={interacting ? finalDrag[1] : Number(component.y) * gridCell}
        left={interacting ? finalDrag[0] : Number(component.x) * gridCell}
        on:change={handleChange}
      />
    {/if}
  {/each}
</Wrapper>
