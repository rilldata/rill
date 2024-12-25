<script lang="ts" context="module">
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import type ResizeHandle from "./ResizeHandle.svelte";
  import ComponentRenderer from "@rilldata/web-common/features/canvas/components/ComponentRenderer.svelte";
  import ComponentTitle from "@rilldata/web-common/features/canvas/ComponentTitle.svelte";

  const options = [0, 0.5, 1];
  const allSides = options
    .flatMap((y) => options.map((x) => [x, y] as [number, number]))
    .filter(([x, y]) => !(x === 0.5 && y === 0.5));
</script>

<script lang="ts">
  export let i: number;
  export let builders: Builder[] = [];
  export let left: number;
  export let top: number;
  export let padding: number;
  export let scale: number;
  export let embed = false;
  export let radius: number;
  export let selected = false;
  export let interacting = false;
  export let width: number;
  export let height: number;
  export let chartView = false;
  export let componentName: string;
  export let instanceId: string;
  export let draggable = false;

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

  // let ResizeHandleComponent: ComponentType<ResizeHandle>;

  // onMount(async () => {
  //   if (!embed) {
  //     ResizeHandleComponent = (await import("./ResizeHandle.svelte")).default;
  //   }
  // });

  $: componentClasses = [
    "component",
    "pointer-events-auto",
    draggable ? "hover:cursor-grab active:cursor-grabbing" : "",
  ].join(" ");
</script>

<div
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  role="presentation"
  data-index={i}
  data-component
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
  on:contextmenu
  on:mousedown
>
  <div class="size-full relative {draggable ? 'touch-none' : ''}">
    <!-- {#if ResizeHandleComponent && !embed}
      {#each allSides as side (side)}
        <svelte:component
          this={ResizeHandleComponent}
          {i}
          {scale}
          {side}
          position={[left, top]}
          dimensions={[width, height]}
          {selected}
          on:change
        />
      {/each}
    {/if} -->

    <div
      class="size-full overflow-hidden flex flex-col flex-none"
      class:shadow-lg={interacting}
      style:border-radius="{radius}px"
    >
      <ComponentTitle {title} {description} />
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
</style>
