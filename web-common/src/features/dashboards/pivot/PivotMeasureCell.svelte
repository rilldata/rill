<script lang="ts">
  export let assembled;
</script>

{#if assembled}
  <span class="text-gray-400">-</span>
{:else}
  <span class="loading-cell" />
{/if}

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }
</style>
