<script lang="ts">
  import { HistoryIcon } from "lucide-svelte";
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

  $: groupedConversations = groupConversationsByDate(conversations);

  function handleOpenChange(isOpen: boolean) {
    if (isOpen) {
      // Using this snapshot prevents the "Current chat" label from jumping during selection
      currentConversationIdSnapshot = currentConversationId;
    }
  }
</script>

<DropdownMenu.Root onOpenChange={handleOpenChange}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      aria-label="Show conversations"
      class="grid place-items-center text-gray-500 hover:bg-gray-200 w-6 h-6"
      style="font-size: 18px;"
    >
      <HistoryIcon size="1em" />
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    class="max-h-80 overflow-y-auto flex flex-col gap-y-1"
    align="end"
  >
    {#if conversations.length === 0}
      <div class="px-3 py-4 text-center text-gray-500 text-sm">
        No conversations yet.
      </div>
    {:else}
      {#each GROUP_ORDER as groupKey}
        {#if groupedConversations[groupKey] && groupedConversations[groupKey].length > 0}
          <DropdownMenu.Group>
            <DropdownMenu.Label class="px-1 text-xs text-gray-500">
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
