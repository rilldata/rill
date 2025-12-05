<script lang="ts">
  import { createQuery } from "@tanstack/svelte-query";
  import { afterUpdate } from "svelte";
  import { derived } from "svelte/store";
  import DelayedSpinner from "../../../entity-management/DelayedSpinner.svelte";
  import { getRuntimeServiceListToolsQueryOptions } from "../../../../runtime-client";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import type { ConversationManager } from "../conversation-manager";
  import { type ChatConfig } from "../types";
  import ChartBlock from "./chart/ChartBlock.svelte";
  import Error from "./Error.svelte";
  import AssistantMessage from "./text/AssistantMessage.svelte";
  import UserMessage from "./text/UserMessage.svelte";
  import PlanningBlock from "./thinking/PlanningBlock.svelte";
  import ThinkingBlock from "./thinking/ThinkingBlock.svelte";

  export let conversationManager: ConversationManager;
  export let layout: "sidebar" | "fullpage";
  export let config: ChatConfig;

  // Prefetch tools metadata for CallMessage display names
  const listToolsQueryOptionsStore = derived(runtime, ($runtime) =>
    getRuntimeServiceListToolsQueryOptions($runtime.instanceId),
  );
  const listToolsQuery = createQuery(listToolsQueryOptionsStore);
  $: tools = $listToolsQuery.data?.tools;

  let messagesContainer: HTMLDivElement;

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();

  // Error handling
  $: conversationQueryError = currentConversation.getConversationQueryError();
  $: hasConversationLoadError = !!$conversationQueryError;
  $: streamErrorStore = currentConversation.streamError;
  $: hasStreamError = !!$streamErrorStore;

  // Message blocks for display
  $: messageBlocksStore = currentConversation.getMessageBlocks();
  $: messageBlocks = $messageBlocksStore;

  // Data for ThinkingBlock (still needs result map for CallMessage rendering)
  $: messages = $getConversationQuery.data?.messages ?? [];
  $: resultMessagesByParentId = new Map(
    messages
      .filter((msg) => msg.type === "result" && msg.tool !== "router_agent")
      .map((msg) => [msg.parentId, msg]),
  );

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
  {#if $getConversationQuery.isLoading}
    <div class="chat-loading">
      <DelayedSpinner isLoading={$getConversationQuery.isLoading} size="24px" />
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
      <div class="chat-empty-subtitle">
        {config.emptyChatLabel}
      </div>
    </div>
  {:else}
    {#each messageBlocks as block (block.id)}
      {#if block.type === "text" && block.message.role === "user"}
        <UserMessage message={block.message} />
      {:else if block.type === "text" && block.message.role === "assistant"}
        <AssistantMessage message={block.message} />
      {:else if block.type === "thinking"}
        <ThinkingBlock
          messages={block.messages}
          {resultMessagesByParentId}
          isComplete={block.isComplete}
          duration={block.duration}
          {tools}
        />
      {:else if block.type === "planning"}
        <PlanningBlock />
      {:else if block.type === "chart"}
        <ChartBlock
          chartType={block.chartData.chartType}
          chartSpec={block.chartData.chartSpec}
        />
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
</style>
