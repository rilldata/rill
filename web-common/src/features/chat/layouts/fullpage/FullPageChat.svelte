<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import Messages from "../../core/messages/Messages.svelte";
  import { ShareChatPopover } from "../../share";
  import ConversationSidebar from "./ConversationSidebar.svelte";
  import {
    conversationSidebarCollapsed,
    toggleConversationSidebar,
  } from "./fullpage-store";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context.ts";

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: conversationManager = getConversationManager(instanceId, {
    conversationState: "url",
  });

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: getConversationQuery = $currentConversationStore?.getConversationQuery();
  $: currentConversation = $getConversationQuery?.data?.conversation ?? null;
  $: hasExistingConversation = !!currentConversation?.id;

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  // Focus on mount with a small delay for component initialization
  onMount(() => {
    // Give the component tree time to fully initialize
    setTimeout(() => {
      chatInputComponent?.focusInput();
    }, 100);
  });

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
    collapsed={$conversationSidebarCollapsed}
    onToggle={toggleConversationSidebar}
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
    {#if hasExistingConversation && currentConversation?.id && organization && project}
      <div class="chat-header">
        <ShareChatPopover
          conversationId={currentConversation.id}
          {instanceId}
          {organization}
          {project}
        />
      </div>
    {/if}
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

  /* Main Chat Area */
  .chat-main {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--surface);
  }

  .chat-header {
    position: absolute;
    top: 0;
    right: 0;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 0.5rem 1rem;
    z-index: 10;
    pointer-events: none;
  }

  .chat-header :global(*) {
    pointer-events: auto;
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
