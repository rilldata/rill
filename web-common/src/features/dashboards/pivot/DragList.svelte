<script context="module" lang="ts">
  import AddField from "./AddField.svelte";
  import PivotChip from "./PivotChip.svelte";
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import { createEventDispatcher } from "svelte";
  import type { PivotChipData } from "./types";
</script>

<script lang="ts">
  export let items: PivotChipData[] = [];
  export let placeholder: string | null = null;
  export let type: "rows" | "columns" | null = null;

  const removable = Boolean(type);
  const horizontal = Boolean(type);

  const dispatch = createEventDispatcher();
  const flipDurationMs = 200;

  function handleConsider(e: CustomEvent<{ items: PivotChipData[] }>) {
    items = e.detail.items;
  }

  function handleFinalize(e: CustomEvent<{ items: PivotChipData[] }>) {
    items = e.detail.items;
    dispatch("update", items);
  }
</script>

<div
  class="flex flex-col gap-y-2 py-2 rounded-sm text-gray-500 w-full"
  class:horizontal
  use:dndzone={{ items, flipDurationMs }}
  on:consider={handleConsider}
  on:finalize={handleFinalize}
>
  {#if !items.length && placeholder}
    {placeholder}
  {/if}
  {#each items as item (item.id)}
    <div class="item" animate:flip={{ duration: flipDurationMs }}>
      <PivotChip
        {removable}
        {item}
        on:remove={() => {
          items = items.filter((i) => i.id !== item.id);
          dispatch("update", items);
        }}
      />
    </div>
  {/each}
  {#if removable}
    <AddField {type} />
  {/if}
</div>

<style type="postcss">
  .item {
    @apply text-center h-6;
  }

  .horizontal {
    @apply flex flex-row bg-slate-50 w-full p-2 gap-x-2 h-10;
    @apply items-center;
  }

  div {
    outline: none !important;
  }
</style>
