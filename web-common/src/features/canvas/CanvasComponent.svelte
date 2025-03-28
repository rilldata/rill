<script lang="ts" context="module">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type {
    V1CanvasItem,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { hideBorder } from "./layout-util";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import { KPIGrid } from "@rilldata/web-common/features/canvas/components/kpi-grid";
  import {
    isCanvasComponentType,
    isChartComponentType,
    getComponentFilterProperties,
  } from "@rilldata/web-common/features/canvas/components/util";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import type { Readable } from "svelte/store";
  import { Chart } from "./components/charts";
  import { Image } from "./components/image";
  import { KPI } from "./components/kpi";
  import { Markdown } from "./components/markdown";
  import { Pivot } from "./components/pivot";
  import { Table } from "./components/table";
  import Toolbar from "./Toolbar.svelte";

  const filterableComponents = new Map([
    ["kpi", KPI],
    ["kpi_grid", KPIGrid],
    ["table", Table],
    ["pivot", Pivot],
  ]);

  const nonFilterableComponents = new Map([
    ["markdown", Markdown],
    ["image", Image],
  ]);
</script>

<script lang="ts">
  export let canvasItem: V1CanvasItem | null;
  export let selected = false;
  export let id: string;
  export let ghost = false;
  export let allowPointerEvents = true;
  export let editable = false;
  export let canvasName: string;
  export let componentResource: V1Resource | undefined;
  export let onMouseDown: (e: MouseEvent) => void = () => {};
  export let onDuplicate: () => void = () => {};
  export let onDelete: () => void = () => {};

  $: ({
    canvasEntity: { componentTimeAndFilterStore },
  } = getCanvasStore(canvasName));

  let open = false;
  let timeAndFilterStore: Readable<TimeAndFilterStore> | undefined;

  $: componentName = canvasItem?.component ?? "";

  $: ({ renderer, rendererProperties } =
    componentResource?.component?.spec ?? {});

  $: isChartType = isChartComponentType(renderer);

  $: title = rendererProperties?.title as string | undefined;
  $: description = rendererProperties?.description as string | undefined;
  $: componentFilters = getComponentFilterProperties(rendererProperties);

  $: isFilterable = filterableComponents.has(renderer ?? "");

  $: hasHeader = !!title || !!description;

  $: if (
    (isChartComponentType(renderer) || isFilterable) &&
    rendererProperties?.metrics_view
  ) {
    timeAndFilterStore = componentTimeAndFilterStore(componentName);
  }
</script>

<article
  role="presentation"
  {id}
  class:selected
  class:editable
  class:opacity-20={ghost}
  style:pointer-events={!allowPointerEvents ? "none" : "auto"}
  class:outline={!hideBorder.has(renderer) || open}
  class:shadow-sm={!hideBorder.has(renderer) || open}
  class="group component-card size-full flex flex-col cursor-pointer z-10 p-0 relative outline-[1px] outline-gray-200 bg-white overflow-hidden rounded-sm"
>
  {#if editable}
    <Toolbar {onDelete} {onDuplicate} bind:dropdownOpen={open} />
  {/if}

  <div
    role="presentation"
    class="size-full grow flex flex-col"
    on:mousedown={onMouseDown}
  >
    {#if componentName}
      {#if !isChartType}
        <ComponentHeader {title} {description} filters={componentFilters} />
      {/if}
      {#if rendererProperties && isCanvasComponentType(renderer)}
        {#if isChartComponentType(renderer) && timeAndFilterStore}
          <Chart
            {canvasName}
            {rendererProperties}
            {renderer}
            {timeAndFilterStore}
          />
        {:else if renderer === "pivot" && timeAndFilterStore}
          <Pivot
            {canvasName}
            {hasHeader}
            {rendererProperties}
            {timeAndFilterStore}
            {componentName}
          />
        {:else if renderer === "table" && timeAndFilterStore}
          <Table
            {canvasName}
            {rendererProperties}
            {timeAndFilterStore}
            {componentName}
            {hasHeader}
          />
        {:else if isFilterable && timeAndFilterStore}
          <svelte:component
            this={filterableComponents.get(renderer)}
            {canvasName}
            {rendererProperties}
            {timeAndFilterStore}
          />
        {:else}
          <svelte:component
            this={nonFilterableComponents.get(renderer)}
            {rendererProperties}
          />
        {/if}
      {:else if componentResource}
        <ComponentError error="Invalid component type" />
      {/if}
    {:else}
      <div class="size-full grid place-content-center">
        <LoadingSpinner size="36px" />
      </div>
    {/if}
  </div>
</article>

<style lang="postcss">
  .component-card.editable:hover {
    @apply shadow-md outline;
  }

  .component-card:has(.component-error) {
    @apply outline-red-200;
  }

  .selected {
    @apply shadow-md outline-primary-400 outline-[1.5px];

    outline-style: solid !important;
  }
</style>
