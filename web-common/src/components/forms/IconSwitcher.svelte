<script lang="ts">
  import type { ComponentType, SvelteComponent } from "svelte";

  export let fields: { id: string; Icon: ComponentType<SvelteComponent> }[];
  export let selected: string | undefined;
  export let onClick: (value: string) => void = () => {};
  export let small = false;
  export let expand = false;
</script>

<div class:small class:expand class="option-wrapper">
  {#each fields as { id, Icon } (id)}
    <button
      on:click={() => onClick(id)}
      class="-ml-[1px] first-of-type:-ml-0 px-2 border border-gray-300 first-of-type:rounded-l-[2px] last-of-type:rounded-r-[2px]"
      class:selected={selected === id}
    >
      <Icon size={small ? "14px" : "16px"} />
    </button>
  {/each}
</div>

<style lang="postcss">
  button {
    @apply flex justify-center items-center;
    @apply capitalize;
  }

  button:hover:not(.selected) {
    @apply bg-slate-50;
  }

  .option-wrapper {
    @apply flex h-6 text-sm w-fit mb-1 rounded-[2px];
  }

  .option-wrapper.small {
    @apply h-6 text-xs;
  }

  .expand {
    @apply w-full;
  }
  .expand button {
    @apply flex-1;
  }

  .option-wrapper > .selected {
    @apply border-primary-500 z-50 text-primary-500;
  }
</style>
