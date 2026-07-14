<!--
  Renders a collapsible "thinking" block containing a progress message and tool calls.
  Messages are displayed in the exact order they were received.
-->
<script lang="ts">
  import * as Collapsible from "../../../../../components/collapsible";
  import Brain from "../../../../../components/icons/Brain.svelte";
  import CaretDownIcon from "../../../../../components/icons/CaretDownIcon.svelte";
  import Markdown from "../../../../../components/markdown/Markdown.svelte";
  import type { V1Tool } from "../../../../../runtime-client";
  import { MessageType } from "../../types";
  import AnimatedDots from "../AnimatedDots.svelte";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { ThinkingBlock } from "./thinking-block";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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
    if (seconds < 1) return m.chat_duration_less_than_second();
    if (seconds < 60) return m.chat_duration_seconds({ count: seconds });
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    if (remainingSeconds === 0)
      return m.chat_duration_minutes({ count: minutes });
    return `${m.chat_duration_minutes({ count: minutes })} ${m.chat_duration_seconds({ count: remainingSeconds })}`;
  }

  function onUserInteraction() {
    hasUserInteracted = true;
  }
</script>

<Collapsible.Root bind:open={isExpanded} class="w-full max-w-full self-start">
  <Collapsible.Trigger>
    {#snippet child({ props })}
      <button {...props} class="thinking-header" onclick={onUserInteraction}>
        <div class="thinking-icon">
          {#if isExpanded}
            <CaretDownIcon size="14" />
          {:else}
            <Brain />
          {/if}
        </div>
        <div class="thinking-title">
          {#if block.isComplete}
            {m.chat_thought_for({ duration: formatDuration(block.duration) })}
          {:else}
            <AnimatedDots>{m.chat_thinking()}</AnimatedDots>
          {/if}
        </div>
      </button>
    {/snippet}
  </Collapsible.Trigger>

  <Collapsible.Content class="pl-5 flex flex-col gap-0">
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
