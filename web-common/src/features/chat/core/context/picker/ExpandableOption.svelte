<script lang="ts">
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import { getAttrs, builderActions } from "bits-ui";
  import { ChevronDownIcon, ChevronRightIcon } from "lucide-svelte";
  import type { Readable } from "svelte/store";
  import type { PickerTreeNode } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
  import type { KeyboardNavigationManager } from "@rilldata/web-common/features/chat/core/context/picker/keyboard-navigation.ts";
  import {
    getIdForContext,
    type InlineContext,
    inlineContextIsWithin,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import { InlineContextConfig } from "@rilldata/web-common/features/chat/core/context/config.ts";
  import type { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import SimpleOption from "@rilldata/web-common/features/chat/core/context/picker/SimpleOption.svelte";

  export let node: PickerTreeNode;
  export let selectedChatContext: InlineContext | null;
  export let keyboardNavigationManager: KeyboardNavigationManager;
  export let uiState: ContextPickerUIState;
  export let searchTextStore: Readable<string>;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  const item = node.item;
  const context = item.context;

  const typeConfig = InlineContextConfig[context.type];
  const typeLabel = typeConfig.typeLabel;
  const icon = typeConfig?.getIcon?.(context);
  const selectedItemId = selectedChatContext
    ? getIdForContext(selectedChatContext)
    : null;

  const { focusedItemStore, enhancePickerNode } = keyboardNavigationManager;
  $: focusedItem = $focusedItemStore;
  $: focused = focusedItem?.id === item.id;
  $: selected = item.id === selectedItemId;

  $: childSelected =
    selectedChatContext !== null &&
    inlineContextIsWithin(context, selectedChatContext);
  $: shouldForceOpen =
    childSelected ||
    item.recentlyUsed ||
    item.currentlyActive ||
    $searchTextStore.length > 0;
  $: if (shouldForceOpen) uiState.expand(item.id);

  const openStore = uiState.getExpandedStore(item.id);

  function onClick() {
    uiState.toggle(item.id);
    focusEditor();
  }
</script>

<Collapsible.Root open={$openStore}>
  <Collapsible.Trigger asChild let:builder>
    <button
      class="context-item parent-context-item"
      class:focused
      type="button"
      {...getAttrs([builder])}
      use:builderActions={{ builders: [builder] }}
      use:enhancePickerNode={item}
      on:click={onClick}
    >
      <input
        type="radio"
        checked={selected}
        on:click|stopPropagation={() => onSelect(context)}
        class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
      />
      <div class="min-w-3.5">
        {#if $openStore && node.children.length}
          <ChevronDownIcon size="12px" strokeWidth={4} />
        {:else if icon}
          <svelte:component this={icon} size="12px" />
        {:else}
          <ChevronRightIcon size="12px" strokeWidth={4} />
        {/if}
      </div>

      <span class="context-item-label">{context.label}</span>

      {#if focused}
        {#if typeLabel}
          <div class="context-item-type-label">
            {typeLabel}
          </div>
        {/if}
      {:else if item.recentlyUsed}
        <span class="parent-context-label">Recently asked</span>
      {:else if item.currentlyActive}
        <span class="parent-context-label">Current</span>
      {/if}
    </button>
  </Collapsible.Trigger>

  <Collapsible.Content class="flex flex-col ml-0.5 gap-y-0.5">
    {#if item.childrenLoading}
      <DelayedSpinner isLoading={item.childrenLoading} />
    {:else}
      {#each node.children as child (child.item.id)}
        {#if child.item.hasChildren}
          <svelte:self
            node={child}
            {selectedChatContext}
            {keyboardNavigationManager}
            {uiState}
            {searchTextStore}
            {onSelect}
            {focusEditor}
          />
        {:else}
          <SimpleOption
            item={child.item}
            {selectedChatContext}
            {keyboardNavigationManager}
            {onSelect}
          />
        {/if}
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
  .context-item.focused {
    @apply bg-accent text-accent-foreground;
  }

  .context-item-type-label {
    @apply grow-0 shrink-0;
    @apply text-xs font-normal text-right text-popover-foreground/60;
  }
</style>
