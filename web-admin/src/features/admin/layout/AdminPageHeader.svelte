<script lang="ts">
  export let title: string;
  export let description: string = "";
</script>

<div class="page-header">
  <h1 class="page-title">{title}</h1>
  {#if description}
    <p class="page-description">{description}</p>
  {/if}
</div>

<style lang="postcss">
  .page-header {
    @apply mb-6;
  }

  .page-title {
    @apply text-xl font-semibold text-slate-900 dark:text-slate-100;
  }

  .page-description {
    @apply text-sm text-slate-500 dark:text-slate-400 mt-1;
  }
</style>
