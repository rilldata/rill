<!-- Renders user prompt messages. -->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";
  import { convertPromptWithInlineContextToComponents } from "@rilldata/web-common/features/chat/core/context/inline-context-convertors.ts";

  export let message: V1Message;

  // Message content
  $: content = extractMessageText(message);
  $: linesOfTextOrComponents =
    convertPromptWithInlineContextToComponents(content);
</script>

<div class="chat-message">
  <div class="chat-message-content">
    {#each linesOfTextOrComponents as textOrComponents, i (i)}
      {#each textOrComponents as { isSvelteComponent, text, component, props }, i (i)}
        {#if isSvelteComponent}
          <svelte:component this={component} {...props} />
        {:else}
          <span>{text}</span>
        {/if}
      {/each}
      {#if i < linesOfTextOrComponents.length - 1}
        <br />
      {/if}
    {/each}
  </div>
</div>

<style lang="postcss">
  .chat-message {
    @apply max-w-[90%] self-end;
  }

  .chat-message-content {
    @apply px-4 py-2 rounded-2xl;
    @apply text-sm leading-relaxed break-words;
    @apply bg-muted text-foreground rounded-br-lg;
  }
</style>
