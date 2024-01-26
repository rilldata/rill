<script lang="ts">
  export let row;
  export let value: string;
  export let assembled = true;
</script>

<div class="flex gap-x-1" style:padding-left={`${row.depth * 2}rem`}>
  {#if value === "LOADING_CELL"}
    <span class="loading-cell" />
  {:else if row.getCanExpand()}
    <button on:click={row.getToggleExpandedHandler()} class="cursor-pointer">
      {row.getIsExpanded() ? "▼" : "▶"}
    </button>
  {/if}
  {value === "LOADING_CELL" ? "" : value}
</div>

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }
</style>
