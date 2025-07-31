<script lang="ts">
  import { page } from "$app/stores";
  import AlertCircle from "../../../../components/icons/AlertCircle.svelte";
  import Resizer from "../../../../layout/Resizer.svelte";
  import { useChatCore } from "../../core/chat";
  import ChatFooter from "../../core/input/ChatFooter.svelte";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import ChatMessages from "../../core/messages/ChatMessages.svelte";
  import ChatHeader from "./ChatHeader.svelte";
  import {
    createSidebarConversationIdStore,
    SIDEBAR_DEFAULTS,
    sidebarActions,
    sidebarWidth,
  } from "./sidebar-store";

  // Extract route parameters
  $: organization = $page.params.organization || "";
  $: project = $page.params.project || "";

  // Create project-specific conversation store
  const sidebarConversationId = createSidebarConversationIdStore(
    organization,
    project,
  );

  // Use core chat logic with sidebar-specific state management
  const {
    listConversationsData,
    currentConversation,
    isConversationLoading,
    loading,
    error,
    messages,
    handleSendMessage,
    createNewConversation,
    selectConversation,
  } = useChatCore({
    initialConversationId: $sidebarConversationId,
    onConversationChange: (id) => {
      sidebarConversationId.set(id);
    },
  });

  // Local UI state
  let input = "";
  let chatInputComponent: ChatInput;

  // Message handling with input focus
  async function onSendMessage(message: string) {
    await handleSendMessage(
      message,
      () => chatInputComponent?.focusInput(), // onSuccess - just focus input for sidebar
      (failedMessage) => {
        input = failedMessage;
      }, // onError
    );
  }

  // Conversation actions with input focus
  function onNewConversation() {
    createNewConversation(() => chatInputComponent?.focusInput());
  }

  function onSelectConversation(conv) {
    selectConversation(conv);
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
        currentTitle={$currentConversation?.title || ""}
        conversations={$listConversationsData?.conversations || []}
        currentConversationId={$currentConversation?.id}
        {onNewConversation}
        {onSelectConversation}
        onClose={sidebarActions.closeChat}
      />
    </div>
    <ChatMessages
      layout="sidebar"
      isConversationLoading={$isConversationLoading}
      loading={$loading}
      messages={$messages}
    />
    {#if $error}
      <div class="chat-input-error">
        <AlertCircle size="1.2em" />
        {$error}
      </div>
    {/if}
    <ChatInput
      bind:this={chatInputComponent}
      bind:value={input}
      disabled={$loading}
      onSend={onSendMessage}
    />
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

  .chat-input-error {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: #991b1b;
    background: #fef7f7;
    border-left: 3px solid #f87171;
    border-radius: 0.375rem;
    margin: 0.5rem 1rem 0 1rem;
    box-sizing: border-box;
  }
</style>
