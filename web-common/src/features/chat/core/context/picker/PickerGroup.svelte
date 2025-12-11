<script lang="ts">
  import * as Collapsible from "@rilldata/web-common/components/collapsible";
  import { getAttrs, builderActions } from "bits-ui";
  import {
    type InlineContext,
    InlineContextConfig,
    inlineContextIsWithin,
    inlineContextsAreEqual,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import type { Readable } from "svelte/store";
  import { CheckIcon, ChevronDownIcon, ChevronRightIcon } from "lucide-svelte";
  import type { InlineContextPickerOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
  import { PickerOptionsHighlightManager } from "@rilldata/web-common/features/chat/core/context/picker/highlight-manager.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  export let parentOption: InlineContextPickerOption;
  export let selectedChatContext: InlineContext | null = null;
  export let highlightManager: PickerOptionsHighlightManager;
  export let searchTextStore: Readable<string>;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  $: ({
    context,
    recentlyUsed,
    currentlyActive,
    children,
    childrenQueryOptions,
  } = parentOption);
  let childrenQuery: ReturnType<typeof createQuery> | undefined;
  $: if (childrenQueryOptions)
    childrenQuery = createQuery(childrenQueryOptions, queryClient);
  $: resolvedChildren =
    (childrenQuery
      ? ($childrenQuery?.data as InlineContextPickerOption["children"])
      : children) ?? [];
  $: if (childrenQuery)
    highlightManager.childrenUpdated(context, $childrenQuery?.data ?? []);

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

  $: parentIcon = InlineContextConfig[context.type]?.getIcon?.(context);

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
      <input
        type="radio"
        checked={parentOptionSelected}
        on:click|stopPropagation={() => onSelect(context)}
        class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
      />
      <div class="min-w-3.5">
        {#if open}
          <ChevronDownIcon size="12px" strokeWidth={4} />
        {:else if parentIcon}
          <!-- On hover show chevron right icon -->
          <svelte:component this={parentIcon} size="12px" />
        {:else}
          <ChevronRightIcon size="12px" strokeWidth={4} />
        {/if}
      </div>
      <span class="context-item-label">{context.label}</span>
      <div
        class="context-item-keyboard-shortcut"
        class:hidden={!parentOptionHighlighted}
      >
        ↑/↓
      </div>
      {#if recentlyUsed}
        <span class="parent-context-label">Recently asked</span>
      {:else if currentlyActive}
        <span class="parent-context-label">Current</span>
      {/if}
    </button>
  </Collapsible.Trigger>
  <Collapsible.Content class="flex flex-col ml-0.5 gap-y-0.5">
    {#each resolvedChildren as childCategory, i}
      {#if i !== 0}<div class="content-separator"></div>{/if}

      {#each childCategory as child}
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
              <svelte:component this={icon} size="14px" />
            </div>
          {:else}
            <div class="context-item-icon"></div>
          {/if}
          <span class="context-item-label">{child.label}</span>
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

  .context-item-icon {
    @apply min-w-3.5 h-2;
  }

  .contents-empty {
    @apply px-2 py-1.5 w-full ui-copy-inactive;
  }

  .content-separator {
    @apply -mx-1 my-1 h-px bg-muted;
  }
</style>
