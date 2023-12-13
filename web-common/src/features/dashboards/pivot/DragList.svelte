<script lang="ts">
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { measureChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import { createEventDispatcher } from "svelte";

  export let items: unknown[] = [];
  export let style: "vertical" | "horizontal" = "vertical";

  const dispatch = createEventDispatcher();
  const flipDurationMs = 200;

  function handleConsider(e) {
    items = e.detail.items;
  }
  function handleFinalize(e) {
    items = e.detail.items;
    dispatch("update", items);
  }

  let listClasses;
  $: if (style === "horizontal") {
    listClasses = "flex flex-row bg-slate-50 w-full p-2 gap-x-2 h-10";
  } else {
    listClasses = "flex flex-col gap-y-2 py-2";
  }
</script>

<div
  class={listClasses}
  use:dndzone={{ items, flipDurationMs }}
  on:consider={handleConsider}
  on:finalize={handleFinalize}
>
  {#each items as item (item.id)}
    <div class="item" animate:flip={{ duration: flipDurationMs }}>
      <Chip
        outline={true}
        {...measureChipColors}
        extraPadding={false}
        extraRounded={false}
        label={item.title}
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
