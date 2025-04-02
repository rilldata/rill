<script lang="ts" context="module">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import { hideBorder } from "./layout-util";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import ComponentError from "@rilldata/web-common/features/canvas/components/ComponentError.svelte";
  import {
    isCanvasComponentType,
    isChartComponentType,
    getComponentFilterProperties,
  } from "@rilldata/web-common/features/canvas/components/util";
  import Toolbar from "./Toolbar.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
</script>

<script lang="ts">
  export let component: BaseCanvasComponent;
  export let selected = false;
  export let ghost = false;
  export let allowPointerEvents = true;
  export let editable = false;
  export let onMouseDown: (e: MouseEvent) => void = () => {};
  export let onDuplicate: () => void = () => {};
  export let onDelete: () => void = () => {};

  let open = false;

  $: ({ id: componentName, specStore, type: renderer, resource } = component);

  $: rendererProperties = $specStore;
  $: componentResource = $resource;

  $: isChartType = isChartComponentType(renderer);

  $: title = rendererProperties?.["title"] as string | undefined;
  $: description = rendererProperties?.["description"] as string | undefined;
  $: componentFilters = getComponentFilterProperties(rendererProperties);

  $: allowBorder = !hideBorder.has(renderer);
</script>

<article
  role="presentation"
  id={componentName}
  class:selected
  class:editable
  class:opacity-20={ghost}
  style:pointer-events={!allowPointerEvents ? "none" : "auto"}
  class:outline={allowBorder || open}
  class:shadow-sm={allowBorder || open}
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
        <svelte:component this={component.component} {component} />
        <!-- {#if isChartComponentType(renderer) && timeAndFilterStore}
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
        {/if} -->
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
