<script lang="ts">
  import { afterUpdate } from "svelte";
  import type { V1Message } from "../../../../runtime-client";
  import DelayedSpinner from "../../../entity-management/DelayedSpinner.svelte";
  import type { ConversationManager } from "../conversation-manager";
  import { MessageContentType, MessageType, ToolName } from "../types";
  import { parseChartData } from "../utils";
  import ChartBlock from "./ChartBlock.svelte";
  import Error from "./Error.svelte";
  import Message from "./Message.svelte";
  import ThinkingBlock from "./ThinkingBlock.svelte";
  import {
    calculateThinkingDuration,
    isThinkingBlockComplete,
  } from "./thinking-block-utils";

  export let conversationManager: ConversationManager;
  export let layout: "sidebar" | "fullpage";

  let messagesContainer: HTMLDivElement;

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();

  // Loading states - access the store from the conversation instance
  $: isStreamingStore = currentConversation.isStreaming;
  $: isStreaming = $isStreamingStore;
  $: isConversationLoading = !!$getConversationQuery.isLoading;

  // Error handling
  $: streamErrorStore = currentConversation.streamError;
  $: conversationQueryError = currentConversation.getConversationQueryError();
  $: hasConversationLoadError = !!$conversationQueryError;
  $: hasStreamError = !!$streamErrorStore;

  // Data
  $: messages = $getConversationQuery.data?.conversation?.messages ?? [];

  // Build a map of result messages by parent ID for correlation with calls (excluding router_agent)
  $: resultMessagesByParentId = new Map(
    messages
      .filter(
        (msg) =>
          msg.type === MessageType.RESULT && msg.tool !== ToolName.ROUTER_AGENT,
      )
      .map((msg) => [msg.parentId, msg]),
  );

  // Filter out tool result messages (but keep router_agent results which are assistant responses)
  $: displayMessages = messages.filter(
    (msg) =>
      msg.type !== MessageType.RESULT || msg.tool === ToolName.ROUTER_AGENT,
  );

  // Group messages: progress messages become headers for thinking blocks with nested tool calls
  $: messageGroups = groupMessages(displayMessages);

  function groupMessages(msgs: V1Message[]) {
    const groups: Array<{
      type: "text" | "thinking";
      message?: V1Message;
      messages?: V1Message[];
    }> = [];

    let currentThinkingBlock: V1Message[] | null = null;

    for (const msg of msgs) {
      if (msg.tool === ToolName.ROUTER_AGENT) {
        // Text message (user/assistant) - close any open thinking block and add as standalone
        if (currentThinkingBlock) {
          groups.push({
            type: "thinking",
            messages: currentThinkingBlock,
          });
          currentThinkingBlock = null;
        }
        groups.push({ type: "text", message: msg });
      } else if (msg.type === MessageType.PROGRESS) {
        // Add to current thinking block or start a new one
        if (currentThinkingBlock) {
          currentThinkingBlock.push(msg);
        } else {
          currentThinkingBlock = [msg];
        }
      } else if (msg.type === MessageType.CALL) {
        // Tool calls are always part of a thinking block
        if (currentThinkingBlock) {
          currentThinkingBlock.push(msg);
        } else {
          // Start a new thinking block (without a progress message yet)
          currentThinkingBlock = [msg];
        }
      }
    }

    // Close any remaining thinking block
    if (currentThinkingBlock) {
      groups.push({
        type: "thinking",
        messages: currentThinkingBlock,
      });
    }

    return groups;
  }

  // Auto-scroll to bottom when messages change or loading state changes
  afterUpdate(() => {
    if (messagesContainer && layout === "sidebar") {
      // For sidebar layout, scroll the messages container
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    } else if (layout === "fullpage") {
      // For fullpage layout, scroll the parent wrapper
      const parentWrapper = messagesContainer.closest(".chat-messages-wrapper");
      if (parentWrapper) {
        parentWrapper.scrollTop = parentWrapper.scrollHeight;
      }
    }
  });
</script>

<div
  class="chat-messages"
  class:sidebar={layout === "sidebar"}
  class:fullpage={layout === "fullpage"}
  bind:this={messagesContainer}
>
  {#if isConversationLoading}
    <div class="chat-loading">
      <DelayedSpinner isLoading={isConversationLoading} size="24px" />
    </div>
  {:else if hasConversationLoadError}
    <Error
      headline="Unable to load conversation"
      error={$conversationQueryError}
    />
  {:else if messages.length === 0}
    <div class="chat-empty">
      <!-- <div class="chat-empty-icon">💬</div> -->
      <div class="chat-empty-title">How can I help you today?</div>
      <div class="chat-empty-subtitle">Happy to help explore your data</div>
    </div>
  {:else}
    {#each messageGroups as group, i (group.type === "text" ? group.message?.id : `thinking-${i}`)}
      {#if group.type === "text" && group.message}
        <Message message={group.message} />
      {:else if group.type === "thinking" && group.messages}
        {@const isComplete = isThinkingBlockComplete(group.messages, messages)}
        {@const duration = calculateThinkingDuration(group.messages)}
        <ThinkingBlock
          messages={group.messages}
          {resultMessagesByParentId}
          {isComplete}
          {duration}
        />

        <!-- Render charts from this thinking block at top level -->
        {#each group.messages as msg (msg.id)}
          {#if msg.tool === ToolName.CREATE_CHART}
            {@const resultMsg = resultMessagesByParentId.get(msg.id)}
            {@const hasResult = !!resultMsg}
            {@const isError =
              resultMsg?.contentType === MessageContentType.ERROR}
            {@const chartData = parseChartData({ input: msg.contentData })}
            {#if chartData && hasResult && !isError}
              <div class="chart-display">
                <ChartBlock
                  chartType={chartData.chartType}
                  chartSpec={chartData.chartSpec}
                />
              </div>
            {/if}
          {/if}
        {/each}
      {/if}
    {/each}
  {/if}
  {#if hasStreamError}
    <Error headline="Failed to generate response" error={$streamErrorStore} />
  {/if}
</div>

<style lang="postcss">
  .chat-messages {
    @apply flex-1;
    @apply flex flex-col gap-2;
    background: var(--surface);
  }

  .chat-messages.sidebar {
    @apply overflow-y-auto;
    @apply px-4;
  }

  .chat-messages.fullpage {
    @apply p-4;
    @apply max-w-3xl mx-auto w-full;
    @apply min-h-full;
  }

  .chat-empty {
    @apply flex flex-col;
    @apply items-center justify-center;
    @apply h-full text-center;
    @apply text-gray-500;
  }

  .chat-messages.fullpage .chat-empty {
    @apply py-16 px-8;
  }

  .chat-empty-title {
    @apply text-base font-semibold;
    @apply text-gray-700 mb-1;
  }

  .chat-messages.fullpage .chat-empty-title {
    @apply text-2xl font-semibold;
    @apply text-gray-900 mb-2;
  }

  .chat-empty-subtitle {
    @apply text-xs text-gray-500;
  }

  .chat-messages.fullpage .chat-empty-subtitle {
    @apply text-base text-gray-500;
  }

  @media (max-width: 640px) {
    .chat-messages.fullpage .chat-empty {
      padding-top: 2rem;
      padding-bottom: 2rem;
      padding-left: 1rem;
      padding-right: 1rem;
    }

    .chat-messages.fullpage .chat-empty-title {
      font-size: 1.25rem;
      line-height: 1.75rem;
    }

    .chat-messages.fullpage .chat-empty-subtitle {
      font-size: 0.875rem;
      line-height: 1.25rem;
    }
  }

  .chart-display {
    @apply w-full max-w-full;
    @apply mt-2 self-start;
  }
</style>
