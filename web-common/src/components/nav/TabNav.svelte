<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let items: Array<{ label: string; value: string }>;
  export let selected: string;
  export let minWidth = "180px";

  const dispatch = createEventDispatcher<{ select: string }>();

  function handleSelect(value: string) {
    selected = value;
    dispatch("select", value);
  }
</script>

<div class="nav-items" style:min-width={minWidth}>
  {#each items as item (item.value)}
    <button
      on:click={() => handleSelect(item.value)}
      class="nav-item"
      class:selected={selected === item.value}
    >
      <span class="text-fg-primary">{item.label}</span>
    </button>
  {/each}
</div>

<style lang="postcss">
  .nav-items {
    @apply flex flex-col gap-y-2;
  }

  .nav-item {
    @apply p-2 flex gap-x-1 items-center;
    @apply rounded-sm;
    @apply text-sm font-medium;
  }

  .selected {
    @apply bg-surface-active;
  }

  .nav-item:hover {
    @apply bg-surface-hover;
  }
</style>
