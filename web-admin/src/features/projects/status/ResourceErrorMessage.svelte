<script lang="ts">
  export let message: string;
</script>

{#if message}
  <span class="text-red-600 max-w-xs line-clamp-2">
    {message}
  </span>
{:else}
  <span>-</span>
{/if}
