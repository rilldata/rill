<script lang="ts">
  import Resizer from "../../../../layout/Resizer.svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { Chat } from "../../core/chat";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import ChatMessages from "../../core/messages/ChatMessages.svelte";
  import ChatHeader from "./ChatHeader.svelte";
  import {
    SIDEBAR_DEFAULTS,
    sidebarActions,
    sidebarWidth,
  } from "./sidebar-store";

  $: ({ instanceId } = $runtime);

  // Initialize chat with browser storage for conversation management
  $: chat = new Chat(instanceId, {
    conversationState: "browserStorage",
  });

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  function onNewConversation() {
    chatInputComponent?.focusInput();
  }
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
      <ChatHeader
        {chat}
        {onNewConversation}
        onClose={sidebarActions.closeChat}
      />
    </div>
    <ChatMessages {chat} layout="sidebar" />
    <ChatInput {chat} bind:this={chatInputComponent} onSend={onMessageSend} />
    <ChatFooter />
  </div>
</div>

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
