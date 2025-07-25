<script lang="ts">
  import Button from "../../../../components/button/Button.svelte";
  import type { V1Conversation } from "../../../../runtime-client";

  export let conversations: V1Conversation[] = [];
  export let currentConversation: V1Conversation | null = null;
  export let onNewConversation: () => void;
  export let onSelectConversation: (conversation: V1Conversation) => void;
</script>

<div class="conversation-sidebar">
  <div class="conversation-sidebar-header">
    <Button
      type="secondary"
      onClick={onNewConversation}
      class="new-conversation-btn"
    >
      + New conversation
    </Button>
  </div>

  <div class="conversation-list">
    {#if conversations?.length}
      {#each conversations as conversation}
        <button
          class="conversation-item"
          class:active={conversation.id === currentConversation?.id}
          on:click={() => onSelectConversation(conversation)}
        >
          <div class="conversation-title">
            {conversation.title || "New conversation"}
          </div>
        </button>
      {/each}
    {:else}
      <div class="no-conversations">No conversations yet</div>
    {/if}
  </div>
</div>

<style lang="postcss">
  /* Conversation Sidebar */
  .conversation-sidebar {
    width: 280px;
    background: #f8f9fa;
    border-right: 1px solid #e5e7eb;
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
  }

  .conversation-sidebar-header {
    padding: 0.75rem;
    border-bottom: 1px solid #e5e7eb;
  }

  /* Custom full-width styling that preserves small height */
  :global(.new-conversation-btn) {
    width: 100% !important;
  }

  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.25rem;
  }

  .conversation-item {
    width: 100%;
    padding: 0.5rem 0.75rem;
    margin-bottom: 0.125rem;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    text-align: left;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .conversation-item:hover {
    background: #e5e7eb;
  }

  .conversation-item.active {
    @apply bg-theme-50 border border-theme-300;
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
