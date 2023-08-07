<script lang="ts">
  export let rowIdx: number;
  export let colIdx: number;
</script>

{#if rowIdx === -1}
  <div>Column {colIdx}</div>
{:else}
  <div>cell {rowIdx},{colIdx}</div>
{/if}
