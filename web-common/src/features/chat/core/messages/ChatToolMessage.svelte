<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import ChatTextMessage from "./ChatTextMessage.svelte";
  import ChatToolCallBlock from "./ChatToolCallBlock.svelte";

  export let message: V1Message;

  // State for collapsible content blocks
  let expandedBlocks: { [key: number]: boolean } = {};

  function toggleBlock(index: number) {
    expandedBlocks[index] = !expandedBlocks[index];
  }

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
        // Tool result should already be merged into this block by the data layer
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
      <ChatTextMessage {message} content={group.content} />
    {:else if group.type === "tool"}
      <!-- Tool Call + Result block -->
      <ChatToolCallBlock
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
  }
</style>
