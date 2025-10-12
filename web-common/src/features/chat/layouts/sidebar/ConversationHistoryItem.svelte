<script lang="ts">
  import * as DropdownMenu from "../../../../components/dropdown-menu";
  import { getRelativeTime } from "../../../../lib/time/relative-time";
  import type { V1Conversation } from "../../../../runtime-client";

  export let conversation: V1Conversation;
  export let isCurrentChat: boolean = false;
  export let showRelativeTime: boolean = false;
  export let onSelect: (conversation: V1Conversation) => void;

  $: relativeTime = showRelativeTime
    ? getRelativeTime(conversation.updatedOn || conversation.createdOn || "")
    : "";

  function handleClick() {
    onSelect(conversation);
  }
</script>

<DropdownMenu.Item
  class="flex items-center gap-2 p-1 cursor-pointer"
  on:click={handleClick}
>
  <div class="min-w-0">
    <div class="text-xs text-gray-900 truncate">
      {conversation.title || "New Conversation"}
    </div>
  </div>
  <div class="flex items-center gap-2 flex-shrink-0">
    {#if relativeTime && !isCurrentChat}
      <span class="text-xs text-gray-500">{relativeTime}</span>
    {/if}
    {#if isCurrentChat}
      <span class="text-xs text-gray-500">Current chat</span>
    {/if}
  </div>
</DropdownMenu.Item>
