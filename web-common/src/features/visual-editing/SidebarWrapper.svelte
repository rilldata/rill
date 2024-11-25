<script lang="ts">
  export let title: string;
</script>

<div
  class="size-full bg-background flex-none flex flex-col select-none rounded-[2px] w-full"
>
  <h1>{title}</h1>

  <div
    class="px-5 flex flex-col gap-y-3 w-full h-full overflow-y-auto overflow-x-visible"
  >
    <slot />
  </div>

  <slot name="footer" />
</div>

<style lang="postcss">
  h1 {
    @apply text-lg font-semibold mb-2;
    @apply pt-6 px-5 mb-2;
  }

  /* h2 {
    @apply text-sm font-medium;
  } */
</style>
