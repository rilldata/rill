<script lang="ts">
  export let inspector = true;
  export let bgClass = "bg-slate-50";
  export let width = 0;
  export let height = 0;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

  $: ({ width, height } = contentRect);
</script>

<main
  class="flex flex-col size-full overflow-hidden {bgClass}"
  bind:contentRect
>
  {#if $$slots.header}
    <header class="w-full h-fit">
      <slot name="header" />
    </header>
  {/if}

  <div class="h-full w-full flex overflow-hidden p-4 pt-0">
    <div class="w-full h-full overflow-hidden">
      <slot name="body" />
    </div>
    {#if inspector}
      <slot name="inspector" />
    {/if}
  </div>
</main>
