<script lang="ts">
  import * as Collapsible from "@rilldata/web-common/components/collapsible/index.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import {
    type InlineContext,
    inlineContextsAreEqual,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import type { MetricsViewContextOption } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
  import type { Readable } from "svelte/store";
  import { CheckIcon, ChevronDownIcon, ChevronRightIcon } from "lucide-svelte";

  export let metricsViewContextOption: MetricsViewContextOption;
  export let selectedChatContext: InlineContext | null = null;
  export let highlightedContext: InlineContext | null = null;
  export let searchTextStore: Readable<string>;
  export let onSelect: (ctx: InlineContext) => void;
  export let focusEditor: () => void;

  $: ({
    metricsViewContext,
    recentlyUsed,
    currentlyActive,
    measures,
    dimensions,
  } = metricsViewContextOption);
  $: metricsViewSelected =
    selectedChatContext !== null &&
    inlineContextsAreEqual(metricsViewContext, selectedChatContext);
  $: withinMetricsViewSelected =
    selectedChatContext?.metricsView === metricsViewContext.metricsView;
  $: metricsViewHighlighted =
    highlightedContext !== null &&
    inlineContextsAreEqual(metricsViewContext, highlightedContext);
  $: withinMetricsViewHighlighted =
    highlightedContext?.metricsView === metricsViewContext.metricsView;

  let open =
    withinMetricsViewSelected ||
    metricsViewContextOption.recentlyUsed ||
    metricsViewContextOption.currentlyActive;
  $: shouldForceOpen =
    withinMetricsViewHighlighted || $searchTextStore.length > 0;
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
      class="context-item metrics-view-context-item"
      class:highlight={metricsViewHighlighted}
      type="button"
      {...getAttrs([builder])}
      use:builderActions={{ builders: [builder] }}
      use:ensureInView={metricsViewHighlighted}
      on:click={focusEditor}
    >
      {#if open}
        <ChevronDownIcon size="12px" strokeWidth={4} />
      {:else}
        <ChevronRightIcon size="12px" strokeWidth={4} />
      {/if}
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
  <Collapsible.Content class="flex flex-col ml-0.5 gap-y-0.5">
    {#each measures as measure (measure.measure)}
      {@const selected =
        selectedChatContext !== null &&
        inlineContextsAreEqual(measure, selectedChatContext)}
      {@const highlighted =
        highlightedContext !== null &&
        inlineContextsAreEqual(measure, highlightedContext)}
      <button
        class="context-item"
        class:highlight={highlighted}
        type="button"
        on:click={() => onSelect(measure)}
        use:ensureInView={highlighted}
      >
        <div class="context-item-checkbox">
          {#if selected}
            <CheckIcon size="12px" />
          {/if}
        </div>
        <div class="square"></div>
        <span>{measure.label}</span>
      </button>
    {/each}

    {#if measures.length > 0 && dimensions.length > 0}
      <div class="content-separator"></div>
    {/if}

    {#each dimensions as dimension (dimension.dimension)}
      {@const selected =
        selectedChatContext !== null &&
        inlineContextsAreEqual(dimension, selectedChatContext)}
      {@const highlighted =
        highlightedContext !== null &&
        inlineContextsAreEqual(dimension, highlightedContext)}
      <button
        class="context-item"
        class:highlight={highlighted}
        type="button"
        on:click={() => onSelect(dimension)}
        use:ensureInView={highlighted}
      >
        <div class="context-item-checkbox">
          {#if selected}
            <CheckIcon size="12px" />
          {/if}
        </div>
        <div class="circle"></div>
        <span>{dimension.label}</span>
      </button>
    {/each}

    {#if measures.length === 0 && dimensions.length === 0}
      <div class="contents-empty">No dimensions or measures found</div>
    {/if}
  </Collapsible.Content>
</Collapsible.Root>

<style lang="postcss">
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
