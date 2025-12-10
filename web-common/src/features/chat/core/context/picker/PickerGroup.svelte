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

  export let parentOption: InlineContextPickerOption;
  export let selectedChatContext: InlineContext | null = null;
  export let highlightedContext: InlineContext | null = null;
  export let searchTextStore: Readable<string>;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  $: ({ context, recentlyUsed, currentlyActive, childContextCategories } =
    parentOption);

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
        on:click={focusEditor}
      >
        {#if open}
          <ChevronDownIcon size="12px" strokeWidth={4} />
        {:else}
          <ChevronRightIcon size="12px" strokeWidth={4} />
        {/if}
        <input
          type="radio"
          checked={parentOptionSelected}
          on:click|stopPropagation={() => onSelect(context)}
          class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
        />
        <span class="text-sm grow">{context.label}</span>
        {#if recentlyUsed}
          <span class="parent-context-label">Recent asked</span>
        {:else if currentlyActive}
          <span class="parent-context-label">Current</span>
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
          <button
            class="context-item"
            class:highlight={highlighted}
            type="button"
            on:click={() => onSelect(ctx)}
            use:ensureInView={highlighted}
          >
            <div class="context-item-checkbox">
              {#if selected}
                <CheckIcon size="12px" />
              {/if}
            </div>
            <div class="square"></div>
            <span>{ctx.label}</span>
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

  .context-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1 w-full;
    @apply cursor-default select-none rounded-sm outline-none;
    @apply text-sm text-left text-wrap break-words;
  }
  .context-item:hover {
    @apply bg-accent text-accent-foreground cursor-pointer;
  }
  .context-item.highlight {
    @apply bg-accent text-accent-foreground;
  }

  .context-item-checkbox {
    @apply w-3 h-3;
  }

  .square {
    @apply w-2 h-2 bg-theme-secondary-600/50;
  }
  .circle {
    @apply w-2 h-2 rounded-full bg-theme-500/50;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }

  .content-separator {
    @apply -mx-1 my-1 h-px bg-muted;
  }
</style>
