<script lang="ts">
  import { getContextOptions } from "@rilldata/web-common/features/chat/core/context/context-options.ts";
  import {
    type ChatContextEntry,
    ChatContextEntryType,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { writable } from "svelte/store";

  export let chatCtx: ChatContextEntry | null = null;
  export let left: number;
  export let bottom: number;
  export let onAdd: (ctx: ChatContextEntry) => void;
  export let focusEditor: () => void;

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
  $: shouldShowTopLevelOptions = chatCtx === null;

  const chatCtxStore = writable(chatCtx);
  $: chatCtxStore.set(chatCtx);
  const contextOptions = getContextOptions(chatCtxStore, searchTextStore);

  let firstLevelSelection: ChatContextEntryType | null = chatCtx?.type ?? null;
  $: isDimensionValueMode =
    firstLevelSelection === ChatContextEntryType.DimensionValues &&
    chatCtx !== null;
  $: firstLevelSelectedOption = isDimensionValueMode
    ? { label: `${chatCtx!.label}`, value: firstLevelSelection }
    : contextTopLevelOptions.find((o) => o.value === firstLevelSelection);
  $: secondLevelOptions = firstLevelSelection
    ? $contextOptions[firstLevelSelection]
    : null;

  export function setText(newText: string) {
    searchTextStore.set(newText);
  }

  export function selectFirst() {
    const option = secondLevelOptions?.[0];
    if (!option) return;
    onOptionSelect(option);
  }

  function unselectMode() {
    if (!shouldShowTopLevelOptions) return;

    firstLevelSelection = null;
    setTimeout(focusEditor);
  }

  function selectMode(e, selection: ChatContextEntryType) {
    e.preventDefault();
    e.stopPropagation();
    firstLevelSelection = selection;

    setTimeout(focusEditor);
  }

  function onOptionSelect(option: Option) {
    if (!firstLevelSelectedOption) return;

    const newCtx = {
      type: firstLevelSelectedOption.value,
      value: option.value,
      subValue: null,
      label: option.label,
    } as ChatContextEntry;

    if (isDimensionValueMode) {
      newCtx.value = chatCtx!.value!;
      newCtx.subValue = option.value;
      newCtx.label = `${chatCtx!.label}: ${option.label}`;
    }

    onAdd(newCtx);
  }
</script>

<!-- bits-ui dropdown component captures focus, so chat text cannot be edited when it is open.
     Newer versions of bits-ui have "trapFocus=false" param but it needs svelte5 upgrade.
     TODO: move to dropdown component after upgrade. -->
<div class="context-dropdown" style="left: {left}px; bottom: {bottom}px">
  {#if firstLevelSelection === null}
    {#each contextTopLevelOptions as option (option.value)}
      <button
        class="content-item content-item-selectable"
        on:click={(e) => selectMode(e, option.value)}
        type="button"
      >
        <span>{option.label}</span>
        <span>{">"}</span>
      </button>
    {/each}
  {:else if firstLevelSelectedOption && secondLevelOptions}
    <button
      class="content-item font-semibold"
      on:click={unselectMode}
      type="button"
    >
      {#if shouldShowTopLevelOptions}
        {"<"}
      {/if}
      {firstLevelSelectedOption.label}
    </button>
    <div class="content-separator"></div>
    {#each secondLevelOptions as option (option.value)}
      <button
        class="content-item content-item-selectable"
        on:click={() => onOptionSelect(option)}
        type="button"
      >
        {option.label}
      </button>
    {:else}
      <div class="contents-empty">No results</div>
    {/each}
  {/if}
</div>

<style lang="postcss">
  .context-dropdown {
    @apply flex flex-col absolute p-1.5 z-50 w-[200px];
    @apply rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .content-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1.5 w-full;
    @apply cursor-default select-none rounded-sm outline-none;
    @apply text-xs text-left text-wrap break-words;
  }
  .content-item-selectable:hover {
    @apply bg-accent text-accent-foreground cursor-pointer;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }

  .content-separator {
    @apply -mx-1 my-1 h-px bg-muted;
  }
</style>
