<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { GridStackManager } from "./grid-stack-manager";
  import "gridstack/dist/gridstack.min.css";

  export let readonly = false;

  let gridStackElement: HTMLElement;
  const gridStackManager = new GridStackManager(readonly);

  onMount(() => {
    if (gridStackElement) {
      gridStackManager.initialize();
    }
  });

  onDestroy(() => {
    gridStackManager.destroy();
  });
</script>

<div bind:this={gridStackElement} class="grid-stack">
  <slot />
</div>

<style lang="postcss">
</style>
