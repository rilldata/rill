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

<main class="size-full flex flex-col dashboard-theme-boundary overflow-hidden">
  {#if filtersEnabled}
    <header
      role="presentation"
      class="bg-background border-b py-4 px-2 w-full h-fit select-none z-50 flex items-center justify-center"
      on:click|self={onClick}
    >
      <CanvasFilters {maxWidth} />
    </header>
  {/if}

  <div
    role="presentation"
    id="canvas-scroll-container"
    class="size-full p-2 pb-48 flex flex-col items-center bg-white select-none overflow-y-auto overflow-x-hidden"
    class:!cursor-grabbing={showGrabCursor}
    on:click|self={onClick}
  >
    <div
      class="w-full h-fit flex flex-col items-center row-container relative"
      style:max-width="{maxWidth}px"
      style:min-width="420px"
      bind:contentRect
    >
      <slot />
    </div>
  </div>
</main>

<style>
  div {
    container-type: inline-size;
    container-name: canvas-container;
  }
</style>
