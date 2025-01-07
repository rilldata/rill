<script lang="ts">
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
</script>

<button
  type="button"
  aria-label="Resize row"
  class="row-resize-handle w-full h-[3px] cursor-row-resize bg-transparent hover:bg-primary-300 z-[50] opacity-0 hover:opacity-100 pointer-events-auto"
  on:mousedown|stopPropagation={(e) => dispatch("resize", e)}
/>

<style>
  .row-resize-handle {
    position: relative;
    transition: opacity 0.2s;
  }
</style>
