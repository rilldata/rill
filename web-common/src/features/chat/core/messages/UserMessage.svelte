<!-- Renders user prompt messages. -->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";
  import DOMPurify from "dompurify";
  import { convertPromptWithInlineContextToHTML } from "@rilldata/web-common/features/chat/core/context/inline-context-convertors.ts";
  import { getInlineChatContextMetadata } from "@rilldata/web-common/features/chat/core/context/inline-context-data.ts";

  export let message: V1Message;

  // Message content
  $: content = extractMessageText(message);

  const contextMetadataStore = getInlineChatContextMetadata();
</script>

<div class="chat-message">
  <div class="chat-message-content">
    {@html DOMPurify.sanitize(
      convertPromptWithInlineContextToHTML(content, $contextMetadataStore),
    )}
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
