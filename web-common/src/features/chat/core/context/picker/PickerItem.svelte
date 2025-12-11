<script lang="ts">
  import {
    type InlineContext,
    inlineContextsAreEqual,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";
  import { CheckIcon } from "lucide-svelte";

  export let context: InlineContext;
  export let selectedChatContext: InlineContext | null = null;
  export let highlightManager: PickerOptionsHighlightManager;
  export let onSelect: (ctx: InlineContext) => void;

  const highlightedContextStore = highlightManager.highlightedContext;
  $: highlightedContext = $highlightedContextStore;
  const mouseContextHighlightHandler =
    highlightManager.mouseOverHandler(context);

  $: selected =
    selectedChatContext !== null &&
    inlineContextsAreEqual(context, selectedChatContext);
  $: highlighted =
    highlightedContext !== null &&
    inlineContextsAreEqual(context, highlightedContext);

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

<button
  class="context-item"
  class:highlight={highlighted}
  type="button"
  on:click={() => onSelect(context)}
  use:ensureInView={highlighted}
  use:mouseContextHighlightHandler
>
  <div class="context-item-checkbox">
    {#if selected}
      <CheckIcon size="12px" />
    {/if}
  </div>
  <div class="square"></div>
  <span class="context-item-label">{context.label}</span>
  <div class="context-item-keyboard-shortcut" class:hidden={!highlighted}>
    ↑/↓
  </div>
</button>

<style lang="postcss">
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
</style>
