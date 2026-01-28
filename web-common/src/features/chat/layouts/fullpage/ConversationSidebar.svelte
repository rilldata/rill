<script lang="ts">
  import { page } from "$app/stores";
  import Button from "../../../../components/button/Button.svelte";
  import HideSidebar from "../../../../components/icons/HideSidebar.svelte";
  import PlusIcon from "../../../../components/icons/PlusIcon.svelte";
  import DelayedContent from "../../../entity-management/DelayedContent.svelte";
  import Spinner from "../../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../../entity-management/types";
  import type { ConversationManager } from "../../core/conversation-manager";

  export let conversationManager: ConversationManager;
  export let collapsed = false;
  export let onToggle: () => void = () => {};
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

<div class="conversation-sidebar" class:collapsed>
  {#if collapsed}
    <!-- Collapsed state: icon-only buttons -->
    <div class="collapsed-actions">
      <span title="Expand sidebar">
        <Button type="secondary" square onClick={onToggle}>
          <HideSidebar side="left" open={false} size="16px" />
        </Button>
      </span>
      <span title="New conversation">
        <Button
          type="secondary"
          square
          href={`/${organization}/${project}/-/ai?new=true`}
          onClick={handleNewConversationButtonClick}
        >
          <PlusIcon size="14px" />
        </Button>
      </span>
    </div>
  {:else}
    <!-- Expanded state: full sidebar -->
    <div class="conversation-sidebar-header">
      <div class="header-row">
        <span title="Collapse sidebar">
          <Button type="secondary" square onClick={onToggle}>
            <HideSidebar side="left" open={true} size="16px" />
          </Button>
        </span>
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
  {/if}
</div>

<style lang="postcss">
  .conversation-sidebar {
    @apply flex flex-col shrink-0 min-h-0 overflow-hidden;
    @apply bg-surface-subtle border-r border-border;
    @apply transition-[width] duration-200 ease-in-out;
    width: 280px;
  }

  .conversation-sidebar.collapsed {
    width: 56px;
  }

  .collapsed-actions {
    @apply flex flex-col gap-2 p-3 items-center;
  }

  .conversation-sidebar-header {
    @apply p-3 border-b border-border;
  }

  .header-row {
    @apply flex gap-2 items-center;
  }

  :global(.new-conversation-btn) {
    flex: 1 !important;
  }

  .conversation-list {
    @apply flex-1 overflow-y-auto p-1 min-h-0;
  }

  .conversation-sidebar-footer {
    @apply shrink-0 p-3 border-t border-border mt-auto;
  }

  .loading-conversations {
    @apply p-2 flex justify-center items-center;
  }

  .conversation-item {
    @apply block w-full py-2 px-3 mb-0.5;
    @apply bg-transparent border-none rounded-md;
    @apply text-left cursor-pointer no-underline;
    color: inherit;
    font-family: inherit;
    font-size: inherit;
    @apply transition-colors duration-200;
  }

  .conversation-item:hover {
    @apply bg-surface-muted;
  }

  .conversation-item.active {
    @apply bg-gray-100;
  }

  .conversation-title {
    @apply text-xs text-fg-primary truncate;
  }

  .no-conversations {
    @apply py-6 px-4 text-center text-fg-secondary text-xs;
  }

  /* Responsive behavior */
  @media (max-width: 768px) {
    .conversation-sidebar:not(.collapsed) {
      width: 240px;
    }
  }

  @media (max-width: 640px) {
    .conversation-sidebar:not(.collapsed) {
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
