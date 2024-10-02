<script lang="ts">
  export let inspector = true;
  export let bgClass = "bg-gray-100";
  export let width = 0;
  export let height = 0;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

  $: ({ width, height } = contentRect);
</script>

<main class="flex flex-col size-full overflow-hidden" bind:contentRect>
  {#if $$slots.header}
    <header class="bg-white w-full h-fit">
      <slot name="header" />
    </header>
  {/if}

  <div class="h-full {bgClass} w-full flex overflow-hidden p-2 pt-0">
    <div class="w-full h-full overflow-hidden">
      <slot name="body" />
    </div>
    {#if inspector}
      <slot name="inspector" />
    {/if}
  </div>
</main>
