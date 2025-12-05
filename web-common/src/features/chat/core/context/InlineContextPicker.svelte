<script lang="ts">
  import { writable } from "svelte/store";
  import {
    getInlineChatContextFilteredOptions,
    type MetricsViewContextOption,
  } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
  import { type InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import MetricsViewGroup from "@rilldata/web-common/features/chat/core/context/MetricsViewGroup.svelte";
  import { InlineContextHighlightManager } from "@rilldata/web-common/features/chat/core/context/inline-context-highlight-manager.ts";

  export let left: number;
  export let bottom: number;
  export let selectedChatContext: InlineContext | null = null;
  export let searchText: string = "";
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  const searchTextStore = writable("");
  $: searchTextStore.set(searchText.replace(/^@/, ""));

  const filteredOptions = getInlineChatContextFilteredOptions(searchTextStore);

  const highlightManager = new InlineContextHighlightManager();
  const highlightedContext = highlightManager.highlightedContext;
  function handleMetricsViewOptionsChanged(
    filteredOptions: MetricsViewContextOption[],
  ) {
    highlightManager.filterOptionsUpdated(filteredOptions);
    // Auto highlight the currently selected context if it is present.
    if (selectedChatContext) {
      highlightManager.highlightContext(selectedChatContext);
    }
  }
  $: handleMetricsViewOptionsChanged($filteredOptions);

  function handleKeyDown(event: KeyboardEvent) {
    switch (event.key) {
      case "ArrowUp":
        highlightManager.highlightPreviousContext();
        break;
      case "ArrowDown":
        highlightManager.highlightNextContext();
        break;
      case "Enter":
        if ($highlightedContext) onSelect($highlightedContext);
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeyDown} />

<!-- bits-ui dropdown component captures focus, so chat text cannot be edited when it is open.
     Newer versions of bits-ui have "trapFocus=false" param but it needs svelte5 upgrade.
     TODO: move to dropdown component after upgrade. -->
<div
  class="inline-chat-context-dropdown block"
  style="left: {left}px; bottom: {bottom}px;"
>
  {#each $filteredOptions as metricsViewContextOption (metricsViewContextOption.metricsViewContext.metricsView)}
    <MetricsViewGroup
      {metricsViewContextOption}
      {selectedChatContext}
      highlightedContext={$highlightedContext}
      {searchTextStore}
      {onSelect}
      {focusEditor}
    />
  {:else}
    <div class="contents-empty">No matches found</div>
  {/each}
</div>

<style lang="postcss">
  .inline-chat-context-dropdown {
    @apply flex flex-col fixed p-1.5 z-50 w-[300px] max-h-[500px] overflow-auto;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }
</style>
