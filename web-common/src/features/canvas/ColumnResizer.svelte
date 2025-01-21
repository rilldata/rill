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
  class="col-resizer bg-primary-300"
  class:clicked={clicked || resizing}
  style="left: {left}px; height: {height}px;"
  on:mousedown|stopPropagation={(e) => dispatch("resize", e)}
  on:click|stopPropagation={(e) => dispatch("click", e)}
/>

<style lang="postcss">
  .col-resizer {
    position: absolute;
    width: 3px;
    cursor: col-resize;
    z-index: 20;
    pointer-events: auto;
    opacity: 0;
    transition:
      opacity 0.2s,
      width 0.2s;
  }

  .col-resizer:hover,
  .col-resizer.clicked {
    opacity: 1;
    width: 4px;
  }
</style>
