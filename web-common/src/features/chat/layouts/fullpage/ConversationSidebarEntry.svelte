<script lang="ts">
  import type { V1Conversation } from "@rilldata/web-common/runtime-client";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import DeleteConversationDialog from "@rilldata/web-common/features/chat/core/DeleteConversationDialog.svelte";
  import { renameConversation } from "@rilldata/web-common/features/chat/core/conversation-actions.ts";
  import { stripPartialChatReference } from "@rilldata/web-common/features/chat/core/context/conversation-title.ts";
  import InlineTitleEditor from "@rilldata/web-common/features/chat/core/input/InlineTitleEditor.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Pencil } from "lucide-svelte";

  let {
    conversation,
    active,
    basePath,
    getTitle,
    onClick,
    onDeleted,
  }: {
    conversation: V1Conversation;
    active: boolean;
    basePath: string;
    getTitle: (conversation: V1Conversation) => string;
    onClick: () => void;
    onDeleted: () => void;
  } = $props();

  let editing = $state(false);
  let confirmingDelete = $state(false);

  let title = $derived(getTitle(conversation));

  const runtimeClient = useRuntimeClient();

  async function rename(newTitle: string) {
    editing = false;
    if (!newTitle || newTitle === conversation.title) return;

    try {
      await renameConversation(runtimeClient, conversation.id!, newTitle);
    } catch (err) {
      console.error("Failed to rename conversation", err);
      eventBus.emit("notification", {
        message: m.chat_conversation_rename_failed(),
        type: "error",
      });
    }
  }
</script>

<div
  class="conversation-item"
  class:active
  class:editing
  data-testid="conversation-item"
  data-conversation-id={conversation.id}
>
  {#if editing}
    <div class="conversation-title-editor">
      <InlineTitleEditor
        value={stripPartialChatReference(conversation.title ?? "")}
        onSave={rename}
        onCancel={() => (editing = false)}
      />
    </div>
  {:else}
    <a
      href={`${basePath}/${conversation.id}`}
      class="conversation-link"
      {title}
      onclick={onClick}
    >
      <div class="conversation-title" data-testid="conversation-title">
        {title}
      </div>
    </a>
    <div class="conversation-actions">
      <Button
        square
        small
        type="tertiary"
        label={m.chat_conversation_rename()}
        onClick={() => (editing = true)}
      >
        <Pencil size="14px" />
      </Button>
      <Button
        square
        small
        type="tertiary"
        label={m.chat_conversation_delete()}
        onClick={() => (confirmingDelete = true)}
      >
        <Trash size="14px" />
      </Button>
    </div>
  {/if}
</div>

<DeleteConversationDialog
  conversationId={conversation.id!}
  bind:open={confirmingDelete}
  {onDeleted}
/>

<style lang="postcss">
  .conversation-item {
    @apply flex items-center w-full mb-0.5 pr-1;
    @apply rounded-md transition-colors duration-200;
  }

  .conversation-item:hover,
  .conversation-item.editing {
    @apply bg-surface-muted;
  }

  .conversation-item.active {
    @apply bg-gray-100;
  }

  .conversation-link {
    @apply flex-1 min-w-0 py-2 pl-3;
    @apply text-left cursor-pointer no-underline;
    color: inherit;
  }

  .conversation-title {
    @apply text-xs text-fg-primary truncate;
  }

  .conversation-title-editor {
    @apply flex-1 min-w-0 py-1.5 pl-3 pr-2;
  }

  .conversation-actions {
    @apply hidden shrink-0 items-center;
  }

  .conversation-item:hover .conversation-actions,
  .conversation-item:focus-within .conversation-actions {
    @apply flex;
  }

  @media (max-width: 640px) {
    .conversation-item {
      flex-shrink: 0;
      min-width: 150px;
    }
  }
</style>
