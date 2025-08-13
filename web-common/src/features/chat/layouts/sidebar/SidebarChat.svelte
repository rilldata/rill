<script lang="ts">
  import { get } from "svelte/store";
  import AlertCircle from "../../../../components/icons/AlertCircle.svelte";
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

  $: listConversationsQuery = chat.listConversationsQuery();
  $: conversations = $listConversationsQuery.data?.conversations ?? [];

  $: pendingMessage = chat.pendingMessage;
  $: currentConversation = chat.getCurrentConversation();
  $: getConversationQuery = $currentConversation?.getConversationQuery();
  $: isSending = $currentConversation?.isSending;

  // Route between optimistic and real messages
  $: displayMessages =
    $currentConversation === null
      ? $pendingMessage
        ? [$pendingMessage]
        : []
      : ($getConversationQuery?.data?.conversation?.messages ?? []);

  // Local UI state
  let chatInputComponent: ChatInput;
  let newConversationDraft = "";

  // Message handling with input focus
  async function handleSend() {
    try {
      if ($currentConversation) {
        // Send message to existing conversation
        await $currentConversation.sendMessage();
      } else {
        // No current conversation, start a new one with the input message
        if (newConversationDraft.trim()) {
          await chat.createConversation(newConversationDraft.trim());
          newConversationDraft = "";
        }
      }

      chatInputComponent?.focusInput();
    } catch (error) {
      console.error("Failed to send message:", error);
    }
  }

  function onNewConversation() {
    chat.enterNewConversationMode();
    chatInputComponent?.focusInput();
  }

  function onSelectConversation(conv: { id: string }) {
    chat.selectConversation(conv.id);
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
        currentConversation={$getConversationQuery?.data?.conversation ?? null}
        {conversations}
        {onNewConversation}
        {onSelectConversation}
        onClose={sidebarActions.closeChat}
      />
    </div>
    <ChatMessages
      layout="sidebar"
      isConversationLoading={!!$getConversationQuery?.isLoading &&
        !$pendingMessage}
      isResponseLoading={$currentConversation ? $isSending : !!$pendingMessage}
      messages={displayMessages}
    />
    {#if $getConversationQuery?.error}
      <div class="chat-input-error">
        <AlertCircle size="1.2em" />
        {$getConversationQuery?.error.message}
      </div>
    {/if}
    <ChatInput
      bind:this={chatInputComponent}
      value={$currentConversation
        ? get($currentConversation.draftMessage)
        : newConversationDraft}
      disabled={$getConversationQuery?.isLoading ||
        ($currentConversation
          ? get($currentConversation.isSending)
          : !!$pendingMessage)}
      onInput={(v) => {
        if ($currentConversation) {
          $currentConversation.draftMessage.set(v);
        } else {
          newConversationDraft = v;
        }
      }}
      onSend={handleSend}
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
