<script context="module" lang="ts">
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import { createEventDispatcher } from "svelte";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
  import PivotChip from "./PivotChip.svelte";
  import type { TimeGrain } from "@rilldata/web-common/lib/time/types";
</script>

<script lang="ts">
  export let items: PivotChipData[] = [];
  export let style: "vertical" | "horizontal" = "vertical";
  export let removable = false;

  const dispatch = createEventDispatcher();
  const flipDurationMs = 200;

  let listClasses: string;

  $: if (style === "horizontal") {
    listClasses = "flex flex-row bg-slate-50 w-full p-2 gap-x-2 h-10";
  } else {
    listClasses = "flex flex-col gap-y-2 py-2";
  }

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
  class="{listClasses} rounded-sm"
  use:dndzone={{ items, flipDurationMs }}
  on:consider={handleConsider}
  on:finalize={handleFinalize}
>
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
</div>

<style type="postcss">
  .item {
    @apply text-center h-6;
  }

  div {
    outline: none !important;
  }
</style>
