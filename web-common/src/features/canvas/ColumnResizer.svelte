<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let left: number;
  export let height: number;
  export let clicked: boolean;
  export let resizing: boolean = false;

  const dispatch = createEventDispatcher();
</script>

<button
  type="button"
  aria-label="Resize column"
  class="col-resize-handle absolute w-[3px] cursor-col-resize bg-transparent hover:bg-primary-300 z-[50] opacity-0 pointer-events-auto"
  class:opacity-100={clicked || resizing}
  class:bg-primary-300={clicked || resizing}
  style="left: {left}px; height: {height}px;"
  on:mousedown|stopPropagation={(e) => dispatch("resize", e)}
  on:click|stopPropagation={(e) => dispatch("click", e)}
/>

<style>
  .col-resize-handle {
    position: absolute;
    transition: opacity 0.2s;
  }
</style>
