<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { onMount } from "svelte";
  import Resizer from "../../../../layout/Resizer.svelte";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import { getConversationManager } from "../../core/conversation-manager";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import Messages from "../../core/messages/Messages.svelte";
  import SidebarHeader from "./SidebarHeader.svelte";
  import {
    SIDEBAR_DEFAULTS,
    sidebarActions,
    sidebarWidth,
  } from "./sidebar-store";

  import type { ChatConfig } from "@rilldata/web-common/features/chat/core/types.ts";

  export let config: ChatConfig;

  $: ({ instanceId } = $runtime);

  // Initialize conversation manager with browser storage for conversation management
  $: conversationManager = getConversationManager(instanceId, {
    conversationState: "browserStorage",
    agent: config.agent,
  });

  let chatInputComponent: ChatInput;

  function onMessageSend() {
    chatInputComponent?.focusInput();
  }

  function onNewConversation() {
    chatInputComponent?.focusInput();
  }

  // Clean up conversation manager resources when switching projects or dashboards
  beforeNavigate(({ from, to }) => {
    const currentProject = from?.params?.project;
    const targetProject = to?.params?.project;
    const currentDashboard = from?.params?.dashboard ?? from?.params?.name;
    const targetDashboard = to?.params?.dashboard ?? to?.params?.name;

    // Clear conversation when switching projects OR when switching dashboards within the same project
    if (
      currentProject !== targetProject ||
      (currentProject === targetProject && currentDashboard !== targetDashboard)
    ) {
      conversationManager.enterNewConversationMode();
    }
  });

  onMount(() => {
    chatInputComponent?.focusInput();
  });
</script>

<div
  class="chat-sidebar"
  style="--sidebar-width: {$sidebarWidth}px;"
  on:click|stopPropagation
  role="presentation"
>
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
    <Messages {conversationManager} layout="sidebar" {config} />
    <ChatInput
      {conversationManager}
      bind:this={chatInputComponent}
      onSend={onMessageSend}
      {config}
    />
  </div>
</div>

<style lang="postcss">
  .chat-sidebar {
    @apply flex flex-col relative h-full bg-surface border;
    width: var(--sidebar-width);
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
