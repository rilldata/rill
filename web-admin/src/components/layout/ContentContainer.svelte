<script lang="ts">
  export let title = "";
  export let maxWidth = 800;
  export let showTitle = true;
</script>

<main
  class="bg-surface-elevated size-full pt-8 pb-16 lg:pt-12 flex flex-col items-center px-8 sm:px-16 lg:px-32 2xl:px-40 overflow-y-auto"
>
  <section class="w-full flex flex-col gap-y-3" style:max-width="{maxWidth}px">
    {#if title && showTitle}
      <h1
        class="text-2xl text-fg-secondary font-semibold"
        aria-label="Container title"
      >
        {title}
      </h1>
    {/if}

    <slot />
  </section>
</main>
