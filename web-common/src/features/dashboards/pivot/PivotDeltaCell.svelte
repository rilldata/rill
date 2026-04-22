<script lang="ts">
  export let assembled: boolean;
  export let formattedValue: string;
  export let value: number | null | undefined;
  export let inverseTheme = false;
</script>

{#if assembled}
  {#if value !== null && value !== undefined}
    <span
      class="pointer-events-none {value > 0
        ? inverseTheme
          ? 'text-kpi-negative'
          : 'text-kpi-positive'
        : value < 0
          ? inverseTheme
            ? 'text-kpi-positive'
            : 'text-kpi-negative'
          : 'text-fg-secondary'}"
    >
      {formattedValue}
    </span>
  {:else}
    <span class="text-fg-secondary pointer-events-none">-</span>
  {/if}
{:else}
  <span class="loading-cell"></span>
{/if}

<style lang="postcss">
  .loading-cell {
    @apply h-2 bg-surface-subtle rounded w-full inline-block;
  }
</style>
