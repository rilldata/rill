<script lang="ts">
  export let title: string;

  export let width: number | undefined = undefined;
</script>

<div
  style:width="{width}px"
  class="border size-full bg-background flex-none flex flex-col select-none rounded-[2px]"
>
  <h1 class="pt-6 px-5">{title}</h1>

  <div
    class="px-5 flex flex-col gap-y-3 w-full h-full overflow-y-auto overflow-x-visible"
  >
    <slot />
  </div>

  <slot name="bottom" />

  <slot name="footer" />
</div>

<style lang="postcss">
  h1 {
    @apply text-lg font-semibold mb-4;
  }

  /* h2 {
    @apply text-sm font-medium;
  } */
</style>
