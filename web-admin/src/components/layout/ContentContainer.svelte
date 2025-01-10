<script lang="ts">
  export let title = "";
  export let maxWidth = 800;
</script>

<main
  class="size-full p-8 lg:pt-12 flex flex-col items-center px-8 sm:px-16 lg:px-32 2xl:px-40 overflow-y-auto"
>
  <section class="w-full flex flex-col gap-y-3" style:max-width="{maxWidth}px">
    {#if title}
      <h1 class="text-2xl font-semibold">
        {title}
      </h1>
    {/if}

    <slot />
  </section>
</main>
