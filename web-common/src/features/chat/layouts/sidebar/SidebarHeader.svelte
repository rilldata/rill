<script lang="ts">
  import { page } from "$app/stores";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import Close from "../../../../components/icons/Close.svelte";
  import PlusIcon from "../../../../components/icons/PlusIcon.svelte";
  import { renameConversation } from "@rilldata/web-common/features/chat/core/conversation-actions.ts";
  import {
    formatConversationTitle,
    stripPartialChatReference,
  } from "@rilldata/web-common/features/chat/core/context/conversation-title.ts";
  import { getInlineChatContextMetadata } from "@rilldata/web-common/features/chat/core/context/metadata.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import InlineTitleEditor from "@rilldata/web-common/features/chat/core/input/InlineTitleEditor.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { type V1Conversation } from "../../../../runtime-client";
  import ConnectClientPopover from "../../connect/ConnectClientPopover.svelte";
  import MemoryButton from "../../memory/MemoryButton.svelte";
  import McpServersPopover from "../../connect/McpServersPopover.svelte";
  import type { ConversationManager } from "../../core/conversation-manager";
  import ShareChatPopover from "../../share/ShareChatPopover.svelte";
  import ConversationHistoryMenu from "./ConversationHistoryMenu.svelte";
  import ConversationTitleMenu from "./ConversationTitleMenu.svelte";

  export let conversationManager: ConversationManager;
  export let onNewConversation: () => void;
  export let onClose: () => void;

  const { adminServer } = featureFlags;
  const runtimeClient = useRuntimeClient();
  const contextMetadata = getInlineChatContextMetadata(runtimeClient);

  let editing = false;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: getConversationQuery = $currentConversationStore?.getConversationQuery();
  $: currentConversationDto = $getConversationQuery?.data?.conversation ?? null;
  $: title = formatConversationTitle(
    currentConversationDto?.title ?? "",
    $contextMetadata,
  );

  $: listConversationsQuery = conversationManager.listConversationsQuery();
  $: conversations = ($listConversationsQuery.data?.conversations ?? []).filter(
    (c) => c.userAgent !== "rill/report",
  );

  function handleNewConversation() {
    conversationManager.enterNewConversationMode();
    onNewConversation();
  }

  function handleSelectConversation(conversation: V1Conversation) {
    conversationManager.selectConversation(conversation.id!);
  }

  async function handleRename(newTitle: string) {
    editing = false;
    const conversation = currentConversationDto;
    if (!conversation?.id || !newTitle || newTitle === conversation.title)
      return;

    try {
      await renameConversation(runtimeClient, conversation.id, newTitle);
    } catch (err) {
      console.error("Failed to rename conversation", err);
      eventBus.emit("notification", {
        message: m.chat_conversation_rename_failed(),
        type: "error",
      });
    }
  }
</script>

<div class="chatbot-header">
  {#if editing && currentConversationDto?.id}
    <!-- Takes the whole header while editing, since references need more room than the title. -->
    <div class="chatbot-title-editor">
      <InlineTitleEditor
        value={stripPartialChatReference(currentConversationDto.title ?? "")}
        size="sm"
        onSave={handleRename}
        onCancel={() => (editing = false)}
      />
    </div>
  {:else}
    <div class="chatbot-title">
      {#if currentConversationDto?.id}
        <ConversationTitleMenu
          conversationId={currentConversationDto.id}
          {title}
          onRename={() => (editing = true)}
          onDeleted={() => conversationManager.enterNewConversationMode()}
        />
      {:else}
        <span class="truncate">{title}</span>
      {/if}
    </div>
    <div class="chatbot-header-actions">
      <MemoryButton />
      <McpServersPopover />
      <ConnectClientPopover />

      <IconButton
        ariaLabel={m.chat_new_conversation()}
        bgGray
        onclick={handleNewConversation}
      >
        <PlusIcon className="text-fg-muted" />
        <svelte:fragment slot="tooltip-content"
          >{m.chat_new_conversation()}</svelte:fragment
        >
      </IconButton>

      {#if $adminServer}
        <ShareChatPopover
          conversationId={currentConversationDto?.id}
          {organization}
          {project}
          disabled={!currentConversationDto?.id}
        />
      {/if}

      <ConversationHistoryMenu
        {conversations}
        contextMetadata={$contextMetadata}
        currentConversationId={currentConversationDto?.id}
        onSelect={handleSelectConversation}
      />

      <IconButton ariaLabel={m.chat_close()} bgGray onclick={onClose}>
        <Close className="text-fg-muted" />
        <svelte:fragment slot="tooltip-content"
          >{m.common_close()}</svelte:fragment
        >
      </IconButton>
    </div>
  {/if}
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
    @apply flex min-w-0 max-w-80 text-fg-secondary text-sm font-semibold;
  }

  .chatbot-title-editor {
    @apply flex-1 min-w-0 min-h-6;
  }

  .chatbot-header-actions {
    display: flex;
    align-items: center;
    gap: 0.125rem;
    flex-shrink: 0;
  }
</style>
