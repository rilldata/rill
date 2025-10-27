<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ConversationContextEntryDisplay from "@rilldata/web-common/features/chat/core/context/ConversationContextEntryDisplay.svelte";
  import { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";

  export let conversation: Conversation;

  $: context = conversation.context;
  $: contextData = context.data;
</script>

{#if $contextData?.length}
  <div class="flex flex-wrap gap-1 items-center mx-1 mt-1">
    {#each $contextData as entry (entry.type)}
      <ConversationContextEntryDisplay {context} {entry} />
    {/each}
    <Button compact noStroke onClick={() => conversation.context.clear()}>
      Clear
    </Button>
  </div>
{/if}
