<script lang="ts">
  import { dndzone } from "svelte-dnd-action";
  import { flip } from "svelte/animate";

  export let items: unknown[] = [];
  export let style: "vertical" | "horizontal" = "vertical";

  const flipDurationMs = 200;

  function handleSort(e) {
    items = e.detail.items;
  }

  let listClasses;
  $: if (style === "horizontal") {
    listClasses = "flex flex-row bg-slate-100 w-full p-2";
  } else {
    listClasses = "flex flex-col px-2";
  }
</script>

<div
  class={listClasses}
  use:dndzone={{ items, flipDurationMs }}
  on:consider={handleSort}
  on:finalize={handleSort}
>
  {#each items as item (item.id)}
    <div class="item" animate:flip={{ duration: flipDurationMs }}>
      {item.title}
    </div>
  {/each}
</div>

<style type="postcss">
  .item {
    @apply text-center h-6;
    border: 1px solid black;
    margin: 0.2em;
    padding: 0.3em;
  }
</style>
