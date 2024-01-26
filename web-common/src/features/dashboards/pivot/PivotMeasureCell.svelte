<script lang="ts">
  export let assembled;
</script>

{#if assembled}
  <span class="no-data-cell">no data</span>
{:else}
  <span class="loading-cell" />
{/if}

<style lang="postcss">
  .no-data-cell {
    @apply italic text-gray-400;
    font-size: 0.925em;
  }

  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }
</style>
