<script lang="ts">
  export let assembled: boolean;
  export let formattedValue: string;
  export let value: number | null | undefined;
</script>

{#if assembled}
  {#if value !== null && value !== undefined}
    <span
      class="pointer-events-none {value > 0 ? 'text-gray-500' : 'text-red-500'}"
    >
      {formattedValue}
    </span>
  {:else}
    <span class="text-gray-400 pointer-events-none">-</span>
  {/if}
{:else}
  <span class="loading-cell" />
{/if}

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-gray-200 rounded w-full inline-block;
  }
</style>
