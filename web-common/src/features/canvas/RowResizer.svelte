<script lang="ts">
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
</script>

<button
  type="button"
  class="row-resizer"
  on:mousedown|preventDefault={(e) => {
    dispatch("resize", e);
  }}
/>

<style>
  .row-resizer {
    position: absolute;
    left: 0;
    right: 0;
    height: 4px;
    background: var(--color-primary-500);
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
