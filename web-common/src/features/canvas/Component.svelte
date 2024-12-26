<script lang="ts" context="module">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import ComponentRenderer from "@rilldata/web-common/features/canvas/components/ComponentRenderer.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let builders: Builder[] = [];
  export let left: number;
  export let top: number;
  export let padding: number;
  // export let scale: number;
  // export let embed = false;
  export let radius: number;
  export let selected = false;
  export let interacting = false;
  export let width: number;
  export let height: number;
  export let chartView = false;
  export let componentName: string;
  export let instanceId: string;
  export let draggable = false;
  export let rowIndex: number;
  export let columnIndex: number;

  $: resourceQuery = useResource(
    instanceId,
    componentName,
    ResourceKind.Component,
  );
  $: ({ data: componentResource } = $resourceQuery);

  $: ({ renderer, rendererProperties } =
    componentResource?.component?.spec ?? {});

  $: title = rendererProperties?.title;
  $: description = rendererProperties?.description;

  $: componentClasses = [
    "component",
    "pointer-events-auto",
    draggable ? "hover:cursor-grab active:cursor-grabbing" : "",
  ].join(" ");
</script>

<!-- FIXME: add data-component-type, need to add type to V1CanvasItem -->
<div
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  role="presentation"
  data-component-index={i}
  data-row-index={rowIndex}
  data-column-index={columnIndex}
  data-selected={selected}
  class={componentClasses}
  {draggable}
  style:z-index={renderer === "select" ? 100 : "auto"}
  style:padding="{padding}px"
  style:left="{left}px"
  style:top="{top}px"
  style:width="{width}px"
  style:height={chartView ? undefined : `${height}px`}
  style:border={selected ? "2px solid var(--color-primary-300)" : "none"}
  style:border-radius={selected ? "2px" : ""}
  on:dragstart
  on:dragend
  on:dragover
  on:drop
  on:mousedown
>
  {#if draggable}
    <div class="drag-handle">
      <DragHandle size="20" className="text-slate-600" />
    </div>
  {/if}
  <div class="size-full relative {draggable ? 'touch-none' : ''}">
    <div
      class="size-full overflow-hidden flex flex-col flex-none"
      class:shadow-lg={interacting}
      style:border-radius="{radius}px"
    >
      {#if title || description}
        <div class="w-full h-fit flex flex-col border-b bg-white p-2">
          {#if title}
            <h1 class="text-slate-700">{title}</h1>
          {/if}
          {#if description}
            <h2 class="text-slate-600 leading-none">{description}</h2>
          {/if}
        </div>
      {/if}
      {#if renderer && rendererProperties}
        <ComponentRenderer {renderer} {componentName} />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .component {
    @apply absolute touch-none;
    &[draggable="true"] {
      @apply select-none;
    }
  }

  .drag-handle {
    @apply absolute top-2 left-2.5  p-1 cursor-grab z-10 opacity-0 transition-opacity duration-200;
  }

  .component:hover .drag-handle {
    @apply opacity-100;
  }

  .drag-handle:active {
    @apply cursor-grabbing;
  }

  h1 {
    font-size: 16px;
    font-weight: 500;
  }

  h2 {
    font-size: 12px;
    font-weight: 400;
  }
</style>
