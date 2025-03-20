<script lang="ts">
  import CanvasFilters from "./filters/CanvasFilters.svelte";

  export let maxWidth: number;
  export let clientWidth = 0;
  export let showGrabCursor = false;
  export let filtersEnabled: boolean | undefined;
  export let onClick: () => void = () => {};

  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);

  $: ({ width: clientWidth } = contentRect);
</script>

{#if filtersEnabled}
  <header
    role="presentation"
    on:click|self={onClick}
    class="bg-background border-b py-4 px-2 w-full select-none"
  >
    <CanvasFilters />
  </header>
{/if}

<div
  role="presentation"
  id="canvas-scroll-container"
  class="size-full overflow-hidden overflow-y-auto p-2 pb-48 flex flex-col items-center bg-white select-none"
  on:click|self={onClick}
  class:!cursor-grabbing={showGrabCursor}
>
  <div
    class="w-full h-fit flex dashboard-theme-boundary flex-col items-center row-container relative"
    style:max-width={maxWidth + "px"}
    style:min-width="420px"
    bind:contentRect
  >
    <slot />
  </div>
</div>

<style>
  div {
    container-type: inline-size;
    container-name: canvas-container;
  }
</style>
