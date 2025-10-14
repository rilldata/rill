<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { onMount } from "svelte";
  import Resizer from "../../../../layout/Resizer.svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { cleanupChatInstance, getChatInstance } from "../../core/chat";
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
  $: chat = getChatInstance(instanceId, {
    conversationState: "browserStorage",
  });

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  function onNewConversation() {
    chatInputComponent?.focusInput();
  }

  // Clean up chat resources when switching projects
  beforeNavigate(({ from, to }) => {
    const currentProject = from?.params?.project;
    const targetProject = to?.params?.project;

    if (currentProject !== targetProject) {
      cleanupChatInstance(instanceId);
    }
  });

  onMount(() => {
    chatInputComponent?.focusInput();
    return eventBus.on("start-chat", onNewConversation);
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
