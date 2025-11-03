<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { onMount } from "svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import Messages from "../../core/messages/Messages.svelte";
  import ConversationSidebar from "./ConversationSidebar.svelte";

  $: ({ instanceId } = $runtime);

  $: conversationManager = getConversationManager(instanceId, {
    conversationState: "url",
  });

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

  // Clean up conversation manager resources when leaving the chat context entirely
  beforeNavigate(({ to }) => {
    const isChatRoute = to?.route?.id?.includes("ai");
    if (!isChatRoute) {
      cleanupConversationManager(instanceId);
    }
  });
</script>

<div class="chat-fullpage">
  <!-- Conversation List Sidebar -->
  <ConversationSidebar
    {conversationManager}
    onConversationClick={() => {
      chatInputComponent?.focusInput();
    }}
    onNewConversationClick={() => {
      chatInputComponent?.focusInput();
    }}
  >
    <svelte:fragment slot="footer">
      <slot name="sidebar-footer" />
    </svelte:fragment>
  </ConversationSidebar>

  <!-- Main Chat Area -->
  <div class="chat-main">
    <div class="chat-content">
      <div class="chat-messages-wrapper">
        <Messages {conversationManager} layout="fullpage" />
      </div>
    </div>

    <div class="chat-input-section">
      <div class="chat-input-wrapper">
        <ChatInput
          {conversationManager}
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
    background: var(--surface);
  }

  /* Main Chat Area */
  .chat-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--surface);
  }

  .chat-content {
    flex: 1;
    overflow: hidden;
    background: var(--surface);
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
    background: var(--surface);
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
  }
</style>
