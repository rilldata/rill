<script lang="ts">
  import * as Collapsible from "@rilldata/web-common/components/collapsible/index.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { ChevronDownIcon } from "lucide-svelte";
  import { writable } from "svelte/store";
  import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
  import { getInlineChatContextFilteredOptions } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
  import {
    type InlineChatContext,
    inlineChatContextsAreEqual,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

  export let conversationManager: ConversationManager;
  export let left: number;
  export let bottom: number;
  export let inlineChatContext: InlineChatContext | null = null;
  export let searchText: string = "";
  export let onSelect: (ctx: InlineChatContext) => void;

  const searchTextStore = writable("");
  $: searchTextStore.set(searchText.replace(/^@/, ""));

  const filteredOptions = getInlineChatContextFilteredOptions(
    searchTextStore,
    conversationManager,
  );
</script>

<!-- bits-ui dropdown component captures focus, so chat text cannot be edited when it is open.
       Newer versions of bits-ui have "trapFocus=false" param but it needs svelte5 upgrade.
       TODO: move to dropdown component after upgrade. -->
<div
  class="inline-chat-context-dropdown block"
  style="left: {left}px; bottom: {bottom}px;"
>
  {#each $filteredOptions as { metricsViewContext, recentlyUsed, currentlyActive, measures, dimensions } (metricsViewContext.values[0])}
    {@const metricsViewSelected =
      inlineChatContext !== null &&
      inlineChatContextsAreEqual(metricsViewContext, inlineChatContext)}
    <Collapsible.Root
      open={currentlyActive || recentlyUsed || $searchTextStore.length > 0}
      class="border-b last:border-b-0"
    >
      <Collapsible.Trigger asChild let:builder>
        <button
          class="context-item metrics-view-context-item"
          type="button"
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
        >
          <ChevronDownIcon size="12px" strokeWidth={4} />
          <input
            type="radio"
            checked={metricsViewSelected}
            on:click|stopPropagation={() => onSelect(metricsViewContext)}
            class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
          />
          <span class="text-sm grow">{metricsViewContext.label}</span>
          {#if recentlyUsed}
            <span class="metrics-view-context-label">Recent asked</span>
          {:else if currentlyActive}
            <span class="metrics-view-context-label">Current</span>
          {/if}
        </button>
      </Collapsible.Trigger>
      <Collapsible.Content class="flex flex-col ml-5 gap-y-0.5">
        {#each measures as measure (measure.values[1])}
          <button
            class="context-item"
            type="button"
            on:click={() => onSelect(measure)}
          >
            <div class="square"></div>
            <span>{measure.label}</span>
          </button>
        {/each}

        {#if measures.length > 0 && dimensions.length > 0}
          <div class="content-separator"></div>
        {/if}

        {#each dimensions as dimension (dimension.values[1])}
          <button
            class="context-item"
            type="button"
            on:click={() => onSelect(dimension)}
          >
            <div class="circle"></div>
            <span>{dimension.label}</span>
          </button>
        {/each}

        {#if measures.length === 0 && dimensions.length === 0}
          <div class="contents-empty">No dimensions or measures found</div>
        {/if}
      </Collapsible.Content>
    </Collapsible.Root>
  {:else}
    <div class="contents-empty">No matches found</div>
  {/each}
</div>

<style lang="postcss">
  .inline-chat-context-dropdown {
    @apply flex flex-col fixed p-1.5 z-50 w-[300px] max-h-[500px] overflow-auto;
    @apply rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .metrics-view-context-item {
    @apply font-semibold;
  }

  .metrics-view-context-label {
    @apply text-xs font-normal text-right text-popover-foreground/60;
  }

  .context-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1 w-full;
    @apply cursor-default select-none rounded-sm outline-none;
    @apply text-sm text-left text-wrap break-words;
  }
  .context-item:hover {
    @apply bg-accent text-accent-foreground cursor-pointer;
  }

  .square {
    @apply w-2 h-2 bg-theme-secondary-600;
  }
  .circle {
    @apply w-2 h-2 rounded-full bg-theme-secondary-600;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }

  .content-separator {
    @apply -mx-1 my-1 h-px bg-muted;
  }
</style>
