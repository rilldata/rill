<script lang="ts">
  let { count }: { count: number } = $props();
</script>

{#if count > 0}
  <span class="text-red-600 font-medium text-xs">{count}</span>
{:else}
  <span class="text-fg-tertiary">—</span>
{/if}
