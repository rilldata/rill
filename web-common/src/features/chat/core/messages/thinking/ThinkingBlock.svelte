<!--
  Renders a collapsible "thinking" block containing a progress message and tool calls.
  Messages are displayed in the exact order they were received.
-->
<script lang="ts">
  import { builderActions, getAttrs } from "bits-ui";
  import * as Collapsible from "../../../../../components/collapsible";
  import Brain from "../../../../../components/icons/Brain.svelte";
  import CaretDownIcon from "../../../../../components/icons/CaretDownIcon.svelte";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { V1Tool } from "../../../../../runtime-client";
  import { MessageType } from "../../types";
  import AnimatedDots from "../AnimatedDots.svelte";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { ThinkingBlock } from "./thinking-block";

  export let block: ThinkingBlock;
  export let tools: V1Tool[] | undefined = undefined;

  let isExpanded = true;
  let hasUserInteracted = false;

  // Track completion state to detect transition
  let wasComplete = false;
  $: {
    // Auto-collapse only on the transition to complete, not continuously
    if (block.isComplete && !wasComplete && !hasUserInteracted) {
      isExpanded = false;
    }
    wasComplete = block.isComplete;
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

  function onUserInteraction() {
    hasUserInteracted = true;
  }
</script>

<Collapsible.Root bind:open={isExpanded} class="w-full max-w-full self-start">
  <Collapsible.Trigger asChild let:builder>
    <button
      class="thinking-header"
      {...getAttrs([builder])}
      use:builderActions={{ builders: [builder] }}
      on:click={onUserInteraction}
    >
      <div class="thinking-icon">
        {#if isExpanded}
          <CaretDownIcon size="14" color="currentColor" />
        {:else}
          <Brain />
        {/if}
      </div>
      <div class="thinking-title">
        {#if block.isComplete}
          Thought for {formatDuration(block.duration)}
        {:else}
          <AnimatedDots>Thinking</AnimatedDots>
        {/if}
      </div>
    </button>
  </Collapsible.Trigger>

  <Collapsible.Content transition={undefined} class="pl-5 flex flex-col gap-0">
    {#each block.messages as msg (msg.id)}
      {#if msg.type === MessageType.PROGRESS}
        {#if msg.contentData}
          <div class="thinking-content">
            <Markdown content={msg.contentData} />
          </div>
        {/if}
      {:else if msg.type === MessageType.CALL}
        <ToolCall
          message={msg}
          resultMessage={block.resultMessagesByParentId.get(msg.id)}
          {tools}
          variant="inline"
        />
      {/if}
    {/each}
  </Collapsible.Content>
</Collapsible.Root>

<style lang="postcss">
  .thinking-header {
    @apply w-full flex items-center gap-1.5 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-fg-secondary transition-colors;
  }

  .thinking-header:hover {
    @apply text-fg-secondary;
  }

  .thinking-icon {
    @apply flex items-center;
  }

  .thinking-title {
    @apply flex-1 text-left font-normal;
  }

  .thinking-content {
    @apply py-1 text-xs leading-relaxed break-words;
  }

  .thinking-content :global(*) {
    @apply text-fg-secondary;
  }

  .thinking-content :global(strong),
  .thinking-content :global(b) {
    @apply text-fg-secondary font-semibold;
  }

  .thinking-content :global(a) {
    @apply text-fg-secondary underline;
  }
</style>
