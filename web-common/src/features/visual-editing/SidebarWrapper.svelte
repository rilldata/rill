<script lang="ts">
  export let title: string;
  export let type: "primary" | "secondary" = "primary";
  export let disableHorizontalPadding = false;
</script>

<div
  class="size-full bg-background flex-none flex flex-col select-none rounded-[2px] w-full"
>
  <div class={type}>
    <h1 class:pb-2={type == "secondary" && !$$slots.header}>{title}</h1>
    <slot name="header" />
  </div>

  <div
    class="{disableHorizontalPadding
      ? ''
      : 'px-5'} flex flex-col gap-y-3 w-full h-full overflow-y-auto overflow-x-visible"
  >
    <slot />
  </div>

  <slot name="footer" />
</div>

<style lang="postcss">
  .primary {
    @apply text-lg font-semibold;
    @apply pt-6 px-5 mb-2;
  }

  .secondary {
    @apply flex flex-col gap-y-1;
    @apply text-sm font-semibold;
    @apply bg-slate-50;
    @apply px-5 pt-2;
  }
</style>
