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
  // Like a select's `multiple`: in a single-choice picker an option is only picked, never shown as
  // selected, so there is no room for the check.
  export let multiple = true;

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
  class:with-description={!!item.description}
  type="button"
  onclick={() => onSelect(item.context)}
  use:enhancePickerNode={item}
>
  {#if multiple}
    <div class="context-item-checkbox">
      {#if selected}
        <CheckIcon size="12px" />
      {/if}
    </div>
  {/if}
  {#if icon}
    <div class="text-gray-500">
      <svelte:component this={icon} size="16px" />
    </div>
  {:else}
    <div class="context-item-icon"></div>
  {/if}

  {#if item.description}
    <div class="context-item-text">
      <span class="context-item-label">{item.context.label}</span>
      <span class="context-item-description">{item.description}</span>
    </div>
  {:else}
    <span class="context-item-label">{item.context.label}</span>
  {/if}
</button>

<style lang="postcss">
  .context-item-label {
    @apply basis-full grow shrink;
    @apply text-sm overflow-hidden whitespace-nowrap text-ellipsis;
  }

  .context-item-text {
    @apply flex flex-col basis-full grow shrink min-w-0;
  }

  .context-item-description {
    @apply text-xs text-fg-muted line-clamp-2;
  }

  .context-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1 w-full;
    @apply cursor-default select-none rounded-sm outline-none;
    @apply text-sm text-left text-wrap break-words;
  }
  /* Align the check and the icon with the label, the first line, instead of centering them on both lines. */
  .context-item.with-description {
    @apply items-start;
  }
  .context-item.with-description > :not(.context-item-text) {
    @apply flex items-center h-5;
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
