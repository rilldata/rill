<script lang="ts">
  export let row;
  export let value: string;
</script>

<div class="flex gap-x-1" style:padding-left={`${row.depth * 2}rem`}>
  <!-- TODO: Change LOADING_CELL to undefined -->
  {#if value === "LOADING_CELL"}
    <span>🔄</span>
  {:else if row.getCanExpand()}
    <button on:click={row.getToggleExpandedHandler()} class="cursor-pointer">
      {row.getIsExpanded() ? "▼" : "▶"}
    </button>
  {/if}
  {value === "LOADING_CELL" ? "" : value}
</div>
