<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let disabled = false;
  export let selected = false;
</script>

<button class="relative group" class:selected {disabled} on:click>
  <slot />
  {#if disabled}
    <div
      class="absolute top-full translate-y-2 z-10 w-fit -translate-x-1/2 left-1/2 group-hover:block hidden"
    >
      <TooltipContent>Coming Soon</TooltipContent>
    </div>
  {/if}
</button>

<style lang="postcss">
  button {
    @apply border-b-2 border-transparent;
    @apply flex items-center relative;
    @apply p-1 gap-x-2;
    @apply font-medium text-xs text-gray-500;
  }

  .selected {
    @apply border-b-2 border-primary-600 text-primary-600;
  }
</style>
