<script lang="ts">
  export let range: string | undefined;
  export let dark = false;
  export let onClick: ((range: string | undefined) => void) | undefined =
    undefined;
</script>

<svelte:element
  this={onClick ? "button" : "span"}
  class:dark
  role={onClick ? "button" : undefined}
  class="element"
  on:click={() => onClick?.(range)}
>
  {range}
</svelte:element>

<style lang="postcss">
  .element {
    @apply bg-surface-muted text-fg-secondary rounded-[2px] px-1 line-clamp-1 truncate flex-none h-5 flex items-center select-none cursor-default;
    font-family: "Source Code Variable", monospace;
    @apply font-medium w-fit;
  }

  .dark.element {
    @apply bg-gray-300 text-fg-inverse;
  }

  .element:hover {
    @apply bg-surface-hover;
  }
</style>
