<script lang="ts">
  import * as Collapsible from "@rilldata/web-common/components/collapsible/index.ts";
  import {
    type InlineChatContext,
    inlineChatContextsAreEqual,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { ChevronDownIcon } from "lucide-svelte";
  import { onMount } from "svelte";
  import { writable } from "svelte/store";
  import {
    getInlineChatContextFilteredOptions, getInlineChatContextMetadata,
  } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
  import { ContextTypeData } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";

  export let inlineChatContext: InlineChatContext | null = null;
  export let onUpdate: () => void;
  export let focusEditor: () => void;

  let left = 0;
  let bottom = 0;
  let chatElement: HTMLSpanElement;
  let open = false;

  const searchTextStore = writable("");
  const filteredOptions = getInlineChatContextFilteredOptions(searchTextStore);
  const contextMetadataStore = getInlineChatContextMetadata();

  $: typeData = inlineChatContext?.type ? ContextTypeData[inlineChatContext.type] : undefined
  $: label = typeData?.getLabel(inlineChatContext!, $contextMetadataStore) ?? "";

  function select(context: InlineChatContext) {
    inlineChatContext = context;
    open = false;
    searchTextStore.set("");

    setTimeout(() => {
      onUpdate();
      focusEditor();
    });
  }

  export function setText(newText: string) {
    searchTextStore.set(newText);
  }

  // Used to build the full prompt from text and context components.
  export function getChatContext() {
    return inlineChatContext;
  }

  export function selectFirst() {
    const option = $filteredOptions?.[0]?.metricsViewContext;
    if (!option) return;
    select(option);
  }

  onMount(() => {
    const rect = chatElement.getBoundingClientRect();
    left = rect.left;
    bottom = window.innerHeight - rect.bottom + 16;

    if (inlineChatContext === null) open = true;
  });
</script>

<span
  bind:this={chatElement}
  class="inline-chat-context"
  contenteditable="false"
>
  {#if inlineChatContext}
    <div class="inline-chat-context-value">
      <span>{label}</span>
      <button on:click={() => (open = !open)} type="button">
        <ChevronDownIcon size="12px" />
      </button>
    </div>
  {/if}

  <!-- bits-ui dropdown component captures focus, so chat text cannot be edited when it is open.
       Newer versions of bits-ui have "trapFocus=false" param but it needs svelte5 upgrade.
       TODO: move to dropdown component after upgrade. -->
  <div
    class="inline-chat-context-dropdown"
    style="left: {left}px; bottom: {bottom}px; display: {open
      ? 'block'
      : 'none'};"
  >
    {#each $filteredOptions as { metricsViewContext, measures, dimensions } (metricsViewContext.values[0])}
      {@const metricsViewSelected =
        inlineChatContext !== null &&
        inlineChatContextsAreEqual(metricsViewContext, inlineChatContext)}
      <Collapsible.Root open={$searchTextStore.length > 2}>
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
              on:click|stopPropagation={() => select(metricsViewContext)}
              class="w-3 h-3 text-blue-600 border-gray-300 focus:ring-blue-500"
            />
            <span class="text-sm">{metricsViewContext.label}</span>
          </button>
        </Collapsible.Trigger>
        <Collapsible.Content class="flex flex-col ml-6 gap-y-0.5">
          {#each measures as measure (measure.values[1])}
            <button
              class="context-item"
              type="button"
              on:click={() => select(measure)}
            >
              <div class="square"></div>
              <span>{measure.label}</span>
            </button>
          {/each}

          {#if measures.length > 0}
            <div class="content-separator"></div>
          {/if}

          {#each dimensions as dimension (dimension.values[1])}
            <button
              class="context-item"
              type="button"
              on:click={() => select(dimension)}
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
</span>

<style lang="postcss">
  .inline-chat-context {
    @apply inline-block gap-1 text-sm underline;
  }

  .inline-chat-context-value {
    @apply flex flex-row items-center gap-x-0.5;
  }

  .inline-chat-context-dropdown {
    @apply flex flex-col fixed p-1.5 z-50 w-[500px] max-h-[500px] overflow-auto;
    @apply rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .metrics-view-context-item {
    @apply font-semibold;
  }

  .context-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1.5 w-full;
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
