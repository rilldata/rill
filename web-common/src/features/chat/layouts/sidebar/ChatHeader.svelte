<script lang="ts">
  import IconButton from "../../../../components/button/IconButton.svelte";
  import Close from "../../../../components/icons/Close.svelte";
  import PlusIcon from "../../../../components/icons/PlusIcon.svelte";
  import type { V1Conversation } from "../../../../runtime-client";
  import ChatConversationDropdown from "./ChatConversationDropdown.svelte";

  export let currentConversation: V1Conversation | null;
  export let conversations: V1Conversation[] = [];
  export let onNewConversation: () => void;
  export let onSelectConversation: (conversation: V1Conversation) => void;
  export let onClose: () => void;
</script>

<div class="chatbot-header">
  <span class="chatbot-title">{currentConversation?.title || ""}</span>
  <div class="chatbot-header-actions">
    <IconButton
      ariaLabel="New conversation"
      bgGray
      on:click={onNewConversation}
    >
      <PlusIcon className="text-gray-500" />
    </IconButton>

    <ChatConversationDropdown
      {conversations}
      currentConversationId={currentConversation?.id}
      onSelect={onSelectConversation}
    />

    <IconButton ariaLabel="Close chat" bgGray on:click={onClose}>
      <Close className="text-gray-500" />
    </IconButton>
  </div>
</div>

<style>
  .chatbot-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    background: #fafafa;
    font-weight: 500;
    font-size: 0.875rem;
    min-height: 1.5rem;
  }
  .chatbot-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 180px;
    color: #111827;
    font-size: 0.75rem;
  }
  .chatbot-header-actions {
    display: flex;
    align-items: center;
    gap: 0.125rem;
    flex-shrink: 0;
  }
</style>
