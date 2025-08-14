<script lang="ts">
  import { onMount } from "svelte";
  import AlertCircle from "../../../../components/icons/AlertCircle.svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { Chat } from "../../core/chat";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import ChatMessages from "../../core/messages/ChatMessages.svelte";
  import ConversationSidebar from "./ConversationSidebar.svelte";

  $: ({ instanceId } = $runtime);

  $: chat = new Chat(instanceId, {
    conversationState: "url",
  });

  $: currentConversationStore = chat.getCurrentConversation();
  $: getConversationQuery = $currentConversationStore.getConversationQuery();
  // Alternative:
  // $: currentConversationQuery = chat.getCurrentConversationQuery();

  let chatInputComponent: ChatInput;

  // Focus on mount with a small delay for component initialization
  onMount(() => {
    // Give the component tree time to fully initialize
    setTimeout(() => {
      chatInputComponent?.focusInput();
    }, 100);
  });

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }
</script>

<div class="chat-fullpage">
  <!-- Conversation List Sidebar -->
  <ConversationSidebar
    {chat}
    onConversationClick={() => {
      chatInputComponent?.focusInput();
    }}
    onNewConversationClick={() => {
      chatInputComponent?.focusInput();
    }}
  />

  <!-- Main Chat Area -->
  <div class="chat-main">
    <div class="chat-content">
      <div class="chat-messages-wrapper">
        <ChatMessages {chat} layout="fullpage" />
      </div>
    </div>

    <div class="chat-input-section">
      <div class="chat-input-wrapper">
        {#if $getConversationQuery?.error}
          <div class="chat-input-error">
            <AlertCircle size="1.2em" />
            {$getConversationQuery?.error.message}
          </div>
        {/if}
        <ChatInput
          {chat}
          onSend={onMessageSend}
          bind:this={chatInputComponent}
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

  .chat-input-error {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: #991b1b;
    background: #fef7f7;
    border-left: 3px solid #f87171;
    border-radius: 0.375rem;
    margin: 0 1rem 0.5rem 1rem;
    box-sizing: border-box;
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
