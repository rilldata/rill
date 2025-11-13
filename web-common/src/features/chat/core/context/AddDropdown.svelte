<script lang="ts">
  import { getContextOptions } from "@rilldata/web-common/features/chat/core/context/context-options.ts";
  import {
    type ChatContextEntry,
    ChatContextEntryType,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { writable } from "svelte/store";

  export let left: number;
  export let bottom: number;
  export let onAdd: (ctx: ChatContextEntry) => void;

  type Option = { label: string; value: string };
  const searchTextStore = writable("");

  const contextTopLevelOptions = [
    {
      value: ChatContextEntryType.Measures,
      label: "Measures",
    },
    {
      value: ChatContextEntryType.Dimensions,
      label: "Dimensions",
    },
    {
      value: ChatContextEntryType.Explore,
      label: "Explores",
    },
  ];

  const contextOptions = getContextOptions(searchTextStore);

  let firstLevelSelection: ChatContextEntryType | null = null;
  $: firstLevelSelectedOption = contextTopLevelOptions.find(
    (o) => o.value === firstLevelSelection,
  );
  $: secondLevelOptions = firstLevelSelection
    ? $contextOptions[firstLevelSelection]
    : null;

  export function setText(newText: string) {
    searchTextStore.set(newText);
  }

  function selectMode(e, selection: ChatContextEntryType) {
    e.preventDefault();
    e.stopPropagation();
    firstLevelSelection = selection;
  }

  function onOptionSelect(option: Option) {
    if (!firstLevelSelectedOption) return;

    onAdd({
      type: firstLevelSelectedOption.value,
      value: option.value,
      label: option.label,
    });
  }
</script>

<div class="context-dropdown" style="left: {left}px; bottom: {bottom}px">
  {#if firstLevelSelection === null}
    {#each contextTopLevelOptions as option (option.value)}
      <button on:click={(e) => selectMode(e, option.value)}>
        <span>{option.label}</span>
        <span>{">"}</span>
      </button>
    {/each}
  {:else if firstLevelSelectedOption && secondLevelOptions}
    <div class="flex flex-row items-center">
      <button on:click={() => (firstLevelSelection = null)}>
        {"<"}
      </button>
      <span>{firstLevelSelectedOption.label}</span>
    </div>
    {#each secondLevelOptions as option (option.value)}
      <button on:click={() => onOptionSelect(option)}>
        {option.label}
      </button>
    {/each}
  {/if}
</div>

<style lang="postcss">
  .context-dropdown {
    @apply flex flex-col absolute p-1.5 z-50 w-[200px];
    @apply rounded-md bg-popover text-popover-foreground shadow-md;
  }
</style>
