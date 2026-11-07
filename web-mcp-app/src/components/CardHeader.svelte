<script lang="ts">
  let { title, caption }: { title: string; caption: string } = $props();
</script>

<header class="card-header">
  <h2 class="title" {title}>{title}</h2>
  {#if caption}
    <p class="caption" title={caption}>{caption}</p>
  {/if}
</header>

<style lang="postcss">
  .card-header {
    @apply flex flex-col gap-0.5 px-4 pt-3 pb-1 min-w-0;
  }
  .title {
    @apply m-0 text-[15px] font-semibold leading-5 text-fg-primary truncate;
  }
  .caption {
    @apply m-0 text-xs leading-4 text-fg-tertiary truncate;
  }
</style>
