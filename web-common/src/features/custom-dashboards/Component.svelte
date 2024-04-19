<script lang="ts" context="module">
  import { onMount } from "svelte";
  import { writable } from "svelte/store";
  import type ResizeHandle from "./ResizeHandle.svelte";
  import type { ComponentType } from "svelte";

  import { builderActions, getAttrs, type Builder } from "bits-ui";

  const zIndex = writable(0);

  const options = [0, 0.5, 1];
  const allSides = options
    .flatMap((y) => options.map((x) => [x, y] as [number, number]))
    .filter(([x, y]) => !(x === 0.5 && y === 0.5));
</script>

<script lang="ts">
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
  export let i: number;
  export let localZIndex = 0;
  export let chartView = false;

  let ResizeHandleComponent: ComponentType<ResizeHandle>;

  onMount(async () => {
    localZIndex = $zIndex;
    zIndex.set(++localZIndex);

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
  style:z-index={localZIndex}
  style:padding="{padding}px"
  style:left="{left}px"
  style:top="{top}px"
  style:width="{width}px"
  style:height={chartView ? undefined : `${height}px`}
  on:contextmenu
  on:mousedown|capture
>
  <div class="size-full relative">
    {#if ResizeHandleComponent && !embed && !chartView}
      {#each allSides as side}
        <ResizeHandleComponent
          {scale}
          {i}
          {side}
          position={[left, top]}
          dimensions={[width, height]}
          {selected}
          on:change
        />
      {/each}
    {/if}

    <div
      class="size-full overflow-hidden"
      class:shadow-lg={interacting}
      style:border-radius="{radius}px"
    >
      <slot />
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply absolute;
  }
</style>
