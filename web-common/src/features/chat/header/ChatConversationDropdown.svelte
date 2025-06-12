<script lang="ts">
  import { HistoryIcon } from "lucide-svelte";
  import * as DropdownMenu from "../../../components/dropdown-menu";
  import type { V1Conversation } from "../../../runtime-client";
  import { GROUP_ORDER, groupConversationsByDate } from "../utils/date-utils";
  import ChatConversationItem from "./ChatConversationItem.svelte";

  export let conversations: V1Conversation[] = [];
  export let currentConversationId: string | undefined = undefined;
  export let onSelect: (conversation: V1Conversation) => void;

  $: groupedConversations = groupConversationsByDate(conversations);
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      aria-label="Show conversations"
      class="grid place-items-center text-gray-500 hover:bg-gray-50 w-6 h-6"
      style="font-size: 18px;"
    >
      <HistoryIcon size="1em" />
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content class="max-h-96" align="end">
    <div class="max-h-80 overflow-y-auto">
      {#if conversations.length === 0}
        <div class="px-3 py-4 text-center text-gray-500 text-sm">
          No conversations yet.
        </div>
      {:else}
        {#each GROUP_ORDER as groupKey}
          {#if groupedConversations[groupKey] && groupedConversations[groupKey].length > 0}
            <DropdownMenu.Group>
              <DropdownMenu.Label class="px-1">
                {groupKey}
              </DropdownMenu.Label>
              {#each groupedConversations[groupKey] as conv}
                <ChatConversationItem
                  conversation={conv}
                  isCurrentChat={conv.id === currentConversationId}
                  showRelativeTime={groupKey === "Today"}
                  {onSelect}
                />
              {/each}
            </DropdownMenu.Group>
          {/if}
        {/each}
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
