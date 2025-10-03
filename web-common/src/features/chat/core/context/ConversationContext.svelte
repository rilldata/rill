<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";

  export let conversation: Conversation;

  $: contextData = conversation.context.context;
</script>

{#if $contextData?.length}
  <div class="flex flex-row gap-x-1 items-center pb-2">
    {#each $contextData as { type, value } (type)}
      <Chip>
        <span slot="body">{value}</span>
      </Chip>
    {/each}
    <Button compact noStroke onClick={() => conversation.context.clear()}
      >Clear</Button
    >
  </div>
{/if}
