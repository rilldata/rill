<script lang="ts">
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import { getAttrs, builderActions } from "bits-ui";
  import {
    type InlineContext,
    inlineContextIsWithin,
    inlineContextsAreEqual,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import type { Readable } from "svelte/store";
  import { CheckIcon, ChevronDownIcon, ChevronRightIcon } from "lucide-svelte";
  import type { InlineContextPickerOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
  import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";

  export let parentOption: InlineContextPickerOption;
  export let selectedChatContext: InlineContext | null = null;
  export let highlightManager: PickerOptionsHighlightManager;
  export let searchTextStore: Readable<string>;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  $: ({ context, recentlyUsed, currentlyActive, childContextCategories } =
    parentOption);
  const highlightedContextStore = highlightManager.highlightedContext;
  $: highlightedContext = $highlightedContextStore;

  $: parentOptionSelected =
    selectedChatContext !== null &&
    inlineContextsAreEqual(context, selectedChatContext);
  $: withinParentOptionSelected =
    selectedChatContext !== null &&
    inlineContextIsWithin(context, selectedChatContext);

  $: parentOptionHighlighted =
    highlightedContext !== null &&
    inlineContextsAreEqual(context, highlightedContext);
  $: withinParentOptionHighlighted =
    highlightedContext !== null &&
    inlineContextIsWithin(context, highlightedContext);

  $: mouseContextHighlightHandler = highlightManager.mouseOverHandler(context);

  let open =
    withinParentOptionSelected ||
    parentOption.recentlyUsed ||
    parentOption.currentlyActive;
  $: shouldForceOpen =
    withinParentOptionHighlighted || $searchTextStore.length > 0;
  $: if (shouldForceOpen) {
    open = true;
  }

  function ensureInView(node: HTMLElement, active: boolean) {
    if (active) {
      node.scrollIntoView({ block: "nearest" });
    }
    return {
      update(active: boolean) {
        if (active) {
          node.scrollIntoView({ block: "nearest" });
        }
      },
    };
  }
</script>

{#if childContextCategories}
  <Collapsible.Root bind:open class="border-b last:border-b-0">
    <Collapsible.Trigger asChild let:builder>
      <button
        class="context-item parent-context-item"
        class:highlight={parentOptionHighlighted}
        type="button"
        {...getAttrs([builder])}
        use:builderActions={{ builders: [builder] }}
        use:ensureInView={parentOptionHighlighted}
        use:mouseContextHighlightHandler
        on:click={focusEditor}
      >
        <div class="min-w-3.5">
          {#if open}
            <ChevronDownIcon size="12px" strokeWidth={4} />
          {:else}
            <ChevronRightIcon size="12px" strokeWidth={4} />
          {/if}
        </div>
        <input
          type="radio"
          checked={parentOptionSelected}
          on:click|stopPropagation={() => onSelect(context)}
          class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
        />
        <span class="context-item-label">{context.label}</span>
        <div
          class="context-item-keyboard-shortcut"
          class:hidden={!parentOptionHighlighted}
        >
          ↑/↓
        </div>
        {#if recentlyUsed}
          <span class="metrics-view-context-label">Recently asked</span>
        {:else if currentlyActive}
          <span class="metrics-view-context-label">Current</span>
        {/if}
      </button>
    </Collapsible.Trigger>
    <Collapsible.Content class="flex flex-col ml-0.5 gap-y-0.5">
      {#each childContextCategories as childContextCategory, i}
        {#if i !== 0}<div class="content-separator"></div>{/if}

        {#each childContextCategory as ctx}
          {@const selected =
            selectedChatContext !== null &&
            inlineContextsAreEqual(ctx, selectedChatContext)}
          {@const highlighted =
            highlightedContext !== null &&
            inlineContextsAreEqual(ctx, highlightedContext)}
          {@const mouseContextHighlightHandler =
            highlightManager.mouseOverHandler(ctx)}

          <button
            class="context-item"
            class:highlight={highlighted}
            type="button"
            on:click={() => onSelect(ctx)}
            use:ensureInView={highlighted}
            use:mouseContextHighlightHandler
          >
            <div class="context-item-checkbox">
              {#if selected}
                <CheckIcon size="12px" />
              {/if}
            </div>
            <div class="square"></div>
            <span class="context-item-label">{ctx.label}</span>
            <div
              class="context-item-keyboard-shortcut"
              class:hidden={!highlighted}
            >
              ↑/↓
            </div>
          </button>
        {/each}
      {/each}
    </Collapsible.Content>
  </Collapsible.Root>
{/if}

<style lang="postcss">
  .parent-context-item {
    @apply font-semibold;
  }

  .parent-context-label {
    @apply text-xs font-normal text-right text-popover-foreground/60;
  }

  .context-item-label {
    @apply text-sm grow;
    @apply overflow-hidden whitespace-nowrap text-ellipsis;
  }

  .context-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1 w-full;
    @apply cursor-default select-none rounded-sm outline-none;
    @apply text-sm text-left text-wrap break-words;
  }
  .context-item:hover {
    @apply cursor-pointer;
  }
  .context-item.highlight {
    @apply bg-accent text-accent-foreground;
  }

  .context-item-checkbox {
    @apply min-w-3 h-3;
  }

  .context-item-keyboard-shortcut {
    @apply min-w-9 text-accent-foreground/60;
  }

  .square {
    @apply min-w-2 h-2 bg-theme-secondary-600/50;
  }
  .circle {
    @apply min-w-2 h-2 rounded-full bg-theme-500/50;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }

  .content-separator {
    @apply -mx-1 my-1 h-px bg-muted;
  }
</style>
