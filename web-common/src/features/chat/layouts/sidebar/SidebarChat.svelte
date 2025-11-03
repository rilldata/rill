<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import Resizer from "../../../../layout/Resizer.svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import Messages from "../../core/messages/Messages.svelte";
  import SidebarHeader from "./SidebarHeader.svelte";
  import {
    SIDEBAR_DEFAULTS,
    sidebarActions,
    sidebarWidth,
  } from "./sidebar-store";

  $: ({ instanceId } = $runtime);

  // Initialize conversation manager with browser storage for conversation management
  $: conversationManager = getConversationManager(instanceId, {
    conversationState: "browserStorage",
  });

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  function onNewConversation() {
    chatInputComponent?.focusInput();
  }

  // Clean up conversation manager resources when switching projects
  beforeNavigate(({ from, to }) => {
    const currentProject = from?.params?.project;
    const targetProject = to?.params?.project;

    if (currentProject !== targetProject) {
      cleanupConversationManager(instanceId);
    }
  });
</script>

<div class="chat-sidebar" style="--sidebar-width: {$sidebarWidth}px;">
  <Resizer
    min={SIDEBAR_DEFAULTS.MIN_SIDEBAR_WIDTH}
    max={SIDEBAR_DEFAULTS.MAX_SIDEBAR_WIDTH}
    basis={SIDEBAR_DEFAULTS.SIDEBAR_WIDTH}
    dimension={$sidebarWidth}
    direction="EW"
    side="left"
    onUpdate={sidebarActions.updateSidebarWidth}
  />
  <div class="chat-sidebar-content">
    <div class="chatbot-header-container">
      <SidebarHeader
        {conversationManager}
        {onNewConversation}
        onClose={sidebarActions.closeChat}
      />
    </div>
    <Messages {conversationManager} layout="sidebar" />
    <ChatInput
      {conversationManager}
      bind:this={chatInputComponent}
      onSend={onMessageSend}
    />
    <ChatFooter />
  </div>
</div>

<style lang="postcss">
  .chat-sidebar {
    position: relative;
    width: var(--sidebar-width);
    height: 100%;
    background: var(--surface);
    border-left: 1px solid var(--border);
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
