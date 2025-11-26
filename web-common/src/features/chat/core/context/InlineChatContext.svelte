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
  import InlineChatContextPicker from "@rilldata/web-common/features/chat/core/context/InlineChatContextPicker.svelte";

  export let conversationManager: ConversationManager;
  export let selectedChatContext: InlineChatContext;
  export let onSelect: (ctx: InlineChatContext) => void;
  export let onDropdownToggle: (open: boolean) => void;
  export let focusEditor: () => void;

  let left = 0;
  let bottom = 0;
  let chatElement: HTMLSpanElement;
  let open = false;
  let tooltipOpen = false;

  const contextMetadataStore = getInlineChatContextMetadata();

  $: typeData = selectedChatContext.type
    ? InlineContextConfig[selectedChatContext.type]
    : undefined;
  $: label =
    typeData?.getLabel(selectedChatContext!, $contextMetadataStore) ?? "";

  $: isMetricsViewContext =
    selectedChatContext.type === ChatContextEntryType.Measure ||
    selectedChatContext.type === ChatContextEntryType.Dimension;
  $: metricsViewName = isMetricsViewContext
    ? InlineContextConfig[ChatContextEntryType.MetricsView]!.getLabel(
        selectedChatContext,
        $contextMetadataStore,
      )
    : "";

  $: supportsEditing =
    selectedChatContext.type === ChatContextEntryType.MetricsView ||
    selectedChatContext.type === ChatContextEntryType.Measure ||
    selectedChatContext.type === ChatContextEntryType.Dimension;

  function toggleDropdown() {
    const rect = chatElement.getBoundingClientRect();
    left = rect.left;
    bottom = window.innerHeight - rect.bottom + 16;

    open = !open;
    onDropdownToggle(open);
    tooltipOpen = false;
  }

  /**
   * Called from editor plugins. Used to make sure opening another component's dropdowns closes this.
   */
  export function closeDropdown() {
    open = false;
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      open = false;
      onDropdownToggle(false);
      focusEditor();
    }
  }
</script>

<svelte:window on:keydown={handleKeyDown} />

<span
  bind:this={chatElement}
  class="inline-chat-context"
  contenteditable="false"
>
  <svelte:element
    this={supportsEditing ? "button" : "div"}
    class="inline-chat-context-value"
    class:cursor-default={!supportsEditing}
    on:click={toggleDropdown}
    type="button"
    role="button"
    tabindex="-1"
  >
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
      <ChevronDownIcon size="12px" />
    {/if}
  </svelte:element>

  {#if supportsEditing && open}
    <InlineChatContextPicker
      {conversationManager}
      {left}
      {bottom}
      {selectedChatContext}
      {onSelect}
      {focusEditor}
    />
  {/if}
</span>

<style lang="postcss">
  .inline-chat-context {
    @apply inline-block gap-1 text-sm underline;
  }

  .inline-chat-context-value {
    @apply flex flex-row items-center gap-x-0.5;
  }
</style>
