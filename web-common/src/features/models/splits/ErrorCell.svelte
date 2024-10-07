<script lang="ts">
  export let error: string | undefined;
</script>

{#if error}
  <span class="text-red-600 font-medium">{error}</span>
{:else}
  <span>-</span>
{/if}
