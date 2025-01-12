<script lang="ts">
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createEventDispatcher } from "svelte";
  import * as defaults from "./constants";
  import DashboardWrapper from "./DashboardWrapper.svelte";
  import PreviewElement from "./PreviewElement.svelte";
  import type { Vector } from "./types";
  import { vector } from "./util";

  const dispatch = createEventDispatcher();
  const zeroVector = [0, 0] as [0, 0];

  export let columns: number | undefined;
  export let items: V1CanvasItem[];
  export let gap: number | undefined;
  export let showGrid = false;
  export let snap = true;
  export let selectedIndex: number | null = null;

  const { canvasEntity } = getCanvasStateManagers();

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);
  let scrollOffset = 0;
  let changing = false;
  let startMouse: Vector = [0, 0];
  let mousePosition: Vector = [0, 0];
  let initialElementDimensions: Vector = [0, 0];
  let initialElementPosition: Vector = [0, 0];
  let dimensionChange: [0 | 1 | -1, 0 | 1 | -1] = [0, 0];
  let positionChange: [0 | 1, 0 | 1] = [0, 0];

  $: ({ instanceId } = $runtime);

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / defaults.DASHBOARD_WIDTH;

  $: gapSize = defaults.DASHBOARD_WIDTH * ((gap ?? defaults.GAP_SIZE) / 1000);
  $: gridCell = defaults.DASHBOARD_WIDTH / (columns ?? defaults.COLUMN_COUNT);
  $: radius = gridCell * defaults.COMPONENT_RADIUS;
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

    items[selectedIndex].x = Math.max(
      0,
      dimensions[0] < 0 ? cellPosition[0] + dimensions[0] : cellPosition[0],
    );
    items[selectedIndex].y = Math.max(
      0,
      dimensions[1] < 0 ? cellPosition[1] + dimensions[1] : cellPosition[1],
    );

    items[selectedIndex].width = Math.max(1, Math.abs(dimensions[0]));
    items[selectedIndex].height = Math.max(1, Math.abs(dimensions[1]));

    dispatch("update", {
      index: selectedIndex,
      position: [items[selectedIndex].x, items[selectedIndex].y],
      dimensions: [items[selectedIndex].width, items[selectedIndex].height],
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
    canvasEntity.setSelectedComponentIndex(selectedIndex);
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
    canvasEntity.setSelectedComponentIndex(selectedIndex);
  }

  $: maxBottom = items.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<svelte:window on:mousemove={handleMouseMove} on:mouseup={handleMouseUp} />

<div
  id="header"
  class="border-b w-fit min-w-full flex flex-col bg-slate-50 slide"
>
  <CanvasFilters />
</div>

<DashboardWrapper
  bind:contentRect
  {changing}
  {gapSize}
  {gridCell}
  {scrollOffset}
  {radius}
  {scale}
  {showGrid}
  height={maxBottom * gridCell * scale}
  width={defaults.DASHBOARD_WIDTH}
  on:click={deselect}
  on:scroll={handleScroll}
>
  <section
    class="flex relative justify-between gap-x-4 py-4 pb-6 px-4"
  ></section>
  {#each items as component, i (i)}
    {@const selected = i === selectedIndex}
    {@const interacting = selected && changing}
    <PreviewElement
      {instanceId}
      {i}
      {scale}
      {component}
      {radius}
      {selected}
      {interacting}
      {gapSize}
      width={interacting
        ? finalResize[0]
        : Number(component.width ?? defaults.COMPONENT_WIDTH) * gridCell}
      height={interacting
        ? finalResize[1]
        : Number(component.height ?? defaults.COMPONENT_HEIGHT) * gridCell}
      top={interacting ? finalDrag[1] : Number(component.y) * gridCell}
      left={interacting ? finalDrag[0] : Number(component.x) * gridCell}
      on:change={handleChange}
      on:delete
    />
  {/each}
</DashboardWrapper>
