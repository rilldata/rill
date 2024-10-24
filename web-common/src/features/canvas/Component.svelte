<script lang="ts" context="module">
  import TemplateRenderer from "@rilldata/web-common/features/templates/TemplateRenderer.svelte";
  import { builderActions, getAttrs, type Builder } from "bits-ui";
  import type { ComponentType } from "svelte";
  import { onMount } from "svelte";
  import {
    ResourceKind,
    useResource,
  } from "../entity-management/resource-selectors";
  import Chart from "./Chart.svelte";
  import type ResizeHandle from "./ResizeHandle.svelte";

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
  export let localZIndex = 0;
  export let chartView = false;
  export let componentName: string;
  export let instanceId: string;

  $: resourceQuery = useResource(
    instanceId,
    componentName,
    ResourceKind.Component,
  );
  $: ({ data: componentResource } = $resourceQuery);

  $: ({
    renderer,
    rendererProperties,
    resolverProperties,
    input,
    output,
    displayName,
    description,
  } = componentResource?.component?.spec ?? {});

  let ResizeHandleComponent: ComponentType<ResizeHandle>;

  onMount(async () => {
    if (!embed) {
      ResizeHandleComponent = (await import("./ResizeHandle.svelte")).default;
    }
  });
</script>

<div
  {...getAttrs(builders)}
  use:builderActions={{ builders }}
  role="presentation"
  data-index={i}
  class="wrapper hover:cursor-pointer active:cursor-grab pointer-events-auto"
  class:!cursor-default={embed}
  style:z-index={renderer === "select" ? 100 : localZIndex}
  style:padding="{padding}px"
  style:left="{left}px"
  style:top="{top}px"
  style:width="{width}px"
  style:height={chartView ? undefined : `${height}px`}
  on:contextmenu
  on:mousedown|capture
>
  <div class="size-full relative">
    {#if ResizeHandleComponent && !embed}
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
    {/if}

    <div
      class="size-full overflow-hidden flex flex-col gap-y-1 flex-none"
      class:shadow-lg={interacting}
      style:border-radius="{radius}px"
    >
      {#if displayName || description}
        <div class="w-full h-fit flex flex-col pb-2">
          {#if displayName}
            <h1 class="text-slate-900">{displayName}</h1>
          {/if}
          {#if description}
            <h2 class="text-slate-600 leading-none">{description}</h2>
          {/if}
        </div>
      {/if}
      {#if renderer === "vega_lite" && rendererProperties?.spec && resolverProperties}
        <Chart {componentName} {chartView} {input} />
      {:else if renderer && rendererProperties}
        <TemplateRenderer
          {chartView}
          {renderer}
          {input}
          {output}
          {resolverProperties}
          {componentName}
        />
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply absolute;
  }

  h1 {
    font-size: 18px;
    font-weight: 600;
  }

  h2 {
    font-size: 14px;
    font-weight: 400;
  }
</style>
