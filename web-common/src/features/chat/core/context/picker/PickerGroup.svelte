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
  import type { InlineContextPickerParentOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
  import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";
  import { InlineContextConfig } from "@rilldata/web-common/features/chat/core/context/config.ts";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";

  export let parentOption: InlineContextPickerParentOption;
  export let selectedChatContext: InlineContext | null = null;
  export let highlightManager: PickerOptionsHighlightManager;
  export let searchTextStore: Readable<string>;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  $: ({
    context,
    openStore,
    recentlyUsed,
    currentlyActive,
    children,
    childrenLoading,
  } = parentOption);
  $: typeConfig = InlineContextConfig[context.type];
  $: resolvedChildren = children ?? [];

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

  $: parentTypeLabel = typeConfig.typeLabel;
  $: parentIcon = typeConfig.getIcon?.(context);

  $: mouseContextHighlightHandler = highlightManager.mouseOverHandler(context);

  $: shouldForceOpen =
    withinParentOptionSelected ||
    withinParentOptionHighlighted ||
    recentlyUsed ||
    currentlyActive ||
    $searchTextStore.length > 0;
  function forceOpen() {
    openStore.set(true);
  }
  $: if (shouldForceOpen) forceOpen();

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

<Collapsible.Root bind:open={$openStore}>
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
      <input
        type="radio"
        checked={parentOptionSelected}
        on:click|stopPropagation={() => onSelect(context)}
        class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
      />
      <div class="min-w-3.5">
        {#if $openStore}
          <ChevronDownIcon size="12px" strokeWidth={4} />
        {:else if parentIcon}
          <!-- On hover show chevron right icon -->
          <svelte:component this={parentIcon} size="12px" />
        {:else}
          <ChevronRightIcon size="12px" strokeWidth={4} />
        {/if}
      </div>

      <span class="context-item-label">{context.label}</span>

      {#if parentOptionHighlighted}
        <div class="context-item-keyboard-shortcut">↑/↓</div>
        {#if parentTypeLabel}
          <div class="context-item-type-label">
            {parentTypeLabel}
          </div>
        {/if}
      {:else if recentlyUsed}
        <span class="parent-context-label">Recently asked</span>
      {:else if currentlyActive}
        <span class="parent-context-label">Current</span>
      {/if}
    </button>
  </Collapsible.Trigger>
  <Collapsible.Content class="flex flex-col ml-0.5 gap-y-0.5">
    {#if childrenLoading}
      <DelayedSpinner isLoading={childrenLoading} />
    {:else}
      {#each resolvedChildren as childCategory (childCategory.type)}
        {#each childCategory.options as child (child.value)}
          {@const selected =
            selectedChatContext !== null &&
            inlineContextsAreEqual(child, selectedChatContext)}
          {@const highlighted =
            highlightedContext !== null &&
            inlineContextsAreEqual(child, highlightedContext)}
          {@const mouseContextHighlightHandler =
            highlightManager.mouseOverHandler(child)}
          {@const icon = InlineContextConfig[child.type]?.getIcon?.(child)}

          <button
            class="context-item"
            class:highlight={highlighted}
            type="button"
            on:click={() => onSelect(child)}
            use:ensureInView={highlighted}
            use:mouseContextHighlightHandler
          >
            <div class="context-item-checkbox">
              {#if selected}
                <CheckIcon size="12px" />
              {/if}
            </div>
            {#if icon}
              <div class="text-gray-500">
                <svelte:component this={icon} size="16px" />
              </div>
            {:else}
              <div class="context-item-icon"></div>
            {/if}

            <span class="context-item-label">{child.label}</span>

            {#if highlighted}
              <div class="context-item-keyboard-shortcut">↑/↓</div>
            {/if}
          </button>
        {/each}
      {:else}
        <div class="contents-empty">No matches found</div>
      {/each}
    {/if}
  </Collapsible.Content>
</Collapsible.Root>

<style lang="postcss">
  .parent-context-item {
    @apply font-semibold;
  }

  .parent-context-label {
    @apply min-w-24 text-xs font-normal text-right text-popover-foreground/60;
  }

  .context-item-label {
    @apply basis-full grow shrink;
    @apply text-sm overflow-hidden whitespace-nowrap text-ellipsis;
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
    @apply grow-0 shrink-0;
    @apply text-accent-foreground/60;
  }

  .context-item-icon {
    @apply min-w-3.5 h-2;
  }

  .context-item-type-label {
    @apply grow-0 shrink-0;
    @apply text-xs font-normal text-right text-popover-foreground/60;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }
</style>
