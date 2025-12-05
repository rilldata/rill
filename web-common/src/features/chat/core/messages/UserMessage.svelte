<!-- Renders user prompt messages. -->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";
  import { convertPromptWithInlineContextToComponents } from "@rilldata/web-common/features/chat/core/context/inline-context-convertors.ts";
  import InlineContext from "@rilldata/web-common/features/chat/core/context/InlineContext.svelte";

  export let message: V1Message;

  // Message content
  $: content = extractMessageText(message);
  $: linesOfTextOrComponents =
    convertPromptWithInlineContextToComponents(content);
</script>

<div class="chat-message">
  <div class="chat-message-content">
    {#each linesOfTextOrComponents as textOrComponents, i (i)}
      {#each textOrComponents as textOrComponent, i (i)}
        {#if textOrComponent.type === "text"}
          <span>{textOrComponent.text}</span>
        {:else if textOrComponent.type === "context"}
          <InlineContext
            selectedChatContext={textOrComponent.context}
            props={{ mode: "readonly" }}
          />
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
