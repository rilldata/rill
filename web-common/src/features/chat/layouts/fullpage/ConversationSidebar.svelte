<script lang="ts">
  import { page } from "$app/stores";
  import Button from "../../../../components/button/Button.svelte";
  import PlusIcon from "../../../../components/icons/PlusIcon.svelte";
  import DelayedContent from "../../../entity-management/DelayedContent.svelte";
  import Spinner from "../../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../../entity-management/types";
  import type { ConversationManager } from "../../core/conversation-manager";

  export let conversationManager: ConversationManager;
  export let onConversationClick: () => void = () => {};
  export let onNewConversationClick: () => void = () => {};

  // Get URL parameters for href construction
  $: ({ organization, project } = $page.params);

  $: currentConversation = conversationManager.getCurrentConversation();
  $: getConversationQuery = $currentConversation?.getConversationQuery();
  $: currentConversationDto = $getConversationQuery?.data?.conversation ?? null;

  $: listConversationsQuery = conversationManager.listConversationsQuery();

  $: conversations = $listConversationsQuery.data?.conversations ?? [];
  $: isLoading = $listConversationsQuery.isLoading;
  $: isError = $listConversationsQuery.isError;

  // Handle conversation item clicks (for focus, navigation handled by href)
  function handleConversationItemClick() {
    onConversationClick();
  }

  // Handle new conversation button click (for focus, navigation handled by href)
  function handleNewConversationButtonClick() {
    conversationManager.enterNewConversationMode();
    onNewConversationClick();
  }
</script>

<div class="conversation-sidebar">
  <div class="conversation-sidebar-header">
    <Button
      type="secondary"
      href={`/${organization}/${project}/-/ai?new=true`}
      class="new-conversation-btn"
      onClick={handleNewConversationButtonClick}
    >
      <PlusIcon size="12px" />
      New conversation
    </Button>
  </div>

  <div class="conversation-list" data-testid="conversation-list">
    {#if isLoading}
      <div class="loading-conversations">
        <DelayedContent visible={isLoading} delay={300}>
          <div class="flex flex-row items-center gap-x-2">
            <Spinner size="1em" status={EntityStatus.Running} />
            Loading conversations...
          </div>
        </DelayedContent>
      </div>
    {:else if isError}
      <div class="error-conversations">Error loading conversations</div>
    {:else if conversations.length}
      {#each conversations as conversation}
        <a
          href={`/${organization}/${project}/-/ai/${conversation.id}`}
          class="conversation-item"
          class:active={conversation.id === currentConversationDto?.id}
          data-testid="conversation-item"
          data-conversation-id={conversation.id}
          on:click={handleConversationItemClick}
        >
          <div class="conversation-title" data-testid="conversation-title">
            {conversation.title || "New conversation"}
          </div>
        </a>
      {/each}
    {:else}
      <div class="no-conversations" data-testid="no-conversations">
        No conversations yet
      </div>
    {/if}
  </div>

  <!-- Footer slot for additional actions (e.g., MCP config button) -->
  <div class="conversation-sidebar-footer">
    <slot name="footer" />
  </div>
</div>

<style lang="postcss">
  .conversation-sidebar {
    width: 280px;
    background: var(--surface);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    min-height: 0;
  }

  .conversation-sidebar-header {
    padding: 0.75rem;
    border-bottom: 1px solid var(--border);
  }

  /* Custom full-width styling that preserves small height */
  :global(.new-conversation-btn) {
    width: 100% !important;
  }

  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.25rem;
    min-height: 0;
  }

  .conversation-sidebar-footer {
    flex-shrink: 0;
    padding: 0.75rem;
    border-top: 1px solid var(--border);
    margin-top: auto;
  }

  .loading-conversations {
    padding: 0.5rem;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .conversation-item {
    display: block;
    width: 100%;
    padding: 0.5rem 0.75rem;
    margin-bottom: 0.125rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    text-align: left;
    cursor: pointer;
    transition: background-color 0.2s;
    text-decoration: none;
    color: inherit;
    font-family: inherit;
    font-size: inherit;
  }

  .conversation-item:hover {
    background: var(--muted);
  }

  .conversation-item.active {
    @apply bg-gray-100;
  }

  .conversation-title {
    font-size: 0.8rem;
    color: #374151;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .no-conversations {
    padding: 1.5rem 1rem;
    text-align: center;
    color: #6b7280;
    font-size: 0.8rem;
  }

  /* Responsive behavior */
  @media (max-width: 768px) {
    .conversation-sidebar {
      width: 240px;
    }
  }

  @media (max-width: 640px) {
    .conversation-sidebar {
      width: 100%;
      height: 200px;
    }

    .conversation-list {
      display: flex;
      flex-direction: row;
      overflow-x: auto;
      gap: 0.25rem;
    }

    .conversation-item {
      flex-shrink: 0;
      min-width: 150px;
    }
  }
</style>
