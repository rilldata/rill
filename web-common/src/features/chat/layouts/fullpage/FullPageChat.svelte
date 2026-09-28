<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context.ts";
  import { onMount } from "svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import ReadOnlyChatInput from "../../core/input/ReadOnlyChatInput.svelte";
  import Messages from "../../core/messages/Messages.svelte";
  import ConnectClientPopover from "../../connect/ConnectClientPopover.svelte";
  import ShareChatPopover from "../../share/ShareChatPopover.svelte";
  import ConversationSidebar from "./ConversationSidebar.svelte";
  import {
    conversationSidebarCollapsed,
    toggleConversationSidebar,
  } from "./fullpage-store";

  // Read-only mode is for visitors who can't send messages, e.g. anonymous recipients of an AI report opened with a magic token.
  // It hides the conversation sidebar and the header actions, and replaces the input with a notice.
  export let readOnly = false;
  // Awaited before a shared conversation is forked (see Conversation.sendMessage).
  export let beforeFork: (() => Promise<void> | void) | undefined = undefined;

  const { adminServer } = featureFlags;

  const runtimeClient = useRuntimeClient();
  $: instanceId = runtimeClient.instanceId;
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: conversationManager = getConversationManager(runtimeClient, {
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
  {#if !readOnly}
    <ConversationSidebar
      {conversationManager}
      basePath={`/${organization}/${project}/-/ai`}
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
      <svelte:fragment slot="collapsed-footer">
        <slot name="sidebar-collapsed-footer" />
      </svelte:fragment>
    </ConversationSidebar>
  {/if}

  <!-- Main Chat Area -->
  <div class="chat-main">
    {#if $adminServer && !readOnly}
      <div class="chat-header">
        <ConnectClientPopover />
        {#if currentConversation?.id}
          <ShareChatPopover
            conversationId={currentConversation.id}
            {organization}
            {project}
          />
        {/if}
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
        {#if readOnly}
          <ReadOnlyChatInput />
        {:else}
          <ChatInput
            {conversationManager}
            onSend={onMessageSend}
            bind:this={chatInputComponent}
            config={projectChat}
            {beforeFork}
          />
        {/if}
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
    @apply flex items-center justify-end gap-x-0.5;
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
