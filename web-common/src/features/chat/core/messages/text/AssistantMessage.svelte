<!-- Renders assistant responses from router_agent. -->
<script lang="ts">
  import { getMetricsResolverQueryToUrlMapperStore } from "@rilldata/web-common/features/chat/core/messages/text/citation-url-mapper.ts";
  import { enhanceCitationLinks } from "@rilldata/web-common/features/chat/core/messages/text/enhance-citation-links.ts";
  import { rewriteCitationUrls } from "@rilldata/web-common/features/chat/core/messages/text/rewrite-citation-urls.ts";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { Conversation } from "../../conversation";
  import FeedbackButtons from "../../feedback/FeedbackButtons.svelte";
  import { extractMessageText } from "../../utils";
  import type { TextBlock } from "./text-block";

  export let block: TextBlock;
  export let conversation: Conversation;
  export let onDownvote: (messageId: string) => void;

  $: message = block.message;
  $: messageId = message.id ?? "";

  const mapperStore = getMetricsResolverQueryToUrlMapperStore(conversation);

  // Safety net: strip wrapper if the LLM wraps the entire response in ```markdown fences
  $: rawContent = extractMessageText(message).replace(
    /^```markdown\n([\s\S]*)\n```$/,
    "$1",
  );
  $: contentPromise = $mapperStore.data
    ? rewriteCitationUrls(rawContent, $mapperStore.data)
    : Promise.resolve(rawContent);
</script>

<div class="chat-message">
  <div class="chat-message-content" use:enhanceCitationLinks={conversation}>
    {#await contentPromise}
      <Markdown content={rawContent} />
    {:then content}
      <Markdown {content} />
    {/await}
  </div>
  <div class="chat-message-actions">
    <FeedbackButtons
      {messageId}
      {conversation}
      feedback={block.feedback}
      {onDownvote}
    />
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

  .chat-message-actions {
    @apply pb-2;
  }
</style>
