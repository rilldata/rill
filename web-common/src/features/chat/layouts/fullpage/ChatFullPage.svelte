<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { isOptimisticConversationId, useChatCore } from "../../core/chat";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import ChatMessages from "../../core/messages/ChatMessages.svelte";
  import ConversationSidebar from "./ConversationSidebar.svelte";
  import { createFullpageConversationIdStore } from "./fullpage-store";

  // Extract route parameters
  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: conversationId = $page.params.conversationId || null;

  // Create project-specific conversation store (reactive)
  $: fullpageConversationId = createFullpageConversationIdStore(
    organization,
    project,
  );

  // Use core chat logic with stable instance
  const {
    currentConversationId,
    listConversationsData,
    currentConversation,
    isConversationLoading,
    loading,
    error,
    messages,
    handleSendMessage,
  } = useChatCore({
    onConversationChange: (id) => {
      fullpageConversationId?.set(id);
    },
  });

  // Local UI state
  let input = "";
  let chatInputComponent: ChatInput;

  // Focus on mount with a small delay for component initialization
  onMount(() => {
    // Give the component tree time to fully initialize
    setTimeout(() => {
      chatInputComponent?.focusInput();
    }, 100);
  });

  // Synchronize conversation state with URL (URL is source of truth)
  // But don't interfere with optimistic conversations
  $: {
    if (
      !conversationId &&
      $currentConversationId !== null &&
      !isOptimisticConversationId($currentConversationId)
    ) {
      currentConversationId.set(null);
      chatInputComponent?.focusInput();
    } else if (conversationId && $currentConversationId !== conversationId) {
      currentConversationId.set(conversationId);
      chatInputComponent?.focusInput();
    }
  }

  // Handle conversation clicks - just focus, URL change handles the rest
  function handleConversationClick() {
    // URL navigation will trigger state sync via reactive statement
    chatInputComponent?.focusInput();
  }

  // Handle new conversation click - just focus, URL change handles the rest
  function handleNewConversationClick() {
    // URL navigation will trigger state sync via reactive statement
    chatInputComponent?.focusInput();
  }

  // Message handling with input focus + navigation
  async function onSendMessage(message: string) {
    await handleSendMessage(
      message,
      (newConversationId) => {
        // If this was a new conversation, navigate to the conversation route
        if (!conversationId && newConversationId) {
          goto(`/${organization}/${project}/-/chat/${newConversationId}`, {
            replaceState: true,
          });
        }
        chatInputComponent?.focusInput();
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
          error={$error}
          loading={$loading}
          messages={$messages}
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
