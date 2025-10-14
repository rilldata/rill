<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import TextMessage from "./TextMessage.svelte";
  import ToolCallBlock from "./ToolCallBlock.svelte";

  export let message: V1Message;

  // State for collapsible content blocks
  let expandedBlocks: { [key: number]: boolean } = {};

  function toggleBlock(index: number) {
    expandedBlocks[index] = !expandedBlocks[index];
  }

  // TODO: This correlation logic should be moved to the Conversation class business layer.
  // Currently we have split responsibility: Conversation handles streaming correlation,
  // but UI handles initial fetch correlation. This creates inconsistent architecture.
  // All correlation should happen in the business layer, with UI just displaying pre-correlated data.

  // Group tool calls with their results within this message
  function groupToolCallsWithResults(content: any[]) {
    const groups: any[] = [];
    const toolResults = new Map();

    // First pass: collect all tool results by ID within this message
    content.forEach((block) => {
      if (block.toolResult && block.toolResult.id) {
        toolResults.set(block.toolResult.id, block.toolResult);
      }
    });

    // Second pass: process blocks and create groups
    content.forEach((block, index) => {
      if (block.text) {
        groups.push({ type: "text", content: block.text, index });
      } else if (block.toolCall) {
        // Streaming merges tool results into toolCall blocks.
        // For initial fetch (GetConversation), calls/results are separate, so we attach via fallback.
        groups.push({
          type: "tool",
          toolCall: block.toolCall,
          toolResult: block.toolResult || toolResults.get(block.toolCall.id),
          index,
        });
      }
      // Skip standalone toolResult blocks as they should be merged with toolCalls
    });

    return groups;
  }

  $: groupedContent = message.content
    ? groupToolCallsWithResults(message.content)
    : [];
</script>

<div class="complex-message-container">
  {#each groupedContent as group, i (i)}
    {#if group.type === "text"}
      <!-- Text block -->
      <TextMessage {message} content={group.content} />
    {:else if group.type === "tool"}
      <!-- Tool Call + Result block -->
      <ToolCallBlock
        toolCall={group.toolCall}
        toolResult={group.toolResult}
        isExpanded={expandedBlocks[i] || false}
        onToggle={() => toggleBlock(i)}
      />
    {/if}
  {/each}
</div>

<style lang="postcss">
  .complex-message-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 90%;
    align-self: flex-start;
    width: 100%;
  }
</style>
