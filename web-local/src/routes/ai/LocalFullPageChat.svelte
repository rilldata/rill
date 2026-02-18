<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { onMount } from "svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    getLocalConversationManager,
    cleanupLocalConversationManager,
  } from "./local-conversation-manager";
  import ChatInput from "@rilldata/web-common/features/chat/core/input/ChatInput.svelte";
  import Messages from "@rilldata/web-common/features/chat/core/messages/Messages.svelte";
  import LocalConversationSidebar from "./LocalConversationSidebar.svelte";
  import {
    conversationSidebarCollapsed,
    toggleConversationSidebar,
  } from "@rilldata/web-common/features/chat/layouts/fullpage/fullpage-store";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context";

  $: ({ instanceId } = $runtime);

  $: conversationManager = getLocalConversationManager(instanceId);

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  // Focus on mount with a small delay for component initialization
  onMount(() => {
    setTimeout(() => {
      chatInputComponent?.focusInput();
    }, 100);
  });

  // Clean up conversation manager resources when leaving the chat context entirely
  beforeNavigate(({ to }) => {
    const isChatRoute = to?.route?.id?.includes("ai");
    if (!isChatRoute) {
      cleanupLocalConversationManager(instanceId);
    }
  });
</script>

<div class="chat-fullpage">
  <!-- Conversation List Sidebar -->
  <LocalConversationSidebar
    {conversationManager}
    collapsed={$conversationSidebarCollapsed}
    onToggle={toggleConversationSidebar}
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
        <Messages
          {conversationManager}
          layout="fullpage"
          config={projectChat}
        />
      </div>
    </div>

    <div class="chat-input-section">
      <div class="chat-input-wrapper">
        <ChatInput
          {conversationManager}
          onSend={onMessageSend}
          bind:this={chatInputComponent}
          config={projectChat}
        />
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
