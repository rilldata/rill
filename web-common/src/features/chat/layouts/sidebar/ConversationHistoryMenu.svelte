<script lang="ts">
  import { HistoryIcon } from "lucide-svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import * as DropdownMenu from "../../../../components/dropdown-menu";
  import type { V1Conversation } from "../../../../runtime-client";
  import ConversationHistoryItem from "./ConversationHistoryItem.svelte";
  import {
    GROUP_ORDER,
    groupConversationsByDate,
  } from "./conversation-grouping";

  export let conversations: V1Conversation[] = [];
  export let currentConversationId: string | undefined = undefined;
  export let onSelect: (conversation: V1Conversation) => void;

  let currentConversationIdSnapshot: string | undefined = currentConversationId;
  let isOpen = false;

  $: groupedConversations = groupConversationsByDate(conversations);

  function handleOpenChange(open: boolean) {
    isOpen = open;
    if (open) {
      // Using this snapshot prevents the "Current chat" label from jumping during selection
      currentConversationIdSnapshot = currentConversationId;
    }
  }
</script>

<DropdownMenu.Root onOpenChange={handleOpenChange}>
  <DropdownMenu.Trigger>
    <IconButton
      ariaLabel="Conversation history"
      bgGray
      active={isOpen}
      disableTooltip={isOpen}
    >
      <HistoryIcon size="16px" class="text-gray-500" />
      <svelte:fragment slot="tooltip-content"
        >Conversation history</svelte:fragment
      >
    </IconButton>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    class="max-h-80 overflow-y-auto flex flex-col gap-y-1"
    align="end"
  >
    {#if conversations.length === 0}
      <div class="px-3 py-4 text-center text-fg-secondary text-sm">
        No conversations yet.
      </div>
    {:else}
      {#each GROUP_ORDER as groupKey}
        {#if groupedConversations[groupKey] && groupedConversations[groupKey].length > 0}
          <DropdownMenu.Group>
            <DropdownMenu.Label class="px-1 text-xs text-fg-secondary">
              {groupKey}
            </DropdownMenu.Label>
            {#each groupedConversations[groupKey] as conv}
              <ConversationHistoryItem
                conversation={conv}
                isCurrentChat={conv.id === currentConversationIdSnapshot}
                showRelativeTime={groupKey === "Today"}
                {onSelect}
              />
            {/each}
          </DropdownMenu.Group>
        {/if}
      {/each}
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
