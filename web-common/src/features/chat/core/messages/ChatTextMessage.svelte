<script lang="ts">
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";

  export let message: V1Message;
  export let content: string;

  $: role = message.role;
</script>

<div class="chat-message chat-message--{role}">
  <div class="chat-message-content">
    {#if role === "assistant"}
      <Markdown {content} />
    {:else}
      {content}
    {/if}
  </div>
</div>

<style lang="postcss">
  .chat-message {
    max-width: 90%;
  }

  .chat-message--user {
    align-self: flex-end;
  }

  .chat-message--assistant {
    align-self: flex-start;
  }

  .chat-message-content {
    padding: 0.375rem 0.5rem;
    border-radius: 1rem;
    font-size: 0.875rem;
    line-height: 1.5;
    word-break: break-word;
  }

  .chat-message--user .chat-message-content {
    @apply bg-primary-400 text-white rounded-br-lg;
  }

  .chat-message--assistant .chat-message-content {
    color: #374151;
  }
</style>
