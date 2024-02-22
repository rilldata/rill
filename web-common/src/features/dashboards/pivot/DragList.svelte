<script context="module" lang="ts">
  import AddField from "./AddField.svelte";
  import PivotChip from "./PivotChip.svelte";
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import { createEventDispatcher } from "svelte";
  import { PivotChipData, PivotChipType } from "./types";
  import { Writable, writable } from "svelte/store";

  const dragging: Writable<null | PivotChipType> = writable(null);
</script>

<script lang="ts">
  export let items: PivotChipData[] = [];
  export let placeholder: string | null = null;
  export let type: "rows" | "columns" | null = null;

  const removable = Boolean(type);
  const horizontal = Boolean(type);

  const dispatch = createEventDispatcher();
  const flipDurationMs = 200;

  $: valid =
    type &&
    $dragging &&
    (type === "columns" || $dragging !== PivotChipType.Measure);

  function handleConsider(e: CustomEvent<{ items: PivotChipData[] }>) {
    items = e.detail.items;
  }

  function handleFinalize(e: CustomEvent<{ items: PivotChipData[] }>) {
    items = e.detail.items;
    dispatch("update", items);
  }
</script>

<div
  class:valid
  class="flex flex-col gap-y-2 py-2 rounded-sm text-gray-500 w-full max-w-full"
  class:horizontal
  use:dndzone={{ items, flipDurationMs, dropFromOthersDisabled: !valid }}
  on:consider={handleConsider}
  on:finalize={handleFinalize}
>
  {#if !items.length && placeholder}
    {placeholder}
  {/if}
  {#each items as item (item.id)}
    <button
      on:mousedown={() => {
        dragging.set(item.type);

        const onMouseUp = () => {
          dragging.set(null);
          window.removeEventListener("mouseup", onMouseUp);
        };

        window.addEventListener("mouseup", onMouseUp);
      }}
      class="item"
      animate:flip={{ duration: flipDurationMs }}
    >
      <PivotChip
        {removable}
        {item}
        on:remove={() => {
          items = items.filter((i) => i.id !== item.id);
          dispatch("update", items);
        }}
      />
    </button>
  {/each}
  {#if removable}
    <AddField {type} />
  {/if}
</div>

<style type="postcss">
  .item {
    @apply text-center;
  }

  .horizontal {
    @apply flex flex-row flex-wrap bg-slate-50 w-full p-1 px-2 gap-x-2 h-fit;
    @apply items-center;
    @apply border border-slate-50;
  }

  div {
    outline: none !important;
  }

  .valid {
    @apply border-blue-400;
  }
</style>
