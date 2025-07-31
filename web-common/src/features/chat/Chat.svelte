<script lang="ts">
  import { createQuery } from "@tanstack/svelte-query";
  import { derived } from "svelte/store";
  import Resizer from "../../layout/Resizer.svelte";
  import {
    getRuntimeServiceGetConversationQueryOptions,
    getRuntimeServiceListConversationsQueryOptions,
    type V1Conversation,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { featureFlags } from "../feature-flags";
  import {
    chatActions,
    chatOpen,
    currentConversationId,
    DEFAULTS,
    isOptimisticId,
    loading,
    sidebarWidth,
  } from "./chat-store";
  import ChatHeader from "./header/ChatHeader.svelte";
  import ChatFooter from "./input/ChatFooter.svelte";
  import ChatInput from "./input/ChatInput.svelte";
  import ChatMessages from "./messages/ChatMessages.svelte";

  const { chat: chatFlag } = featureFlags;

  // Local UI state
  let input = "";
  let chatInputComponent: ChatInput;

  // Focus input when chat opens
  $: if ($chatOpen && chatInputComponent) {
    // Small delay to ensure DOM is ready
    setTimeout(() => {
      if (
        chatInputComponent &&
        typeof chatInputComponent.focusInput === "function"
      ) {
        chatInputComponent.focusInput();
      }
    }, 100);
  }

  // API clients
  const listConversationsQueryOptionsStore = derived(runtime, ($runtime) =>
    getRuntimeServiceListConversationsQueryOptions($runtime.instanceId, {
      query: {
        enabled: !!$chatOpen,
      },
    }),
  );
  const listConversationsQuery = createQuery(
    listConversationsQueryOptionsStore,
  );
  $: ({ data: listConversationsData } = $listConversationsQuery);

  const getConversationQueryOptionsStore = derived(
    [runtime, currentConversationId],
    ([$runtime, $currentConversationId]) =>
      getRuntimeServiceGetConversationQueryOptions(
        $runtime.instanceId,
        $currentConversationId || "",
        undefined,
        {
          query: {
            enabled:
              $chatOpen &&
              !!$currentConversationId &&
              !isOptimisticId($currentConversationId),
          },
        },
      ),
  );
  const getConversationQuery = createQuery(getConversationQueryOptionsStore);
  $: ({ data: getConversationData } = $getConversationQuery);
  $: currentConversation = getConversationData?.conversation || null;

  async function handleSendMessage(message: string) {
    try {
      await chatActions.sendMessage(message);
      // Refocus the input after sending
      if (
        chatInputComponent &&
        typeof chatInputComponent.focusInput === "function"
      ) {
        chatInputComponent.focusInput();
      }
    } catch (error) {
      // If sending failed, restore the message to the input so user can retry
      input = message;
      console.error("Failed to send message:", error);
    }
  }

  function createNewConversation() {
    chatActions.createNewConversation();
    if (
      chatInputComponent &&
      typeof chatInputComponent.focusInput === "function"
    ) {
      chatInputComponent.focusInput();
    }
  }

  function selectConversation(conv: V1Conversation) {
    chatActions.selectConversation(conv);
  }

  function onClose() {
    chatActions.closeChat();
  }
</script>

{#if $chatOpen && $chatFlag}
  <div class="chat-sidebar" style="--sidebar-width: {$sidebarWidth}px;">
    <Resizer
      min={DEFAULTS.MIN_SIDEBAR_WIDTH}
      max={DEFAULTS.MAX_SIDEBAR_WIDTH}
      basis={DEFAULTS.SIDEBAR_WIDTH}
      dimension={$sidebarWidth}
      direction="EW"
      side="left"
      onUpdate={chatActions.updateSidebarWidth}
    />
    <div class="chat-sidebar-content">
      <div class="chatbot-header-container">
        <ChatHeader
          currentTitle={currentConversation?.title || ""}
          conversations={listConversationsData?.conversations || []}
          currentConversationId={currentConversation?.id}
          onNewConversation={createNewConversation}
          onSelectConversation={selectConversation}
          {onClose}
        />
      </div>
      <ChatMessages isConversationLoading={$getConversationQuery.isLoading} />
      <ChatInput
        bind:this={chatInputComponent}
        bind:value={input}
        disabled={$loading}
        onSend={handleSendMessage}
      />
      <ChatFooter />
    </div>
  </div>
{/if}

<style lang="postcss">
  .chat-sidebar {
    position: relative;
    width: var(--sidebar-width);
    height: 100%;
    background: #ffffff;
    border-left: 1px solid #e5e7eb;
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
  }

  .chat-sidebar-content {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    flex: 1;
  }

  .chatbot-header-container {
    position: relative;
    flex-shrink: 0;
  }
</style>
