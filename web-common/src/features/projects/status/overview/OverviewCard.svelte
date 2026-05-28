<script lang="ts">
  export let title: string;
  export let viewAllHref: string | undefined = undefined;
</script>

<section class="section">
  <div class="section-header">
    <h3 class="section-title">
      <slot name="title">{title}</slot>
    </h3>
    {#if viewAllHref}
      <a href={viewAllHref} class="view-all">View all</a>
    {:else}
      <slot name="header-right" />
    {/if}
  </div>
  <slot />
</section>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .view-all {
    @apply text-xs text-primary-500 no-underline;
  }
  .view-all:hover {
    @apply text-primary-600;
  }
</style>
