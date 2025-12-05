<!--
  Renders a collapsible "thinking" block containing a progress message and tool calls.
  Messages are displayed in the exact order they were received.
-->
<script lang="ts">
  import Brain from "../../../../../components/icons/Brain.svelte";
  import CaretDownIcon from "../../../../../components/icons/CaretDownIcon.svelte";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { V1Message, V1Tool } from "../../../../../runtime-client";
  import { MessageType } from "../../types";
  import CallMessage from "./CallMessage.svelte";
  import ShimmerText from "./ShimmerText.svelte";

  export let messages: V1Message[];
  export let resultMessagesByParentId: Map<string | undefined, V1Message>;
  export let isComplete: boolean;
  export let duration: number;
  export let tools: V1Tool[] | undefined = undefined;

  let isExpanded = true;
  let hasUserInteracted = false;

  $: headerText = isComplete
    ? `Thought for ${formatDuration(duration)}`
    : "Thinking...";

  // Auto-collapse when thinking completes, unless user has interacted
  $: if (isComplete && !hasUserInteracted) {
    isExpanded = false;
  }

  function formatDuration(seconds: number): string {
    if (seconds < 1) return "less than a second";
    if (seconds === 1) return "1 second";
    if (seconds < 60) return `${seconds} seconds`;

    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;

    if (remainingSeconds === 0) {
      return minutes === 1 ? "1 minute" : `${minutes} minutes`;
    }

    return `${minutes} ${minutes === 1 ? "minute" : "minutes"} ${remainingSeconds} ${remainingSeconds === 1 ? "second" : "seconds"}`;
  }

  function toggleExpanded() {
    hasUserInteracted = true;
    isExpanded = !isExpanded;
  }
</script>

<div class="thinking-block">
  <button class="thinking-header" on:click={toggleExpanded}>
    <div class="thinking-icon">
      {#if isExpanded}
        <CaretDownIcon size="14" color="currentColor" />
      {:else}
        <Brain />
      {/if}
    </div>
    <div class="thinking-title">
      {#if !isComplete}
        <ShimmerText>{headerText}</ShimmerText>
      {:else}
        {headerText}
      {/if}
    </div>
  </button>

  {#if isExpanded}
    <div class="thinking-messages">
      {#each messages as msg (msg.id)}
        {#if msg.type === MessageType.PROGRESS}
          {#if msg.contentData}
            <div class="thinking-content">
              <Markdown content={msg.contentData} />
            </div>
          {/if}
        {:else if msg.type === MessageType.CALL}
          <CallMessage
            message={msg}
            resultMessage={resultMessagesByParentId.get(msg.id)}
            {tools}
          />
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  .thinking-block {
    @apply w-full max-w-full self-start;
  }

  .thinking-header {
    @apply w-full flex items-center gap-1.5 px-1 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-gray-500 transition-colors;
  }

  .thinking-header:hover {
    @apply text-gray-600;
  }

  .thinking-icon {
    @apply flex items-center;
  }

  .thinking-title {
    @apply flex-1 text-left font-normal;
  }

  .thinking-messages {
    @apply pl-5 flex flex-col gap-0;
    @apply max-h-96 overflow-y-auto;
  }

  .thinking-content {
    @apply px-1 py-1 text-xs leading-relaxed break-words;
  }

  .thinking-content :global(*) {
    @apply text-gray-500;
  }

  .thinking-content :global(strong),
  .thinking-content :global(b) {
    @apply text-gray-600 font-semibold;
  }

  .thinking-content :global(a) {
    @apply text-gray-600 underline;
  }
</style>
