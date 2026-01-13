<script lang="ts">
  export let assembled;
  export let isShowMoreRow = false;
</script>

{#if isShowMoreRow}
  <!-- Render blank for show more rows -->
{:else if assembled}
  <span class="text-muted-foreground">-</span>
{:else}
  <span class="loading-cell" />
{/if}

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }
</style>
