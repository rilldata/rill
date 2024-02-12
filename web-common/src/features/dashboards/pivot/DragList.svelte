<script context="module" lang="ts">
  import AddField from "./AddField.svelte";
  import PivotChip from "./PivotChip.svelte";
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import { createEventDispatcher } from "svelte";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
  import type { TimeGrain } from "@rilldata/web-common/lib/time/types";
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

  function onSelectTimeGrain(item: PivotChipData, timeGrain: TimeGrain) {
    items = items.map((i) => {
      if (i.id !== item.id) return i;

      return {
        id: timeGrain.grain,
        title: timeGrain.label,
        type: PivotChipType.Time,
      };
    });

    dispatch("update", items);
  }
</script>

<div
  class="container"
  class:horizontal
  use:dndzone={{ items, flipDurationMs }}
  on:consider={handleConsider}
  on:finalize={handleFinalize}
>
  {#if !items.length && placeholder}
    <p class="text-gray-500">{placeholder}</p>
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
        on:select-time-grain={(e) => {
          onSelectTimeGrain(item, e.detail.timeGrain);
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

  .container {
    @apply flex flex-col gap-y-2 py-2 rounded-sm;
  }

  .horizontal {
    @apply flex flex-row bg-slate-50 w-full p-2 gap-x-2 h-10;
    @apply items-center;
  }

  div {
    outline: none !important;
  }
</style>
