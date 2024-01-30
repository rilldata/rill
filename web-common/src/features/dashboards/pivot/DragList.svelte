<script context="module" lang="ts">
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    measureChipColors,
    timeChipColors,
    defaultChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import { createEventDispatcher } from "svelte";
  import type { ChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";

  const colors: Record<PivotChipType, ChipColors> = {
    Time: timeChipColors,
    Measure: measureChipColors,
    Dimension: defaultChipColors,
  };
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

  function handleConsider(e) {
    items = e.detail.items;
  }
  function handleFinalize(e) {
    items = e.detail.items;
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
      <Chip
        outline
        {removable}
        {...colors[item.type]}
        extraPadding={false}
        extraRounded={item.type !== PivotChipType.Measure}
        label={item.title}
        on:remove={() => {
          items = items.filter((i) => i.id !== item.id);
          dispatch("update", items);
        }}
      >
        <div slot="body" class="font-semibold">{item.title}</div>
      </Chip>
    </div>
  {/each}
</div>

<style type="postcss">
  .item {
    @apply text-center h-6;
  }
</style>
