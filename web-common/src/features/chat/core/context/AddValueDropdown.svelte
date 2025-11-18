<script lang="ts">
  import {
    type ChatContextEntry,
    type ContextMetadata,
    ContextTypeData,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { getContextDimensionValuesQueryOptions } from "@rilldata/web-common/features/chat/core/context/get-context-dimension-values.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";

  export let chatCtx: ChatContextEntry;
  export let left: number;
  export let bottom: number;
  export let metadata: ContextMetadata;
  export let onAdd: (ctx: ChatContextEntry) => void;

  type Option = { label: string; value: string };
  const searchTextStore = writable("");

  const chatCtxStore = writable(chatCtx);
  $: chatCtxStore.set(chatCtx);

  const dimensionValues = createQuery(
    getContextDimensionValuesQueryOptions(chatCtxStore, searchTextStore),
    queryClient,
  );
  $: dimensionValueOptions = $dimensionValues.data ?? [];

  export function setText(newText: string) {
    searchTextStore.set(newText);
  }

  function onOptionSelect(option: Option) {
    const newCtx = {
      ...chatCtx,
      subValue: option.value,
    };
    newCtx.label =
      ContextTypeData[newCtx.type]?.getLabel(newCtx, metadata) ?? option.value;
    onAdd(newCtx);
  }
</script>

<span>{chatCtx.label}</span>

<div class="context-dropdown" style="left: {left}px; bottom: {bottom}px">
  {#each dimensionValueOptions as option (option.value)}
    <button
      class="content-item"
      on:click={() => onOptionSelect(option)}
      type="button"
    >
      {option.label}
    </button>
  {/each}
</div>

<style lang="postcss">
  .context-dropdown {
    @apply flex flex-col absolute p-1.5 z-50 w-[200px];
    @apply rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .content-item {
    @apply flex flex-row items-center gap-x-2;
    @apply cursor-default select-none rounded-sm px-2 py-1.5 text-xs outline-none;
  }
</style>
