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
  import ShareChatPopover from "../../share/ShareChatPopover.svelte";
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
    <div class="chat-header">
      <ShareChatPopover
        conversationId={currentConversation?.id}
        {instanceId}
        {organization}
        {project}
        disabled={!currentConversation?.id}
      />
    </div>
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
    @apply flex h-full w-full;
    background: var(--surface);
  }

  .chat-main {
    @apply relative flex-1 flex flex-col overflow-hidden;
    background: var(--surface);
  }

  .chat-header {
    @apply absolute top-0 right-0;
    @apply flex items-center justify-end;
    @apply py-2 px-4 z-10 pointer-events-none;
  }

  .chat-header :global(*) {
    @apply pointer-events-auto;
  }

  .chat-content {
    @apply flex-1 overflow-hidden flex flex-col;
    background: var(--surface);
  }

  .chat-messages-wrapper {
    @apply flex-1 overflow-y-auto w-full flex flex-col;
  }

  .chat-input-section {
    @apply shrink-0 p-4 flex justify-center;
    background: var(--surface);
  }

  .chat-input-wrapper {
    @apply w-full max-w-3xl flex flex-col gap-2;
  }

  @media (max-width: 768px) {
    .chat-messages-wrapper,
    .chat-input-wrapper {
      max-width: none;
      padding-left: 1rem;
      padding-right: 1rem;
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
