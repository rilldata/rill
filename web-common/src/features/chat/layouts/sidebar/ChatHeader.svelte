<script lang="ts">
  import IconButton from "../../../../components/button/IconButton.svelte";
  import Close from "../../../../components/icons/Close.svelte";
  import PlusIcon from "../../../../components/icons/PlusIcon.svelte";
  import { type V1Conversation } from "../../../../runtime-client";
  import type { Chat } from "../../core/chat";
  import ChatConversationDropdown from "./ChatConversationDropdown.svelte";
  import ChatContextOptions from "./ChatContextOptions.svelte";
  import type { StateManagers } from "../../../dashboards/state-managers/state-managers";

  export let chat: Chat;
  export let onNewConversation: () => void;
  export let onClose: () => void;
  export let stateManagers: StateManagers | undefined = undefined;
  export let includeFilters: boolean = true;
  export let includeTimeRange: boolean = true;
  export let onContextOptionsChange: (options: {
    includeFilters: boolean;
    includeTimeRange: boolean;
  }) => void = () => {};

  $: currentConversationStore = chat.getCurrentConversation();
  $: getConversationQuery = $currentConversationStore?.getConversationQuery();
  $: currentConversationDto = $getConversationQuery?.data?.conversation ?? null;

  $: listConversationsQuery = chat.listConversationsQuery();
  $: conversations = $listConversationsQuery.data?.conversations ?? [];

  function handleNewConversation() {
    chat.enterNewConversationMode();
    onNewConversation();
  }

  function handleSelectConversation(conversation: V1Conversation) {
    chat.selectConversation(conversation.id!);
  }
</script>

<div class="chatbot-header">
  <span class="chatbot-title">{currentConversationDto?.title || ""}</span>
  <div class="chatbot-header-actions">
    <ChatContextOptions
      {stateManagers}
      {includeFilters}
      {includeTimeRange}
      onOptionsChange={onContextOptionsChange}
    />

    <IconButton
      ariaLabel="New conversation"
      bgGray
      on:click={handleNewConversation}
    >
      <PlusIcon className="text-gray-500" />
    </IconButton>

    <ChatConversationDropdown
      {conversations}
      currentConversationId={currentConversationDto?.id}
      onSelect={handleSelectConversation}
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
