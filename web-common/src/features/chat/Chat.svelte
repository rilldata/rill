<script lang="ts">
  import Resizer from "../../layout/Resizer.svelte";
  import {
    createRuntimeServiceGetConversation,
    createRuntimeServiceListConversations,
    type V1Conversation,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import {
    chatActions,
    chatOpen,
    currentConversation,
    loading,
    sidebarWidth,
  } from "./chat-store";
  import ChatHeader from "./header/ChatHeader.svelte";
  import ChatFooter from "./input/ChatFooter.svelte";
  import ChatInput from "./input/ChatInput.svelte";
  import ChatMessages from "./messages/ChatMessages.svelte";
  import { DEFAULTS } from "./utils/storage";

  // Close chat using the centralized action
  function onClose() {
    chatActions.closeChat();
  }

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

  // API clients - these need to be reactive to runtime and currentConversation
  $: listConversations = createRuntimeServiceListConversations(
    $runtime.instanceId,
  );
  $: getConversation = createRuntimeServiceGetConversation(
    $runtime.instanceId,
    $currentConversation?.id || "",
    {
      query: {
        enabled: !!$currentConversation?.id,
      },
    },
  );

  // Load messages when conversation data is fetched
  $: if ($getConversation.data?.conversation?.messages) {
    chatActions.loadMessages($getConversation.data.conversation.messages);
  }

  // Update current conversation when fetched data changes (for new conversations)
  $: if ($getConversation.data?.conversation && !$currentConversation) {
    chatActions.updateCurrentConversation($getConversation.data.conversation);
  }

  // Send message
  async function handleSendMessage(message: string) {
    try {
      await chatActions.sendMessage(message);
    } catch (error) {
      // If sending failed, restore the message to the input so user can retry
      input = message;
      console.error("Failed to send message:", error);
    }
  }

  // Create a new conversation
  function createNewConversation() {
    chatActions.createNewConversation();
  }

  // Select a conversation from the list
  function selectConversation(conv: V1Conversation) {
    chatActions.selectConversation(conv);
    // Refetch conversation data to get latest messages
    $getConversation.refetch();
  }
</script>

{#if $chatOpen}
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
          currentTitle={$currentConversation?.title || ""}
          conversations={$listConversations.data?.conversations || []}
          currentConversationId={$currentConversation?.id}
          onNewConversation={createNewConversation}
          onSelectConversation={selectConversation}
          {onClose}
        />
      </div>
      <ChatMessages />
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
