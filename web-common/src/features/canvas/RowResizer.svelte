<script lang="ts">
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
</script>

<button
  type="button"
  class="row-resizer bg-primary-300"
  on:mousedown|preventDefault={(e) => {
    dispatch("resize", e);
  }}
/>

<style lang="postcss">
  .row-resizer {
    position: absolute;
    left: 0;
    right: 0;
    height: 4px;
    cursor: row-resize;
    z-index: 20;
    top: 50%;
    transform: translateY(-50%);
    pointer-events: auto;
  }

  .row-resizer:hover {
    height: 6px; /* Slightly increase height on hover for better visual feedback */
  }
</style>
