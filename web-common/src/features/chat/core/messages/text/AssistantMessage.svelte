<!-- Renders assistant responses from router_agent. -->
<script lang="ts">
  import { enhanceCitationLinks } from "@rilldata/web-common/features/chat/core/messages/text/enhance-citation-links.ts";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../../runtime-client";
  import { extractMessageText } from "../../utils";

  export let message: V1Message;

  // Message content and styling
  $: content = extractMessageText(message);
</script>

<div class="chat-message">
  <div class="chat-message-content" use:enhanceCitationLinks>
    <Markdown {content} />
  </div>
</div>

<style lang="postcss">
  .chat-message {
    @apply max-w-full;
  }

  .chat-message-content {
    @apply py-2;
    @apply text-sm leading-relaxed break-words;
    @apply text-fg-primary;
  }
</style>
