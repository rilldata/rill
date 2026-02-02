<script lang="ts">
  import { page } from "$app/stores";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import Close from "../../../../components/icons/Close.svelte";
  import PlusIcon from "../../../../components/icons/PlusIcon.svelte";
  import { type V1Conversation } from "../../../../runtime-client";
  import { runtime } from "../../../../runtime-client/runtime-store";
  import type { ConversationManager } from "../../core/conversation-manager";
  import ShareChatPopover from "../../share/ShareChatPopover.svelte";
  import ConversationHistoryMenu from "./ConversationHistoryMenu.svelte";

  export let conversationManager: ConversationManager;
  export let onNewConversation: () => void;
  export let onClose: () => void;

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: getConversationQuery = $currentConversationStore?.getConversationQuery();
  $: currentConversationDto = $getConversationQuery?.data?.conversation ?? null;

  $: listConversationsQuery = conversationManager.listConversationsQuery();
  $: conversations = $listConversationsQuery.data?.conversations ?? [];

  function handleNewConversation() {
    conversationManager.enterNewConversationMode();
    onNewConversation();
  }

  function handleSelectConversation(conversation: V1Conversation) {
    conversationManager.selectConversation(conversation.id!);
  }
</script>

<div class="chatbot-header">
  <span class="chatbot-title">{currentConversationDto?.title || ""}</span>
  <div class="chatbot-header-actions">
    <IconButton
      ariaLabel="New conversation"
      bgGray
      on:click={handleNewConversation}
    >
<PlusIcon className="text-gray-500" />
      <svelte:fragment slot="tooltip-content">New conversation</svelte:fragment>
    </IconButton>

    <ShareChatPopover
      conversationId={currentConversationDto?.id}
      {instanceId}
      {organization}
      {project}
      disabled={!currentConversationDto?.id}
    />

    <ConversationHistoryMenu
      {conversations}
      currentConversationId={currentConversationDto?.id}
      onSelect={handleSelectConversation}
    />

    <IconButton ariaLabel="Close chat" bgGray on:click={onClose}>
      <Close className="text-gray-500" />
      <svelte:fragment slot="tooltip-content">Close</svelte:fragment>
    </IconButton>
  </div>
</div>

<style lang="postcss">
  .chatbot-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    font-weight: 500;
    font-size: 0.875rem;
    min-height: 1.5rem;
  }

  .chatbot-title {
    @apply text-fg-secondary text-sm font-semibold;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 180px;
  }

  .chatbot-header-actions {
    display: flex;
    align-items: center;
    gap: 0.125rem;
    flex-shrink: 0;
  }
</style>
