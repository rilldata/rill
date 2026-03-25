<script lang="ts">
  export let title = "";
  export let maxWidth = 800;
  export let showTitle = true;
  export let flush = false;
</script>

<main
  class="bg-surface-base size-full pt-8 pb-16 lg:pt-12 flex flex-col items-center overflow-y-auto"
  class:padded={!flush}
  style="scrollbar-gutter: stable;"
>
  <section class="w-full flex flex-col gap-y-3" style:max-width="{maxWidth}px">
    {#if title && showTitle}
      <h1
        class="text-2xl text-fg-primary font-semibold"
        aria-label="Container title"
      >
        {title}
      </h1>
    {/if}

    <slot />
  </section>
</main>

<style lang="postcss">
  .padded {
    @apply px-8 sm:px-16 lg:px-32 2xl:px-40;
  }
</style>
