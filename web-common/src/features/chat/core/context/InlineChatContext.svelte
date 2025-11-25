<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2/index.ts";
  import { builderActions, getAttrs } from "bits-ui";
  import { ChevronDownIcon } from "lucide-svelte";
  import { getInlineChatContextMetadata } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";
  import {
    ChatContextEntryType,
    InlineContextConfig,
    type InlineChatContext,
  } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
  import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
  import AddInlineChatDropdown from "@rilldata/web-common/features/chat/core/context/AddInlineChatDropdown.svelte";

  export let conversationManager: ConversationManager;
  export let inlineChatContext: InlineChatContext | null = null;
  export let onSelect: (ctx: InlineChatContext) => void;

  let left = 0;
  let bottom = 0;
  let chatElement: HTMLSpanElement;
  let open = false;
  let tooltipOpen = false;

  const contextMetadataStore = getInlineChatContextMetadata();

  $: typeData = inlineChatContext?.type
    ? InlineContextConfig[inlineChatContext.type]
    : undefined;
  $: label =
    typeData?.getLabel(inlineChatContext!, $contextMetadataStore) ?? "";

  $: isMetricsViewContext =
    inlineChatContext?.type === ChatContextEntryType.Measure ||
    inlineChatContext?.type === ChatContextEntryType.Dimension;
  $: metricsViewName = isMetricsViewContext
    ? InlineContextConfig[ChatContextEntryType.MetricsView]!.getLabel(
        inlineChatContext!,
        $contextMetadataStore,
      )
    : "";

  $: supportsEditing =
    !inlineChatContext ||
    inlineChatContext.type === ChatContextEntryType.MetricsView ||
    inlineChatContext.type === ChatContextEntryType.Measure ||
    inlineChatContext.type === ChatContextEntryType.Dimension;

  function toggleDropdown() {
    const rect = chatElement.getBoundingClientRect();
    left = rect.left;
    bottom = window.innerHeight - rect.bottom + 16;

    open = !open;
    tooltipOpen = false;
  }
</script>

<span
  bind:this={chatElement}
  class="inline-chat-context"
  contenteditable="false"
>
  {#if inlineChatContext}
    <div class="inline-chat-context-value">
      {#if metricsViewName}
        <Tooltip.Root bind:open={tooltipOpen}>
          <Tooltip.Trigger asChild let:builder>
            <span
              {...getAttrs([builder])}
              use:builderActions={{ builders: [builder] }}
            >
              {label}
            </span>
          </Tooltip.Trigger>
          <Tooltip.Content>
            From {metricsViewName}
          </Tooltip.Content>
        </Tooltip.Root>
      {:else}
        <span>{label}</span>
      {/if}
      {#if supportsEditing}
        <button on:click={toggleDropdown} type="button">
          <ChevronDownIcon size="12px" />
        </button>
      {/if}
    </div>
  {/if}

  {#if supportsEditing && open}
    <AddInlineChatDropdown
      {conversationManager}
      {left}
      {bottom}
      {inlineChatContext}
      {onSelect}
    />
  {/if}
</span>

<style lang="postcss">
  .inline-chat-context {
    @apply inline-block gap-1 text-sm underline;
  }

  .inline-chat-context-value {
    @apply flex flex-row items-center gap-x-0.5 cursor-pointer;
  }
</style>
