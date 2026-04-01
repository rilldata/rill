<script lang="ts">
  export let count: number;
</script>

{#if count > 0}
  <span class="text-red-600 font-medium text-xs">{count}</span>
{:else}
  <span class="text-fg-tertiary">—</span>
{/if}
