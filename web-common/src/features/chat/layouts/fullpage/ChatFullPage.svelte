<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { useChatCore } from "../../core/chat";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import ChatMessages from "../../core/messages/ChatMessages.svelte";
  import ConversationSidebar from "./ConversationSidebar.svelte";

  // Props for routing and navigation
  export let routeType: "new" | "conversation" | undefined = undefined;
  export let conversationId: string | undefined = undefined;
  export let organization: string;
  export let project: string;

  // Use core chat logic
  const {
    currentConversationId,
    createNewConversation,
    selectConversation,
    listConversationsData,
    currentConversation,
    isConversationLoading,
    loading,
    handleSendMessage,
  } = useChatCore();

  // Local UI state
  let input = "";
  let chatInputComponent: ChatInput;
  let hasInitiallyFocused = false;

  // Focus on mount for page refreshes (when we're already on the right conversation)
  onMount(() => {
    // For page refreshes where we're already on the right conversation
    if (
      routeType === "conversation" &&
      conversationId === $currentConversationId
    ) {
      setTimeout(() => {
        focusInput();
        hasInitiallyFocused = true;
      }, 50); // Slightly longer delay for page refresh
    }
  });

  // Handle transition to new conversation route
  $: if (routeType === "new") {
    createNewConversation();
    // Focus after creating new conversation (same as conversation selection)
    setTimeout(() => focusInput(), 0);
  }

  // Handle conversation selection when routeType, conversationId, or conversation data changes
  $: if (
    routeType === "conversation" &&
    conversationId &&
    $listConversationsData?.conversations &&
    $currentConversationId !== conversationId
  ) {
    const conversations = $listConversationsData.conversations || [];
    const foundConversation = conversations.find(
      (conv) => conv.id === conversationId,
    );
    if (foundConversation) {
      selectConversation(foundConversation);
      // Focus immediately after prop change, not after navigation delay
      setTimeout(() => focusInput(), 0);
    }
  }

  // Focus input management with robust retry logic
  function focusInput() {
    if (
      chatInputComponent &&
      typeof chatInputComponent.focusInput === "function"
    ) {
      // Use requestAnimationFrame to ensure DOM is ready
      requestAnimationFrame(() => {
        chatInputComponent.focusInput();

        // Verify focus stuck after a brief delay
        setTimeout(() => {
          if (document.activeElement?.tagName !== "TEXTAREA") {
            chatInputComponent.focusInput();
          }
        }, 50);
      });
    }
  }

  // Handle conversation clicks - focus after navigation
  function handleConversationClick() {
    // Focus immediately after navigation starts, not after delay
    setTimeout(() => focusInput(), 0);
  }

  // Handle new conversation click - focus after navigation
  function handleNewConversationClick() {
    // Focus immediately after navigation starts, not after delay
    setTimeout(() => focusInput(), 0);
  }

  // Message handling with input focus + navigation
  async function onSendMessage(message: string) {
    await handleSendMessage(
      message,
      (conversationId) => {
        // If this was a new conversation, navigate to the conversation route
        if (routeType === "new" && conversationId) {
          goto(`/${organization}/${project}/-/chat/${conversationId}`, {
            replaceState: true,
          });
          // Focus immediately after navigation starts, not after arbitrary delay
          setTimeout(() => {
            focusInput();
          }, 0);
        } else {
          // For existing conversations, focus immediately since no navigation
          focusInput();
        }
      },
      (failedMessage) => {
        input = failedMessage;
      }, // onError
    );
  }
</script>

<div class="chat-fullpage">
  <!-- Conversation List Sidebar -->
  <ConversationSidebar
    conversations={$listConversationsData?.conversations || []}
    currentConversation={$currentConversation}
    onConversationClick={handleConversationClick}
    onNewConversationClick={handleNewConversationClick}
  />

  <!-- Main Chat Area -->
  <div class="chat-main">
    <div class="chat-content">
      <div class="chat-messages-wrapper">
        <ChatMessages
          layout="fullpage"
          isConversationLoading={$isConversationLoading}
        />
      </div>
    </div>

    <div class="chat-input-section">
      <div class="chat-input-wrapper">
        <ChatInput
          bind:this={chatInputComponent}
          bind:value={input}
          disabled={$loading}
          onSend={onSendMessage}
        />
        <ChatFooter />
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .chat-fullpage {
    display: flex;
    height: 100%;
    width: 100%;
    background: #ffffff;
  }

  /* Main Chat Area */
  .chat-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: #ffffff;
  }

  .chat-content {
    flex: 1;
    overflow: hidden;
    background: #f9fafb;
    display: flex;
    flex-direction: column;
  }

  .chat-messages-wrapper {
    flex: 1;
    overflow-y: auto;
    width: 100%;
    display: flex;
    flex-direction: column;
  }

  .chat-input-section {
    flex-shrink: 0;
    background: #f9fafb;
    padding: 1rem;
    display: flex;
    justify-content: center;
  }

  .chat-input-wrapper {
    width: 100%;
    max-width: 48rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  /* Override core ChatMessages background for full-page layout */
  .chat-fullpage :global(.chat-messages) {
    background: #f9fafb;
    padding: 2rem 1rem;
    min-height: 100%;
  }

  /* Enhance welcome message for full-page layout */
  .chat-fullpage :global(.chat-empty) {
    padding: 4rem 2rem;
  }

  .chat-fullpage :global(.chat-empty-title) {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin-bottom: 0.5rem;
  }

  .chat-fullpage :global(.chat-empty-subtitle) {
    font-size: 1rem;
    color: #6b7280;
  }

  /* Responsive behavior for full-page layout */
  @media (max-width: 768px) {
    .chat-messages-wrapper,
    .chat-input-wrapper {
      max-width: none;
      padding: 0 1rem;
    }

    .chat-input-section {
      padding: 1rem;
    }
  }

  @media (max-width: 640px) {
    .chat-fullpage {
      flex-direction: column;
    }

    .chat-fullpage :global(.chat-empty) {
      padding: 2rem 1rem;
    }

    .chat-fullpage :global(.chat-empty-title) {
      font-size: 1.25rem;
    }

    .chat-fullpage :global(.chat-empty-subtitle) {
      font-size: 0.875rem;
    }
  }
</style>
