<script lang="ts">
  import {
    getIdForContext,
    type InlineContext,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import type { KeyboardNavigationManager } from "@rilldata/web-common/features/chat/core/context/picker/keyboard-navigation.ts";
  import { InlineContextConfig } from "@rilldata/web-common/features/chat/core/context/config.ts";
  import { CheckIcon } from "lucide-svelte";
  import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
  import { getInlineChatContextMetadata } from "@rilldata/web-common/features/chat/core/context/metadata.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let item: PickerItem;
  export let selectedChatContext: InlineContext | null;
  export let keyboardNavigationManager: KeyboardNavigationManager;
  export let onSelect: (ctx: InlineContext) => void;

  const runtimeClient = useRuntimeClient();

  const typeConfig = InlineContextConfig[item.context.type];
  const contextMetadataStore = getInlineChatContextMetadata(runtimeClient);
  $: icon = typeConfig?.getIcon?.(item.context, $contextMetadataStore);
  const selectedItemId = selectedChatContext
    ? getIdForContext(selectedChatContext)
    : null;

  const { focusedItemStore, enhancePickerNode } = keyboardNavigationManager;
  $: focusedItem = $focusedItemStore;
  $: focused = focusedItem?.id === item.id;
  $: selected = item.id === selectedItemId;
</script>

<button
  class="context-item"
  class:focused
  type="button"
  on:click={() => onSelect(item.context)}
  use:enhancePickerNode={item}
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

  <span class="context-item-label">{item.context.label}</span>
</button>

<style lang="postcss">
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
    @apply bg-surface-subtle text-fg-primary;
  }

  .context-item-checkbox {
    @apply min-w-3 h-3;
  }

  .context-item-icon {
    @apply min-w-3.5 h-2;
  }
</style>
