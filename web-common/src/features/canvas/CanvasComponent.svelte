<script lang="ts" context="module">
  import ComponentHeader from "@rilldata/web-common/features/canvas/ComponentHeader.svelte";
  import ComponentRenderer from "@rilldata/web-common/features/canvas/components/ComponentRenderer.svelte";
  import {
    getComponentFilterProperties,
    isChartComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { hideBorder } from "./layout-util";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
</script>

<script lang="ts">
  import Toolbar from "./Toolbar.svelte";

  export let canvasItem: V1CanvasItem | null;
  export let selected = false;
  export let id: string;
  export let ghost = false;
  export let allowPointerEvents = true;
  export let editable = false;
  export let onMouseDown: (e: MouseEvent) => void = () => {};
  export let onDuplicate: () => void = () => {};
  export let onDelete: () => void = () => {};

  const {
    canvasEntity: {
      spec: { getComponentResourceFromName },
    },
  } = getCanvasStateManagers();

  let open = false;

  $: componentName = canvasItem?.component ?? "";

  $: component = getComponentResourceFromName(componentName);
  $: ({ renderer, rendererProperties } = $component ?? {});

  $: isChartType = isChartComponentType(renderer);

  $: title = rendererProperties?.title;
  $: description = rendererProperties?.description;
  $: componentFilters = getComponentFilterProperties(rendererProperties);
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
      {#if renderer && rendererProperties}
        <ComponentRenderer {renderer} {rendererProperties} {componentName} />
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
