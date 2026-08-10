<!-- Renders assistant responses from router_agent. -->
<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { enhanceCitationLinks } from "@rilldata/web-common/features/chat/core/messages/text/enhance-citation-links.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ClipboardCheck } from "lucide-svelte";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { Conversation } from "../../conversation";
  import FeedbackButtons from "../../feedback/FeedbackButtons.svelte";
  import { extractMessageText } from "../../utils";
  import type { TextBlock } from "./text-block";

  export let block: TextBlock;
  export let conversation: Conversation;
  export let onDownvote: (messageId: string) => void;
  export let onAddToEval: ((messageId: string) => void) | undefined = undefined;

  $: message = block.message;
  $: messageId = message.id ?? "";

  // Safety net: strip wrapper if the LLM wraps the entire response in ```markdown fences
  $: content = extractMessageText(message).replace(
    /^```markdown\n([\s\S]*)\n```$/,
    "$1",
  );
</script>

<div class="chat-message">
  <div class="chat-message-content" use:enhanceCitationLinks={conversation}>
    <Markdown {content} />
  </div>
  <div class="chat-message-actions">
    <FeedbackButtons
      {messageId}
      {conversation}
      feedback={block.feedback}
      {onDownvote}
    />
    {#if onAddToEval}
      <IconButton
        ariaLabel={m.add_to_eval()}
        onclick={() => onAddToEval?.(messageId)}
      >
        <ClipboardCheck size="14px" class="text-fg-muted" />
        <svelte:fragment slot="tooltip-content">
          {m.add_to_eval()}
        </svelte:fragment>
      </IconButton>
    {/if}
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
    @apply pb-2 flex items-center gap-1;
  }
</style>
