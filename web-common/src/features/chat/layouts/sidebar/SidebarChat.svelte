<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
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
  import type { StateManagers } from "../../../dashboards/state-managers/state-managers";
  import { writable } from "svelte/store";

  export let stateManagers: StateManagers | undefined = undefined;

  $: ({ instanceId } = $runtime);

  // Context options state (default to true)
  let includeFilters = true;
  let includeTimeRange = true;

  // Create a reactive store for context options that the chat can subscribe to
  const contextOptionsStore = writable({ includeFilters, includeTimeRange });
  $: contextOptionsStore.set({ includeFilters, includeTimeRange });

  // Initialize chat with browser storage for conversation management
  $: chat = getChatInstance(instanceId, {
    conversationState: "browserStorage",
    dashboardContext: stateManagers,
    contextOptions: contextOptionsStore,
  });

  let chatInputComponent: ChatInput;

  function handleContextOptionsChange(options: {
    includeFilters: boolean;
    includeTimeRange: boolean;
  }) {
    includeFilters = options.includeFilters;
    includeTimeRange = options.includeTimeRange;
  }

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
        {stateManagers}
        {includeFilters}
        {includeTimeRange}
        {onNewConversation}
        onClose={sidebarActions.closeChat}
        onContextOptionsChange={handleContextOptionsChange}
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
