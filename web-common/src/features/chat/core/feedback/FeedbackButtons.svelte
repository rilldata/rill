<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import { ThumbsDown, ThumbsUp } from "lucide-svelte";
  import { slide } from "svelte/transition";
  import type { Conversation } from "../conversation";
  import type { FeedbackData } from "../messages/text/text-block";

  export let messageId: string;
  export let conversation: Conversation;
  export let feedback: FeedbackData | undefined;
  export let onDownvote: (messageId: string) => void;

  // Derive state from props
  $: hasPositiveFeedback = feedback?.sentiment === "positive";
  $: hasNegativeFeedback = feedback?.sentiment === "negative";
  $: feedbackResponse = feedback?.response ?? null;
  $: isPending = feedback?.isPending ?? false;
  $: isStreaming = conversation.isStreaming;
  $: isDisabled = $isStreaming;

  function handleUpvote() {
    if (isDisabled) return;
    void conversation.submitFeedback(messageId, "positive");
  }

  function handleDownvote() {
    if (isDisabled) return;
    onDownvote(messageId);
  }
</script>

<div class="feedback-container">
  <div class="feedback-buttons">
    <IconButton
      size={24}
      disabled={isDisabled}
      ariaPressed={hasPositiveFeedback}
      ariaLabel="Upvote response"
      on:click={handleUpvote}
    >
      <ThumbsUp
        size={14}
        class={hasPositiveFeedback ? "text-primary-500" : "text-gray-400"}
      />
      <svelte:fragment slot="tooltip-content"
        >This response was helpful</svelte:fragment
      >
    </IconButton>
    <IconButton
      size={24}
      disabled={isDisabled}
      ariaPressed={hasNegativeFeedback}
      ariaLabel="Downvote response"
      on:click={handleDownvote}
    >
      <ThumbsDown
        size={14}
        class={hasNegativeFeedback ? "text-primary-500" : "text-gray-400"}
      />
      <svelte:fragment slot="tooltip-content"
        >This response needs improvement</svelte:fragment
      >
    </IconButton>
  </div>

  {#if isPending}
    <div class="feedback-loading" in:slide={{ duration: 200 }}>
      Analyzing feedback...
    </div>
  {:else if feedbackResponse}
    <p class="feedback-response" in:slide={{ duration: 200 }}>
      {feedbackResponse}
    </p>
  {/if}
</div>

<style lang="postcss">
  .feedback-container {
    @apply flex flex-col gap-1;
  }

  .feedback-buttons {
    @apply flex items-center gap-0.5;
  }

  .feedback-loading {
    @apply text-xs text-gray-400 italic;
    @apply pl-1 py-1;
  }

  .feedback-response {
    @apply text-xs text-gray-500 italic;
    @apply pl-1 py-1;
  }
</style>
